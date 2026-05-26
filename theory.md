# A theory of `obsidian-tools`

## What it is

This system models a single fact about the world: when a plugin author submits a plugin to the official Obsidian community directory, the maintainers of `obsidianmd/obsidian-releases` run a set of checks against the plugin's `manifest.json` — some automated, some by-eye — and reject submissions that fail. Those checks are not a schema; they are an accreted, mostly-prose set of conventions (no "obsidian" in the ID or name, no "plugin" suffix, certain banned URLs, a description style guide, etc.).

`validate-plugin-manifest` is a local, pre-flight approximation of that gate. The point is to let a plugin author run on their own machine, before opening a PR, the same checks they will face from a reviewer. Everything else in the design follows from that purpose.

The vocabulary of the code maps directly onto the world it models: `Manifest` is the wire format Obsidian defines; `ValidationResult` is the verdict; "errors" are the things that would block the PR; "warnings" are the things a reviewer would ask you to change but might still merge. The split is not aesthetic — it is the difference between automated rejection and human review, and the code preserves it end-to-end: `IsValid()` looks at errors only, and `main()` keys its exit code off `IsValid()`. A failing validation exits non-zero so CI integrations can gate on it; a manifest with only warnings exits zero, because that is how Obsidian itself behaves.

## The organizing ideas

**The validator is the spec.** There is no JSON schema, no rule DSL, no declarative table of constraints. Each rule is a hand-written Go function (`validateID`, `validateName`, …) that reads a manifest and pushes messages onto a shared `ValidationResult`. This is the right shape for the problem: the upstream rules themselves are heterogeneous — substring checks, regex matches, length limits, exact-URL bans — and any attempt to unify them under a schema would lose more in expressiveness than it gained in tidiness. The price is that the rules are only as discoverable as the function names, and only as testable as someone bothers to write cases for.

**Within a validator, abort on unusable input; across validators, accumulate.** When a required field is empty, the rule reports "X is required" and returns, because the remaining checks for that field would produce nonsense ("regex doesn't match the empty string", etc.). But validators do not abort the whole run on the first error — each validator is called regardless of what the prior one found. A maximally broken manifest reports all nine errors in one pass, not one. This is essential to the user experience: the author runs the tool once and gets a punch list.

The asymmetric early-return (inside, not across) is load-bearing and easy to miss. It was added in the February 2025 refactor; the original code didn't have required-field guards at all, which means "missing field" used to manifest as a cascade of follow-on errors. A future maintainer who tries to "simplify" by either removing the early returns or making them propagate further will reintroduce that noise or hide diagnostics.

**The pair API on `ValidationResult` (`AddError`/`AddErrorf`, `AddWarning`/`AddWarningf`) is stylistic, not structural.** The original collapsed both into one method that always formatted. The split exists so call sites with no interpolation can pass a literal string and read cleanly; it does not encode any semantic difference between formatted and unformatted messages.

**Output is a real seam.** `OutputFormat` is a typed string with two values; `PrintResultsTo(io.Writer, ...)` is the engine; `PrintResults(...)` is a thin stdout-targeting wrapper that exists so `main()` reads cleanly and tests can drive a `bytes.Buffer`. The seam matters because text output is decorative (emoji headers, bullet lists) while JSON output is contractually structured (`{"valid": bool, "errors": [], "warnings": []}` with `omitempty`). The JSON shape is what makes the tool usable from CI. The fact that warnings appear in JSON output even when `valid: true` is part of the contract — a CI job can decide for itself whether to surface them.

**Severity is a binary, not a level.** There is no INFO/WARN/ERROR/FATAL ladder. There are errors and there are warnings, and that maps onto "would Obsidian's automated gate reject this" vs. "would a human reviewer mention this." Inventing more levels would be a category error against the world being modeled.

## The seams

The system's outer boundary is narrow: a file path in, exit code and either text or JSON out. The internal boundary that matters is between `ValidateManifest` (which knows about the rules) and `PrintResultsTo` (which knows about presentation). The `ValidationResult` is the lingua franca across that boundary; nothing else crosses it.

The most consequential external boundary is the one that the code cannot enforce: the relationship between this validator and the upstream `obsidianmd/obsidian-releases/.github/workflows/validate-plugin-entry.yml` it shadows. The rules in `main.go` are a manual transcription of what that workflow checks, performed once. There is no automation, no version pin, no drift detector. If upstream adds a rule, this tool silently keeps passing manifests that would fail there. This is the theory's thinnest point. A maintainer who understands what the tool is for should periodically reconcile against upstream; a maintainer who treats it as a closed system will let it rot quietly.

The workspace-level structure suggests a second seam that does not yet exist. The repo is named `obsidian-tools` (plural), and `CLAUDE.md` describes a "package-per-tool" convention. But the `Taskfile.yml` builds exactly one binary, and the only tool present is the validator. The plural in the name is aspirational; the structure is provisioning for tools that have not arrived. A maintainer adding a second tool needs to know that the `build`/`run`/`BINARY_NAME` task targets are hard-wired to the one current binary and will need to be generalized.

## What's easy, what's not

The system is shaped to absorb new field-level rules cheaply. The recipe — add a `validateThing` function, push messages onto the result, append a call inside `ValidateManifest`, update test counts — is mechanical and isolated. A new banned URL, a new disallowed substring, a new length limit, all fit naturally.

A second output format (say, a JUnit-style XML for CI integration) is also cheap: add an `OutputFormat` constant, add a `printXyzResults`, extend the `switch`. The `io.Writer` plumbing is already there.

What does _not_ fit naturally is anything that breaks the per-field structure. Cross-field rules ("if `isDesktopOnly` is true then `minAppVersion` must be ≥ X.Y") have no idiomatic home — they could go in `ValidateManifest` directly, but doing that erodes the "one validator per field" pattern that holds the file together. A maintainer who needs a few such rules can squeeze them in; a maintainer who needs many should consider whether the underlying pattern still earns its keep.

Making the validator usable as a library is also non-trivial. Everything lives under `package main`, including types that an importer would want (`Manifest`, `ValidationResult`, `ValidateManifest`). Lifting them into an importable package is mostly mechanical but touches every file. The current structure is correct for "a CLI," not "a CLI plus a library."

The test suite is shaped around aggregate behavior, not individual rules. `getInvalidManifestTestCase` constructs a manifest that trips nine rules at once and asserts `wantErrors: 9`. This is efficient but brittle: adding or removing a rule changes the count and the failure message tells you only that the count is wrong, not which rule. A maintainer who adds a rule should expect to update the count in at least one test case and may want to add a focused single-rule test case so future failures localize. The current cases (`valid`, `invalid`, `missing required fields`, `warning only`) are an outline of intent, not exhaustive coverage.

A maintainer who understood the theory would look for upstream-rule changes first when something feels stale. A maintainer who didn't would be tempted to generalize the validators (a `validateField(name, value, rules...)` helper, a rule registry, a schema) and would gain very little while making the code less faithful to its source. The rules read like prose because their source is prose; that is the design, not an accident.

## Uncertainties and tensions

A few claims here are inferences and could be wrong:

- I'm reading the early-return-on-empty pattern as a deliberate noise-suppression mechanism, because the pre-refactor code lacked it and the refactor commit message talks about "improve error handling." That's consistent with the diff but the commit doesn't say so in those words.
- I'm reading the deletion of `testdata/*.json` in the same refactor as a deliberate move from black-box to in-process testing. The commit doesn't justify it. It's possible the fixtures were simply unused and got swept up; the effect is the same either way, but the intent is inferred.
- The pair API split (`AddError` vs `AddErrorf`) might be more than stylistic — perhaps the author wanted `AddError(string)` to be unambiguously safe against accidental format-string injection from a future caller. I cannot confirm this from the code alone; in practice, all current call sites pass static literals.
- The `validateURLs` predicate `manifest.FundingURL != "" && manifest.FundingURL == "https://obsidian.md/pricing"` has a redundant empty check (the equality check already excludes empty strings). This reads as a vestige from the pre-refactor version, where the URL field was treated as conditionally-checked. It is harmless but it is also a small signal that the rule was lifted verbatim rather than re-thought.
- The two version-format validators (`validateVersion`, `validateMinAppVersion`) are nearly identical and could be unified. They aren't, and the "one validator per field" convention argues against unifying them. But the convention itself is unstated; I'm reading it from the shape of the code. A different maintainer might reasonably refactor them and not be wrong.
- The version regex `^[0-9.]+$` is permissive enough that `"."` or `"1..."` pass. The README says version "follows proper format," which oversells what the regex enforces. This is a small discrepancy between stated and actual behavior — probably nobody has cared because real manifests don't look like that.

The largest standing tension is between the repo's plural name and aspirational `CLAUDE.md` ("each tool is a standalone `package main` directory") and the reality of a single tool with build infrastructure that assumes one binary. Either a second tool will arrive and force generalization, or it won't and the framing will continue to overstate the scope. A maintainer should not invest in the multi-tool framing until a concrete second tool is at hand; doing so prematurely would create abstractions in search of a use.
