# Repository guidelines

gomlquickjs is an independent QuickJS compatibility port written in GoML.
Read `README.md` for supported behavior and limitations, and `UPSTREAM.toml`
for the pinned compatibility baseline. Preserve the existing license.

Use Linux amd64, Go 1.26 or newer on `PATH`, Bash, curl, tar, sha256sum and
`just`. Run `just toolchain` to install the checksum-pinned released GoML
toolchain. Alternatively, set `GOML` to an installed executable while retaining
its executable-relative resources. A sibling compiler repository is unnecessary.

Sources use `.goml`. Run `just fmt` after edits and before tests or commits.
Use `just test` for module tests and `just smoke` for both command-line tools.
Run `just ci` for integration or build changes. Keep build products under
`_artifact/`. Do not add code comments. Preserve recoverable diagnostics and
explicit compatibility limits.

Regenerate Unicode tables with the generator and the pinned upstream sources;
do not hand-edit generated data. Keep implementation and test changes in
separate Conventional Commits.
