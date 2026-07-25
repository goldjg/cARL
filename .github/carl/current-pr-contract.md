<!-- version: 1.8.0 -->
# Current PR Contract

## Goal

Implement Pack Phase 5: Policy Provenance and Explanation:

1. Add top-level `carl explain <pack-id>` and `carl trace` commands.
2. Explain pack-level policy provenance from canonical repository state:
   source, canonical definition, selection/profile activation, dependency
   expansion, precedence, and override decisions.
3. Report deterministic, schema-versioned human and JSON output.
4. Preserve local-only, read-only, offline behaviour.
5. State explicitly that the commands report policy evaluation provenance,
   not hidden model reasoning or chain-of-thought.

## Contract status

complete

## Previous contract

Pack Phase 4 is delivered. Its completed contract remains historical evidence;
the durable registry, installation, integrity, and offline-first constraints
promoted to memory, trust boundaries, and architecture remain binding.

## Non-goals

- No parsing of natural-language instruction text into individual rules.
- No compiled policy intermediate representation.
- No hidden model reasoning, chain-of-thought, prompt, or session inspection.
- No mutation of selection, profiles, registries, installed provenance, or
  runtime state.
- No conflict auto-resolution or change to existing composition semantics.
- No new registry, publishing, signing, repository-graph, harness, CI, or
  release behaviour.
- No new third-party dependencies.

## Approved scope

- `cmd/carl/main.go`
- `internal/pack/explain.go` (new)
- `internal/pack/explain_command.go` (new)
- `internal/pack/explain_test.go` (new)
- Focused existing tests under `internal/pack/` if required
- `.github/carl/plans/pack-phase5-policy-provenance.md` (new)
- `README.md`, `CLI.md`, `ARCHITECTURE.md`, `GLOSSARY.md`, `ROADMAP.md`
- `.github/carl/memory.md`
- `embedded/assets/.github/carl/memory.md`
- `.github/carl/trust-boundaries.md`
- This contract file

## Forbidden scope

- No writes to `.github/carl/runtime.json`, `.github/carl/packs.json`,
  `.github/carl/profiles.json`, `.github/carl/registries.json`, or
  `.github/carl/installed-packs.json`.
- No network access from `explain` or `trace`.
- No execution of instruction-pack content.
- No inference of policy order from filesystem enumeration.
- No silent suppression or automatic resolution of composition conflicts.
- No edits to harness adapters, instruction packs, CI, release workflows, or
  dependency manifests.
- No destructive repository operations.

## Architectural constraints

- The explained policy unit is an instruction pack. Individual prose rules
  remain opaque until a future structured policy representation is approved.
- Existing pack discovery, profile activation, dependency expansion,
  precedence, and override semantics remain the source of evaluation truth.
- Canonical definitions are repository-relative instruction-pack paths; the
  explanation distinguishes bundled, repository-local, and registry-managed
  sources without making a registry a policy authority.
- `explain` reports one discoverable pack whether or not it is effective.
- `trace` reports the complete effective pack set, ordered exactly like
  `carl pack effective`, plus structured evaluation decisions.
- Permitted overrides are resolved decisions. Invalid or mutual overrides
  remain explicit conflicts and produce a non-zero exit.
- Output is deterministic and uses schema version 1.
- Existing commands and JSON fields remain backward compatible.

## Security constraints

- Treat repository pack, selection, profile, and provenance artefacts as
  untrusted input and reuse their existing strict validation.
- Keep both commands read-only, network-free, and confined to the current
  repository plus embedded assets.
- Emit repository-relative canonical paths only.
- Do not expose secrets, prompts, session data, hidden reasoning, or
  chain-of-thought.
- Fail explicitly on invalid metadata or unresolved composition conflicts.

## Trust boundaries

- Current repository state remains the highest authority for local pack and
  activation facts.
- Pack metadata, selection, profiles, and installed provenance remain
  validated inputs rather than authorities.
- Explanation output is derived diagnostic evidence, not a new canonical
  governance source.
- Registry provenance describes artifact origin and integrity only; it does
  not establish publisher identity.

## Expected files

- `.github/carl/current-pr-contract.md`
- `.github/carl/plans/pack-phase5-policy-provenance.md`
- `cmd/carl/main.go`
- `internal/pack/explain.go`
- `internal/pack/explain_command.go`
- `internal/pack/explain_test.go`
- `README.md`
- `CLI.md`
- `ARCHITECTURE.md`
- `GLOSSARY.md`
- `ROADMAP.md`
- `.github/carl/memory.md`
- `embedded/assets/.github/carl/memory.md`
- `.github/carl/trust-boundaries.md`

## Contract assertions

1. `carl explain <pack-id>` deterministically reports whether the pack is in
   the effective policy, its canonical definition and source, structured
   activation/dependency provenance, precedence, and resolved override state.
2. `carl trace` deterministically reports the full effective set in precedence
   order plus activation, dependency, ordering, override, and conflict
   decisions; unresolved conflicts produce a non-zero exit in human and JSON
   modes.
3. Both commands are schema-versioned, local-only, read-only, and make no
   network or repository writes.
4. Output explicitly states its epistemic boundary: pack-level policy
   provenance is reported, while natural-language rule evaluation, prompts,
   hidden reasoning, and chain-of-thought are not.
5. Existing pack discovery, selection, profile, registry, installation,
   update, and effective-set behaviour remains unchanged and fully validated.

## Validation plan

- `gofmt -w internal/pack/explain.go internal/pack/explain_command.go internal/pack/explain_test.go cmd/carl/main.go`
- `go test ./internal/pack`
- `go test ./...`
- `go vet ./...`
- `go build ./cmd/carl`
- `git diff --check`
- Manual human and JSON smoke checks for `carl explain` and `carl trace`

## cARL/docs update expectation

Expected.

Phase 5 adds durable commands, output semantics, and an explanation-output
trust boundary. CLI, architecture, roadmap, glossary, memory, embedded memory,
and trust-boundary documentation must be reconciled.

## Stop conditions

Stop and escalate if:

- meaningful explanation requires parsing or executing natural-language pack
  contents as individual rules;
- policy provenance requires inspecting prompts, sessions, hidden reasoning,
  or model chain-of-thought;
- the commands require network access or repository mutation;
- implementation requires changing existing composition or conflict semantics;
- validation cannot prove deterministic output and read-only behaviour.

## Escalation triggers

Escalate if:

- a compiled policy intermediate representation becomes necessary;
- a new canonical governance authority is required;
- individual-rule override semantics must be invented;
- compatibility requires changing existing JSON fields or exit codes;
- documentation conflicts with current architecture or durable memory.

## Context reset notes

When complete, retain this contract as the delivered Phase 5 record until the
next task supersedes it. Promote only stable command semantics and the
diagnostic-not-authoritative explanation boundary into durable artefacts.

## Completion checklist

- [x] Implementation stayed inside approved scope.
- [x] Forbidden scope was not touched.
- [x] Contract assertions were validated directly.
- [x] Tests and manual smoke checks were run or gaps reported.
- [x] Documentation and cARL artefacts were reconciled.
- [x] This contract was marked complete.
