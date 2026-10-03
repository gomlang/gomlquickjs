#!/usr/bin/env bash
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repository_root"
binary="$repository_root/_artifact/bin/cmd/qjs/qjs"
runner="$repository_root/_artifact/bin/cmd/run_test262/run_test262"
output="$repository_root/_artifact/smoke"
mkdir -p "$output"

"$binary" --help > "$output/qjs-help.txt"
"$runner" --help > "$output/test262-help.txt"
"$binary" -e 'print(1 + 2); print(JSON.stringify({answer: 42})); Promise.resolve(6).then(x => print(x * 7));' > "$output/eval.actual"
printf '3\n{"answer":42}\n42\n' > "$output/eval.expected"
diff -u "$output/eval.expected" "$output/eval.actual"

printf 'export const answer = 42;\n' > "$output/dependency.mjs"
printf 'import {answer} from "./dependency.mjs"; print(await Promise.resolve(answer));\n' > "$output/main.mjs"
"$binary" "$output/main.mjs" > "$output/module.actual"
printf '42\n' > "$output/module.expected"
diff -u "$output/module.expected" "$output/module.actual"

printf 'print(6 * 7);\n' | "$binary" > "$output/stdin.actual"
diff -u "$output/module.expected" "$output/stdin.actual"

set +e
"$binary" -e 'throw new Error("smoke failure");' > "$output/error.stdout" 2> "$output/error.stderr"
error_status=$?
set -e
test "$error_status" = 1
grep -F 'qjs: runtime error' "$output/error.stderr" > /dev/null
grep -F '<cmdline>:' "$output/error.stderr" > /dev/null

"$runner" --harness "$output/unused-harness" testdata/smoke/test262-pass.js > "$output/test262-pass.txt"
grep -F '[raw] PASS' "$output/test262-pass.txt" > /dev/null
set +e
"$runner" --harness "$output/unused-harness" testdata/smoke/test262-fail.js > "$output/test262-fail.stdout" 2> "$output/test262-fail.stderr"
failure_status=$?
set -e
test "$failure_status" = 1
grep -F '[raw] FAIL' "$output/test262-fail.stderr" > /dev/null
