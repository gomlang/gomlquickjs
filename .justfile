set shell := ["bash", "-euo", "pipefail", "-c"]

goml := env("GOML", justfile_directory() + "/_artifact/toolchain/bin/goml")
build_jobs := env("GOML_BUILD_JOBS", "2")
test_jobs := env("GOML_TEST_JOBS", "2")

default: build

toolchain:
    bash toolchain/install.sh

fmt:
    "{{goml}}" fmt

fmt-check:
    "{{goml}}" fmt --check

check:
    "{{goml}}" check

test:
    "{{goml}}" test --jobs "{{test_jobs}}" --timeout 10m

build:
    "{{goml}}" build --jobs "{{build_jobs}}"

smoke: build
    bash tools/smoke.sh

test-tools:
    go test tools/compare_test262.go tools/compare_test262_test.go

test262-compare quickjs:
    systemd-run --user --scope --quiet --slice=gomlquickjs.slice -p MemoryMax=1536M -p MemorySwapMax=1G -p CPUQuota=150% -p TasksMax=128 env GOMAXPROCS=2 GOMEMLIMIT=512MiB "{{goml}}" build --jobs 1
    systemd-run --user --scope --quiet --slice=gomlquickjs.slice -p MemoryMax=1536M -p MemorySwapMax=1G -p CPUQuota=150% -p TasksMax=128 env GOMAXPROCS=2 GOMEMLIMIT=512MiB go run tools/compare_test262.go --quickjs "{{quickjs}}"

ci: fmt-check test-tools test smoke

clean:
    rm -rf _artifact
