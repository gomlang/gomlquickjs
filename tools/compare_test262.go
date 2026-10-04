package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type testFile struct {
	Status  string            `json:"status"`
	Results map[string]string `json:"results,omitempty"`
	Reasons []string          `json:"reasons,omitempty"`
}

type suiteResult struct {
	Variants map[string]int      `json:"variants"`
	Files    map[string]testFile `json:"files"`
}

type comparison struct {
	CreatedAt       string            `json:"created_at"`
	Revisions       map[string]string `json:"revisions"`
	Files           int               `json:"files"`
	Native          suiteResult       `json:"quickjs"`
	GoML            suiteResult       `json:"gomlquickjs"`
	Matrix          map[string]int    `json:"file_status_matrix"`
	Regressions     []string          `json:"regressions"`
	CoverageGaps    []string          `json:"coverage_gaps"`
	CoverageReasons map[string]int    `json:"coverage_gap_reasons"`
	NativeFailures  []string          `json:"native_failures_passed_by_goml"`
}

var nativeHeader = regexp.MustCompile(`^\d+: (\S+)(.*)$`)
var nativeSummary = regexp.MustCompile(`Result: (\d+)/(\d+) errors?(?:, (\d+) excluded)?(?:, (\d+) skipped)?`)
var gomlLine = regexp.MustCompile(`^(.*?) \[(sloppy|strict|module|raw)\] (PASS|FAIL|SKIP)(?: (.*))?$`)

func readLines(path string, visit func(string) error) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 4<<20)
	for scanner.Scan() {
		if err := visit(scanner.Text()); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func readNative(report, stderr string) (suiteResult, error) {
	result := suiteResult{Variants: map[string]int{"PASS": 0, "FAIL": 0}, Files: map[string]testFile{}}
	current := ""
	executed, failures, skipped := 0, 0, 0
	err := readLines(report, func(line string) error {
		if match := nativeHeader.FindStringSubmatch(line); match != nil {
			current = strings.TrimPrefix(filepath.ToSlash(match[1]), "test262/test/")
			if _, exists := result.Files[current]; exists {
				return fmt.Errorf("duplicate native result: %s", current)
			}
			entry := testFile{Status: "PASS"}
			if strings.Contains(match[2], "SKIPPED") {
				entry.Status = "SKIP"
				skipped++
			} else if strings.Contains(match[2], "@noStrict") || strings.Contains(match[2], "@onlyStrict") || strings.Contains(match[2], "module") {
				executed++
			} else {
				executed += 2
			}
			result.Files[current] = entry
		} else if line == "  FAILED" {
			entry, exists := result.Files[current]
			if !exists || entry.Status == "SKIP" {
				return fmt.Errorf("native failure without an executed test")
			}
			entry.Status = "FAIL"
			result.Files[current] = entry
			failures++
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	data, err := os.ReadFile(stderr)
	if err != nil {
		return result, err
	}
	match := nativeSummary.FindStringSubmatch(string(data))
	if match == nil {
		return result, fmt.Errorf("native runner did not finish; see %s", stderr)
	}
	failed, _ := strconv.Atoi(match[1])
	total, _ := strconv.Atoi(match[2])
	skippedTotal, _ := strconv.Atoi(match[4])
	if executed != total || failures != failed || skipped != skippedTotal {
		return result, fmt.Errorf("native report and summary disagree: report %d/%d failures, %d skipped; summary %d/%d, %d skipped", failures, executed, skipped, failed, total, skippedTotal)
	}
	result.Variants["PASS"] = executed - failures
	result.Variants["FAIL"] = failures
	return result, nil
}

func readGoML(logs []string, suite string) (suiteResult, error) {
	result := suiteResult{Variants: map[string]int{"PASS": 0, "FAIL": 0, "SKIP": 0}, Files: map[string]testFile{}}
	for _, log := range logs {
		err := readLines(log, func(line string) error {
			match := gomlLine.FindStringSubmatch(line)
			if match == nil {
				return fmt.Errorf("unrecognized runner output in %s: %s", log, line)
			}
			name, err := filepath.Rel(filepath.Join(suite, "test"), match[1])
			if err != nil || strings.HasPrefix(name, "..") {
				return fmt.Errorf("test outside suite: %s", match[1])
			}
			name = filepath.ToSlash(name)
			entry := result.Files[name]
			if entry.Results == nil {
				entry.Results = map[string]string{}
			}
			if _, exists := entry.Results[match[2]]; exists {
				return fmt.Errorf("duplicate GoML result: %s [%s]", name, match[2])
			}
			entry.Results[match[2]] = match[3]
			if entry.Status == "" || match[3] == "FAIL" || entry.Status == "SKIP" && match[3] == "PASS" {
				entry.Status = match[3]
			}
			if match[4] != "" && !slices.Contains(entry.Reasons, match[4]) {
				entry.Reasons = append(entry.Reasons, match[4])
			}
			result.Files[name] = entry
			result.Variants[match[3]]++
			return nil
		})
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func compare(native, goml suiteResult, files []string) comparison {
	result := comparison{CreatedAt: time.Now().UTC().Format(time.RFC3339), Files: len(files), Native: native, GoML: goml, Matrix: map[string]int{}, CoverageReasons: map[string]int{}, Regressions: []string{}, CoverageGaps: []string{}, NativeFailures: []string{}}
	for _, file := range files {
		n, g := native.Files[file], goml.Files[file]
		if n.Status == "" {
			n.Status = "EXCLUDED"
		}
		if g.Status == "" {
			g.Status = "MISSING"
		}
		result.Matrix[n.Status+"/"+g.Status]++
		if n.Status == "PASS" && g.Status == "FAIL" {
			result.Regressions = append(result.Regressions, file)
		}
		if n.Status == "PASS" && g.Status == "SKIP" {
			result.CoverageGaps = append(result.CoverageGaps, file)
			for _, reason := range g.Reasons {
				result.CoverageReasons[reason]++
			}
		}
		if n.Status == "FAIL" && g.Status == "PASS" {
			result.NativeFailures = append(result.NativeFailures, file)
		}
	}
	return result
}

func runCommand(directory, stdout, stderr string, timeout time.Duration, environment []string, command ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	out, err := os.Create(stdout)
	if err != nil {
		return err
	}
	defer out.Close()
	errout := out
	if stderr != stdout {
		errout, err = os.Create(stderr)
		if err != nil {
			return err
		}
		defer errout.Close()
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir, cmd.Stdout, cmd.Stderr = directory, out, errout
	cmd.Env = append(os.Environ(), environment...)
	err = cmd.Run()
	if ctx.Err() != nil {
		return fmt.Errorf("%s timed out after %s", command[0], timeout)
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return nil
	}
	return err
}

func revision(directory string) (string, error) {
	data, err := exec.Command("git", "-C", directory, "rev-parse", "HEAD").Output()
	return strings.TrimSpace(string(data)), err
}

func execute() error {
	quickjs := flag.String("quickjs", "", "C QuickJS checkout matching UPSTREAM.toml, with run-test262 and patched test262 installed")
	output := flag.String("output", "_artifact/test262-compare-"+time.Now().Format("20060102-150405"), "new output directory for logs and summary.json")
	runner := flag.String("runner", "_artifact/bin/cmd/run_test262/run_test262", "GoML Test262 executable")
	jobs := flag.Int("jobs", 1, "parallel GoML runner processes")
	batchSize := flag.Int("batch-size", 128, "test files per GoML process")
	timeout := flag.Duration("timeout", 10*time.Minute, "timeout per GoML batch")
	requireCoverage := flag.Bool("require-coverage", false, "fail when GoML skips a test that C QuickJS passes")
	flag.Parse()
	if *quickjs == "" || *jobs < 1 || *batchSize < 1 || *timeout <= 0 || flag.NArg() != 0 {
		return fmt.Errorf("provide --quickjs DIR and positive jobs, batch-size, and timeout")
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	for _, path := range []*string{quickjs, output, runner} {
		*path, err = filepath.Abs(*path)
		if err != nil {
			return err
		}
	}
	suite := filepath.Join(*quickjs, "test262")
	revisions := map[string]string{}
	for name, directory := range map[string]string{"gomlquickjs": root, "quickjs": *quickjs, "test262": suite} {
		revisions[name], err = revision(directory)
		if err != nil {
			return err
		}
	}
	baseline, err := os.ReadFile("UPSTREAM.toml")
	if err != nil {
		return err
	}
	if !strings.Contains(string(baseline), `commit = "`+revisions["quickjs"]+`"`) {
		return fmt.Errorf("C QuickJS checkout does not match UPSTREAM.toml")
	}
	makefile, err := os.ReadFile(filepath.Join(*quickjs, "Makefile"))
	if err != nil {
		return err
	}
	if !strings.Contains(string(makefile), "TEST262_COMMIT?="+revisions["test262"]) {
		return fmt.Errorf("Test262 checkout does not match C QuickJS's pinned revision")
	}
	if _, err := os.Stat(*output); !os.IsNotExist(err) {
		return fmt.Errorf("output directory must not exist: %s", *output)
	}
	if err := os.MkdirAll(*output, 0755); err != nil {
		return err
	}
	files, names := []string{}, []string{}
	err = filepath.WalkDir(filepath.Join(suite, "test"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".js") {
			name, err := filepath.Rel(filepath.Join(suite, "test"), path)
			if err != nil {
				return err
			}
			files, names = append(files, path), append(names, filepath.ToSlash(name))
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no Test262 files found")
	}
	if err := os.WriteFile(filepath.Join(*output, "files.txt"), []byte(strings.Join(files, "\n")+"\n"), 0644); err != nil {
		return err
	}
	for name, directory := range map[string]string{"gomlquickjs": root, "quickjs": *quickjs, "test262": suite} {
		patch, err := exec.Command("git", "-C", directory, "diff", "HEAD", "--").Output()
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(*output, name+".patch"), patch, 0644); err != nil {
			return err
		}
	}
	report, stderr := filepath.Join(*output, "c-report.txt"), filepath.Join(*output, "c-stderr.log")
	fmt.Println("Running pinned C QuickJS with test262.conf and both execution modes")
	err = runCommand(*quickjs, filepath.Join(*output, "c-stdout.log"), stderr, 15*time.Minute, nil, filepath.Join(*quickjs, "run-test262"), "-c", "test262.conf", "-a", "-T", "1", "-e", "/dev/null", "-r", report)
	if err != nil {
		return err
	}
	native, err := readNative(report, stderr)
	if err != nil {
		return err
	}
	fmt.Printf("C QuickJS: %d passed, %d failed variants\n", native.Variants["PASS"], native.Variants["FAIL"])
	logs := []string{}
	batches := make(chan int)
	var workers sync.WaitGroup
	var lock sync.Mutex
	var batchError error
	completed, total := 0, (len(files)+*batchSize-1) / *batchSize
	for index := range total {
		logs = append(logs, filepath.Join(*output, fmt.Sprintf("goml-%03d.log", index)))
	}
	for range *jobs {
		workers.Go(func() {
			for index := range batches {
				start := index * *batchSize
				args := append([]string{*runner, "--harness", filepath.Join(suite, "harness")}, files[start:min(start+*batchSize, len(files))]...)
				err := runCommand(root, logs[index], logs[index], *timeout, []string{"TZ=America/Los_Angeles", "GOMAXPROCS=1", "GOMEMLIMIT=512MiB"}, args...)
				lock.Lock()
				if err != nil {
					batchError = errors.Join(batchError, fmt.Errorf("batch %d: %w", index, err))
				}
				completed++
				if completed%10 == 0 || completed == total {
					fmt.Printf("GoML batches: %d/%d\n", completed, total)
				}
				lock.Unlock()
			}
		})
	}
	for index := range total {
		batches <- index
	}
	close(batches)
	workers.Wait()
	if batchError != nil {
		return batchError
	}
	goml, err := readGoML(logs, suite)
	if err != nil {
		return err
	}
	if len(goml.Files) != len(files) {
		return fmt.Errorf("incomplete GoML results: %d/%d files", len(goml.Files), len(files))
	}
	result := compare(native, goml, names)
	result.Revisions = revisions
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(*output, "summary.json"), append(data, '\n'), 0644); err != nil {
		return err
	}
	fmt.Printf("GoML: %d passed, %d failed, %d skipped variants\n", goml.Variants["PASS"], goml.Variants["FAIL"], goml.Variants["SKIP"])
	fmt.Printf("C PASS / GoML FAIL: %d files; C PASS / GoML SKIP: %d files\n", len(result.Regressions), len(result.CoverageGaps))
	fmt.Printf("Results: %s\n", filepath.Join(*output, "summary.json"))
	if goml.Variants["FAIL"] > 0 || *requireCoverage && len(result.CoverageGaps) > 0 {
		return fmt.Errorf("comparison requirements were not met")
	}
	return nil
}

func main() {
	if err := execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
