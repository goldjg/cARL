# `carl plan` Command Plan

## Plan metadata
- PR / branch: feature/carl-plan-command
- Status: Completed
- Author: goldjg
- Created: 2026-06-18
- Last updated: 2026-06-18

## Task summary
Add `carl plan` as a first-class CLI command that treats `.github/carl/plans/*.md`
as cARL artefacts — discovering, validating, and summarising them.

## Goal
Provide operators and agents with a single command to inspect the state of all
plan files in the repository. Show title, lifecycle status, and purpose for each
plan; surface structural validation warnings inline.

## Approved scope
- `internal/plan/plan.go` — command implementation
- `internal/plan/plan_test.go` — contract-assertion tests
- `cmd/carl/main.go` — register `plan` command
- `CLI.md` — command reference documentation
- `.github/carl/memory.md` — durable facts update
- `ROADMAP.md` — mark delivered
- `.github/carl/current-pr-contract.md` — active contract

## Acceptance criteria
1. `carl plan` with no plans directory outputs "No plans found." and exits 0.
2. `carl plan` with valid plans outputs title, status, and purpose for each.
3. Plans missing `## Plan metadata` show an inline Warning.
4. Plans with empty `Status:` show an inline Warning.
5. Plans are listed sorted lexicographically by filename.
6. Command always returns nil (read-only; never modifies files).

## Contract assertions
1. No plans directory or empty directory → "No plans found." and return nil.
2. Fully-populated plan → correct title/status/purpose, no warnings.
3. Missing `## Plan metadata` → warning: "missing ## Plan metadata section".
4. Empty `Status:` → warning: "Status not set in ## Plan metadata".
5. Sorted lexicographically by filename.
6. Always returns nil.
7. Non-.md files are silently ignored.
