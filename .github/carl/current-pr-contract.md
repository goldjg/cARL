<!-- version: 1.0.4 -->
# Current PR Contract

## Contract status

Complete

## Goal

Document cARL's interoperability configuration with JFrog Boost v0.14.4 and
later, preserving canonical governance artefacts from Boost file optimisation
while allowing other Boost capabilities to remain enabled. Promote Antigravity
to production support based on successful native validation and keep CLI,
tests, and documentation consistent.

## Non-goals

- Do not add a runtime or build dependency on JFrog Boost.
- Do not change cARL policy evaluation or its trust boundary.
- Do not disable Boost globally or claim a partnership, endorsement,
  certification, or joint support.
- Do not modify embedded assets, harness adapters, or unrelated
  documentation.

## Approved scope

- `.github/carl/current-pr-contract.md` and `.github/carl/memory.md`.
- `README.md`, `CLI.md`, and `ARCHITECTURE.md` for Boost compatibility and
  Antigravity's production support status.
- `internal/harness/harness.go` and focused tests in
  `internal/harness/harness_test.go` and `internal/version/version_test.go`
  for the Antigravity support tier.
- Create a branch, commit the approved scoped changes, push it, and open a
  pull request.
- Preserve unrelated working-tree changes outside these paths.

## Forbidden scope

- Do not rewrite or transform `.github/carl/**` artefacts beyond this active
  contract and the durable memory correction.
- Do not change JFrog Boost configuration outside documentation examples.
- Do not alter cARL's canonical governance authority or harness adapter
  semantics.
- Do not rewrite historical completed plans; current support statements in
  durable memory should reflect user-confirmed production status.
- Do not stage or commit changes outside the approved scope.
- Do not change unrelated local files.

## Architectural constraints

- cARL remains the authority for its governance and policy.
- `.github/carl/**` contains canonical governance/control-plane artefacts and
  must not be a Boost file-optimisation target.
- Harness adapters are projections/loaders, not canonical governance truth.
- Boost remains an independent optional tool; its path-ignore configuration
  defines an interoperability boundary, not a runtime integration.
- Describe Boost path-ignore support as available from v0.14.4 onward.

## Security constraints

- Do not suggest disabling Boost globally.
- Document the narrow repo-relative `.github/carl/**` optimisation exclusion
  and retain other Boost features, including normal CLI filtering, file
  optimisation, MCP optimisation, and code indexing.
- Do not claim that Boost becomes part of cARL's policy engine or trust
  boundary.

## Expected files

- `.github/carl/current-pr-contract.md`
- `README.md`

## Contract assertions

1. cARL supports JFrog Boost v0.14.4+ when `.github/carl/**` is excluded from
   file optimisation, with the requested config and `boost doctor` guidance.
2. Antigravity is production in the adapter registry, tests, CLI output, and
   current documentation; Cursor remains theoretical.
3. The PR contains only the approved Boost interoperability and Antigravity
   production-support changes.

## Validation requirements

- Inspect the final diff and preserve pre-existing edits.
- Run `git diff --check`.
- Validate Markdown links and required configuration/output examples.
- Run any existing Markdown/docs validator if available; do not add tooling
  for this documentation-only change.
- Run `go test ./internal/harness ./internal/version`.
- Verify only the approved files are staged and included in the PR.

## Stop conditions

- Stop if a required change would modify runtime behaviour, embedded assets,
  harness adapter semantics, or files outside the approved scope.
- Stop if JFrog Boost's documented path-ignore support or version threshold
  conflicts with the user-provided configuration requirements.

## Escalation triggers

- Any need to expand beyond the approved documentation and contract files.
- Any ambiguity about whether canonical cARL artefacts should be excluded
  from a broader Boost optimisation mode.

## cARL/docs update expectation

Update this contract to scope the task, the README with durable Boost
interoperability guidance, and `.github/carl/memory.md` to reflect the
user-confirmed Antigravity production-support status. No other canonical
governance artefacts need changes because the Boost configuration is an
optional interoperability boundary, not a cARL policy or runtime change.

## Context reset notes

This contract covers the JFrog Boost compatibility documentation and
Antigravity production support change. Close it after validation and PR
creation.
