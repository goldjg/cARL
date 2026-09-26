<!-- version: 1.0.1 -->
# Current PR Contract

## Contract status

Complete

## Goal

Implement an initial optional semantic evaluation capability in the cARL CLI,
using a provider-neutral `SemanticEvaluator` abstraction and TypeSafe JEV as
the first supported backend, with `carl review` as the initial vertical slice
for bounded local or CI code review.

## Non-goals

- Do not make JEV or any hosted provider mandatory.
- Do not make cARL a model router, generic agent framework, or generic
  evaluator marketplace.
- Do not let JEV or any provider decide cARL policy outcomes directly.
- Do not replace deterministic cARL rules, pack composition, or governance
  artefacts.
- Do not send whole repositories, environment variables, secrets, credentials,
  private keys, or obvious secret-bearing file contents to external providers.
- Do not add multiple providers beyond the first JEV adapter.
- Do not add SARIF or GitHub Checks output in this PR.
- Do not modify release workflows, publishing automation, or unrelated
  harness adapter behaviour.

## Approved scope

- New provider-neutral semantic evaluation models, interfaces, context/state
  builder, review policy mapping, output formatting, and failure handling.
- A TypeSafe JEV evaluator adapter that normalizes provider responses into
  cARL-owned semantic signal types.
- `carl review` with bounded git diff collection, optional task/PR text input,
  dry-run payload inspection, text/JSON/Markdown output, and deterministic
  exit codes.
- Optional repository configuration for semantic evaluation under a cARL-owned
  config file, including API-key environment-variable lookup only.
- Unit and integration-style tests that mock provider behaviour and avoid live
  network/API-key requirements.
- Documentation updates describing configuration, privacy/trust boundary,
  dry-run, CI usage, and the distinction between semantic signals and cARL
  policy decisions.
- Durable cARL memory/trust-boundary updates when needed to record the new
  external semantic-evaluation boundary.

## Forbidden scope

- Do not persist API keys or accept API keys directly in repository
  configuration.
- Do not log, print, serialize, or include secrets/API keys in diagnostics,
  exceptions, dry-run payloads, JSON output, telemetry, or files.
- Do not hard-code TypeSafe or JEV concepts into provider-independent cARL
  policy models.
- Do not execute tests or repository commands as part of review state
  collection unless explicitly supplied or already available.
- Do not perform remote mutations, create releases, alter CI permissions, or
  publish artifacts.
- Do not modify unrelated local files or overwrite user-owned policy state
  outside this contract.

## Architectural constraints

- JEV is an optional semantic signal provider; cARL remains the policy
  authority.
- The generic evaluator interface, state, question, signal, finding, and
  result models must not be named after JEV.
- The JEV adapter may translate between cARL-owned bounded question primitives
  and provider-specific API shapes, but must not encode cARL policy decisions.
- Context collection and external-provider payload construction must remain
  clearly separated and inspectable.
- Review context must be bounded, deterministic where practical, and auditable
  by category; avoid blindly reading or sending the full repository.
- Use existing Go standard-library-first patterns and repository command
  conventions; add dependencies only with explicit justification.

## Security constraints

- Treat semantic evaluator providers as external hosted services.
- Redact obvious secrets from collected diff/context and dry-run payloads.
- Fail open by default when semantic evaluation is unavailable or unconfigured,
  producing warnings/diagnostics without breaking existing cARL behaviour.
- Handle missing/invalid API key, network failures, timeouts, malformed
  responses, unsupported model, rate limiting, and provider outage without
  exposing secrets or confusing stack traces in normal output.
- Ensure tests never require real network access or a live TypeSafe API key.

## Expected files

- `cmd/carl/main.go`
- `internal/semantic/**`
- `internal/review/**`
- `internal/cmdutil/**` if exit-code support requires small shared helpers
- `CLI.md`
- `README.md`
- `ARCHITECTURE.md`
- `.github/carl/current-pr-contract.md`
- `.github/carl/memory.md` and/or `.github/carl/trust-boundaries.md` if
  durable semantic-evaluation boundary documentation is needed
- Focused `*_test.go` files for new behaviour

## Contract assertions

1. When semantic evaluation is disabled or unconfigured, existing cARL
   behaviour remains unchanged and `carl review` fails open with clear
   diagnostics rather than requiring JEV.
2. `carl review --dry-run` reports the semantic questions and data categories
   that would be sent without calling the provider and without exposing
   secrets.
3. JEV responses normalize into provider-independent semantic signals; cARL
   review policy maps those signals to findings and a decision.
4. Text, JSON, and Markdown outputs represent the same normalized review
   result, and JSON includes complete normalized signals/findings/decision for
   CI consumption.
5. Exit codes distinguish pass, warn/review-required/fail, and
   execution/configuration failure without requiring human-readable parsing.

## Validation requirements

- Run focused tests for semantic evaluator, JEV normalization, review policy,
  redaction, dry-run, output formats, and exit codes.
- Run `go test ./...` after implementation when feasible.
- Run `go build ./cmd/carl` and `git diff --check`.
- Confirm no real network/API key is required by tests.

## Stop conditions

Stop and report if:

- the provider API contract cannot be represented without embedding cARL
  policy inside the JEV adapter;
- implementing the feature requires storing secrets in repository files;
- bounded context collection cannot avoid sending obvious secret-bearing
  content;
- a required change would alter release automation, publish artifacts, or
  rewrite unrelated governance/harness state;
- validation requires live TypeSafe credentials or external network access.

## Escalation triggers

- Any need for a new dependency.
- Any ambiguity in the TypeSafe JEV API that would materially affect external
  request shape beyond a documented initial adapter assumption.
- Any proposed fail-closed CI behaviour before repository policy explicitly
  requires semantic evaluation.
- Any trust-boundary change beyond optional outbound semantic evaluation.

## cARL/docs update expectation

Required. This PR adds a new optional command, external provider trust
boundary, and durable architectural behaviour, so CLI/docs and durable cARL
memory/trust-boundary documentation must be reconciled before final response.

## Context reset notes

After this PR completes, reset the contract or mark it complete so future
tasks do not inherit semantic-evaluation implementation scope.

## Completion evidence

Completed on 2026-09-26 in the `feature/semantic-evaluation-jev` worktree.
The implementation adds provider-neutral semantic evaluation models,
bounded review state collection, the optional TypeSafe JEV adapter, `carl
review`, dry-run payload inspection, text/JSON/Markdown output, fail-open
diagnostics, deterministic decision exit codes, tests, and documentation for
the new external semantic-evaluation trust boundary. Validation completed with
`go test ./...`, `go vet ./...`, `go build -buildvcs=false ./cmd/carl`, and
`git diff --check`.
