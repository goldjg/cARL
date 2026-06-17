<!-- version: 1.2.0 -->
# Durable Architectural Truth Cache

This cache stores durable project truths that should persist beyond a
single task. Update it only when a stable fact, decision, invariant, or
unresolved question should carry forward.

## Project purpose
cARL (Cognitive Agent Runtime Layer) is a reusable governance and
instruction layer for GitHub Copilot coding agents. It provides modular
instruction packs, durable memory artefacts, and cARLv2 cognition
governance for consistent, secure, maintainable, and governed
AI-assisted development.

## Non-goals
<!-- Populate with explicitly out-of-scope outcomes to prevent scope creep. -->

## Architecture summary
`.github/copilot-instructions.md` is the root operating model.
`.github/instructions/` contains modular single-concern instruction
packs. `.github/carl/` contains cARLv2 governance artefacts and
templates, not instruction-pack logic.

## Core invariants
- Instruction packs should remain modular and focused on a single
  concern.
- The root Copilot instructions act as the repository constitution.
- Existing language, platform, cloud, and core packs should not be
  modified as side effects of unrelated changes.
- cARLv2 artefacts should reduce semantic rediscovery without becoming
  a per-turn session diary.
- Prompt-as-code should be used for substantial, long, or
  boundary-sensitive agent tasks.

## Trust boundaries
<!-- Populate with critical trust boundaries and required validation expectations. -->

## Known sharp edges
- Long nested prompts in agent UIs may truncate or misparse; prefer
  committed plan files for boundary-sensitive work.
- Agents may over-anchor on completed PR contracts; distinguish durable
  invariants from historical PR constraints.
- Model availability and capability can vary; fallback models must
  preserve the active PR contract.
- Repeated corrective prompting is a failure signal; reset the session
  or switch model instead of continuing prompt ping-pong.

## Field findings
- PR3 review gap: green tests alone do not prove contract compliance;
  validate that tests assert approved acceptance criteria rather than
  implementation drift.
- Non-trivial output-schema-heavy, validation-heavy, trust-boundary, or
  failure-mode-sensitive work should define 3-5 contract assertions
  before implementation and check them during validation.

## Canonical validation commands
<!-- Populate with validated commands that prove expected behavior in this repository. -->

## Current operating assumptions
Model availability is not a stable invariant. The PR contract remains
the source of truth across model fallback.

## Open questions
<!-- Populate with unresolved questions that should persist into future work. -->

## Last updated
2026-06-17 by cARL bootstrap
