package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestLog(t *testing.T, directory, name, content string) string {
	t.Helper()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestNativeReportRequiresCompletionAndCountsExpectedErrorsCorrectly(t *testing.T) {
	directory := t.TempDir()
	report := writeTestLog(t, directory, "report", "0: test262/test/negative.js  @negative\nSyntaxError: expected\n1: test262/test/module.js  module\n  FAILED\n2: test262/test/skipped.js  SKIPPED\n")
	stderr := writeTestLog(t, directory, "stderr", "Result: 1/3 errors, 2 excluded, 1 skipped, 1 new\n")
	result, err := readNative(report, stderr)
	if err != nil {
		t.Fatal(err)
	}
	if result.Variants["PASS"] != 2 || result.Variants["FAIL"] != 1 || result.Files["negative.js"].Status != "PASS" || result.Files["module.js"].Status != "FAIL" || result.Files["skipped.js"].Status != "SKIP" {
		t.Fatalf("incorrect native results: %+v", result)
	}
	for _, content := range []string{"progress without completion", "Result: 0/3 errors, 1 skipped\n", "Result: 1/4 errors, 1 skipped\n", "Result: 1/3 errors\n"} {
		writeTestLog(t, directory, "stderr", content)
		if _, err := readNative(report, stderr); err == nil {
			t.Fatalf("accepted incomplete or inconsistent summary: %s", content)
		}
	}
}

func TestGoMLReportPreservesMixedVariantsAndRejectsDuplicateResults(t *testing.T) {
	directory := t.TempDir()
	suite := filepath.Join(directory, "suite")
	file := filepath.Join(suite, "test", "case.js")
	log := writeTestLog(t, directory, "goml.log", file+" [sloppy] FAIL runtime error\n"+file+" [strict] PASS\n")
	result, err := readGoML([]string{log}, suite)
	if err != nil {
		t.Fatal(err)
	}
	if result.Files["case.js"].Status != "FAIL" || result.Variants["PASS"] != 1 || result.Variants["FAIL"] != 1 {
		t.Fatalf("lost variant failure: %+v", result)
	}
	if _, err := readGoML([]string{log, log}, suite); err == nil {
		t.Fatal("accepted duplicate results")
	}
	writeTestLog(t, directory, "goml.log", "panic: runner crashed\n")
	if _, err := readGoML([]string{log}, suite); err == nil {
		t.Fatal("accepted runner crash")
	}
}

func TestComparisonSeparatesRegressionsCoverageAndNativeFailures(t *testing.T) {
	native := suiteResult{Files: map[string]testFile{
		"regression.js":  {Status: "PASS"},
		"coverage.js":    {Status: "PASS"},
		"improvement.js": {Status: "FAIL"},
		"shared-skip.js": {Status: "SKIP"},
	}}
	goml := suiteResult{Files: map[string]testFile{
		"regression.js":  {Status: "FAIL"},
		"coverage.js":    {Status: "SKIP", Reasons: []string{"host realms"}},
		"improvement.js": {Status: "PASS"},
		"shared-skip.js": {Status: "SKIP"},
		"extra.js":       {Status: "PASS"},
	}}
	result := compare(native, goml, []string{"regression.js", "coverage.js", "improvement.js", "shared-skip.js", "extra.js", "missing.js"})
	if strings.Join(result.Regressions, ",") != "regression.js" || strings.Join(result.CoverageGaps, ",") != "coverage.js" || strings.Join(result.NativeFailures, ",") != "improvement.js" || result.CoverageReasons["host realms"] != 1 || result.Matrix["EXCLUDED/MISSING"] != 1 || result.Matrix["EXCLUDED/PASS"] != 1 {
		t.Fatalf("incorrect comparison: %+v", result)
	}
}
