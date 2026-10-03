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

ci: fmt-check test smoke

clean:
    rm -rf _artifact
