# Obsidian Tools

A workspace of Go tools for Obsidian plugin development. Currently ships one tool: `validate-plugin-manifest`, a local pre-flight check that approximates the rules Obsidian's [validate-plugin-entry workflow](https://github.com/obsidianmd/obsidian-releases/blob/master/.github/workflows/validate-plugin-entry.yml) applies when reviewing community plugins.

This README is for maintainers. End-user usage is summarized briefly at the bottom.

## Quick start

```bash
git clone https://github.com/philoserf/obsidian-tools.git
cd obsidian-tools
brew bundle      # installs go, golangci-lint, prettier
task build       # produces ./obsidian-validate-plugin-manifest
task ci          # lint + test
```

Go 1.25, standard library only. No third-party dependencies.

## Where to read next

- [theory.md](theory.md) — the conceptual model a maintainer needs to hold in mind: what the system is for, the load-bearing invariants, where the seams are, what changes are easy vs. hard.
- [walkthrough.md](walkthrough.md) — linear, executable tour of the code from `main()` outward. Re-verifiable with `uvx showboat verify walkthrough.md`.
- [CLAUDE.md](CLAUDE.md) — guidance for Claude Code sessions; duplicates some of the conventions below.

## Layout

```
.
├── validate-plugin-manifest/   # the one tool — package main
│   ├── main.go                 # CLI, Manifest, ValidationResult, validators
│   ├── output.go               # OutputFormat, text + JSON renderers
│   ├── validator_test.go       # table-driven validator tests
│   └── output_test.go          # output renderer tests
├── Taskfile.yml                # build/test/lint/fix/run/ci
├── Brewfile                    # dev dependencies
├── go.mod                      # module root, Go 1.25
└── .github/workflows/          # Pages publish only (no CI build/test yet)
```

Each tool is its own `package main` directory. Adding a second tool means adding a sibling directory and extending the `build`/`run` targets in `Taskfile.yml`, which are currently hard-wired to the one binary.

## Development workflow

```bash
task build   # go build into ./obsidian-validate-plugin-manifest
task test    # go test ./...                  (pass flags via `--`, e.g. -cover)
task lint    # golangci-lint run ./...
task fix     # golangci-lint --fix, gofumpt, prettier on markdown
task run     # build + run the binary         (pass flags via `--`, e.g. --json)
task ci      # lint + test
```

`task` with no args lists everything. Pass extra flags after `--`:

```bash
task test -- -run TestValidateManifest -cover
task run  -- --manifest path/to/manifest.json --json
```

Direct `go` commands work too — Task is convenience, not requirement.

## Code standards

- **Go 1.25**, standard library only. Don't add dependencies without a strong reason.
- **Formatter:** `gofumpt` (stricter than `gofmt`). `task fix` applies it.
- **Linter:** `golangci-lint`. `task lint` runs it; `task fix` auto-fixes what it can.
- **Markdown:** `prettier`. Same `task fix` covers it.
- Pre-compile regexes as package-level `var`. See `main.go` for the pattern.
- Tests are table-driven with `t.Parallel()` at both the outer test and each subtest. Use a `TestCase` struct and a `check*` helper for assertions.

## Adding a validation rule

The shape is the same across all current rules — see `theory.md` for why.

1. Write `validateThing(manifest *Manifest, result *ValidationResult)` in `main.go`.
2. If the field is required, guard with an empty check that `AddError`s and returns. Otherwise let subsequent checks accumulate.
3. Use `AddError`/`AddWarning` for literal messages, `AddErrorf`/`AddWarningf` when formatting is needed.
4. Call `validateThing(manifest, &result)` from `ValidateManifest`.
5. Update `validator_test.go` — at minimum, adjust the `wantErrors` / `wantWarnings` counts in the affected `getXxxTestCase` factory. Tests assert on counts, not messages, so a new rule will shift them.

Errors block (CLI exits non-zero); warnings don't. That distinction maps onto Obsidian's automated-rejection vs. human-reviewer split — don't blur it.

## Adding an output format

`OutputFormat` is a typed string in `output.go`. To add one:

1. Add a new constant (e.g. `JUnitOutput OutputFormat = "junit"`).
2. Write `printJUnitResults(writer io.Writer, result ValidationResult)`.
3. Extend the `switch` in `PrintResultsTo`.
4. Wire a flag in `main()` if it should be user-selectable.

Renderers write to an `io.Writer`, not stdout directly, so tests can drive a `bytes.Buffer`.

## Releases

No release automation. Built locally; the binary is gitignored. The Jekyll Pages workflow publishes the README only and is not part of the build pipeline.

Dependabot watches `gomod` and `github-actions` weekly (`.github/dependabot.yml`).

## Upstream drift

The validation rules are a manual transcription of [Obsidian's validate-plugin-entry workflow](https://github.com/obsidianmd/obsidian-releases/blob/master/.github/workflows/validate-plugin-entry.yml). There is no automated drift check. When in doubt, diff against upstream and reconcile. See `theory.md` for context.

## End-user reference

```bash
./obsidian-validate-plugin-manifest                                     # validate ./manifest.json
./obsidian-validate-plugin-manifest --manifest path/to/manifest.json
./obsidian-validate-plugin-manifest --manifest path/to/manifest.json --json   # for CI
./obsidian-validate-plugin-manifest --quiet                             # errors/warnings only
./obsidian-validate-plugin-manifest --version
```

Exit code is non-zero when validation reports any errors. Warnings alone don't fail the run.
