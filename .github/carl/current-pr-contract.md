<!-- version: 1.5.0 -->
# Current PR Contract

## Goal

Implement Pack Phase 3: Profiles and Agent Roles:

1. Add a schema-versioned, committed `.github/carl/profiles.json` model.
2. Support organisation and repository defaults, named profiles, and
   role-specific and task-specific pack overlays.
3. Add `carl pack profile list`, `show`, `activate`, and `clear`.
4. Make `state.active` and `carl pack effective` profile-driven when the
   profile artefact exists.

## Contract status

active

## Non-goals

- No remote or organisation profile registry.
- No profile inheritance or version resolution.
- No pack dependency downloading.
- No policy-IR compiler or policy explanation command.
- No harness authority model changes.
- No repo-map redesign.

## Approved scope

- `cmd/carl/main.go` (subcommand description only, if required)
- `internal/pack/*`
- Relevant tests under `internal/pack/`
- `.github/carl/profiles.json` as a user-owned artefact (schema only; not
  committed in this repository)
- `README.md`, `CLI.md`, `ARCHITECTURE.md`, `GLOSSARY.md`, `ROADMAP.md`
- `.github/carl/memory.md`
- `embedded/assets/.github/carl/memory.md`
- This contract file

## Forbidden scope

- No edits to CI or release workflows.
- No writes to `.github/carl/runtime.json`.
- No changes to install or repair ownership semantics.
- No unrelated refactors outside the pack package.
- No destructive repository operations.
- No new third-party dependencies.

## Architectural constraints

- Offline-first and deterministic; no runtime network access.
- Profile state is explicit committed data, never inferred from filesystem or
  load order.
- Profiles may activate only packs in the repository selection.
- Organisation defaults, repository defaults, profile packs, role overlays,
  and task overlays compose additively.
- Pack dependencies and precedence continue to use the Phase 2 composition
  rules.
- JSON output remains schema-versioned and stable; JSON errors are structured
  with non-zero exit status.
- When `profiles.json` is absent, selected packs remain active as a legacy
  compatibility fallback.

## Security constraints

- Never commit secrets.
- Treat `profiles.json` as untrusted input and validate every identifier and
  pack reference before use.
- Reject unknown profiles, roles, tasks, packs, duplicate profile IDs, and
  contradictory active contexts explicitly.
- Writes stay inside the repository root.

## Files expected to change

- `.github/carl/current-pr-contract.md`
- `internal/pack/profile.go` (new)
- `internal/pack/profile_test.go` (new)
- `internal/pack/pack.go`
- `internal/pack/compose.go`
- Relevant existing tests under `internal/pack/`
- `README.md`, `CLI.md`, `ARCHITECTURE.md`, `GLOSSARY.md`, `ROADMAP.md`
- `.github/carl/memory.md`
- `embedded/assets/.github/carl/memory.md`

## Contract assertions

1. Profiles are read and written deterministically from
   `.github/carl/profiles.json` with strict schema and identifier validation.
2. Active pack seeds are the additive union of organisation defaults,
   repository defaults, active-profile packs, and the active role/task
   overlays, with an explicit reason for every seed.
3. Profile references must resolve to known selected packs; unknown packs,
   duplicate profiles, and invalid active profile/role/task contexts fail
   explicitly without fallback.
4. `carl pack profile list/show/activate/clear` provides human and
   schema-versioned JSON output; activation writes only `profiles.json`.
5. `carl pack effective` expands active profile seeds through Phase 2
   dependency and precedence rules, while repositories without
   `profiles.json` retain selected-as-active behaviour.

## Tests / validation

- `go test ./internal/pack`
- `go test ./...`
- `go build ./cmd/carl`

## Stop conditions

Stop and escalate if:

- profile activation requires modifying `runtime.json`;
- profile resolution requires harness-specific core semantics;
- preserving existing repositories would require silently accepting malformed
  profile data.

## Escalation triggers

Escalate if:

- organisation defaults require a remote authority or registry;
- profiles require removing packs from the selected set;
- documentation updates conflict with canonical architecture claims.

## Context reset notes

When complete, retain this contract as the delivered Phase 3 record until the
next task supersedes it.
