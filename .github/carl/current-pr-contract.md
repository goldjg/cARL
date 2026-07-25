<!-- version: 1.9.0 -->
# Current PR Contract

## Goal

Implement Pack Phase 6: Cognitive Repository Graph as a bounded,
backward-compatible extension of `carl map`:

1. Add a schema-versioned graph to `.github/carl/repo-map.json`.
2. Represent repository components, key artefacts, Go package dependencies,
   trust-boundary classifications, policy attachment points, criticality,
   agent-relevant context, and direct change impact.
3. Report the evidence coverage and limitations of ownership, dependency,
   data-flow, policy, and change-impact knowledge.
4. Preserve deterministic, local-only, offline behaviour and all existing
   inventory fields consumed by `carl reconcile`.

## Contract status

complete

## Previous contract

Pack Phase 5 is delivered. Its completed contract is historical evidence.
The durable deterministic, schema-versioned, local-only, provenance, and
diagnostic-not-authoritative constraints promoted to memory, architecture,
and trust boundaries remain binding.

## Non-goals

- No runtime instrumentation or dynamic data-flow tracing.
- No speculative ownership, runtime-flow, or policy-activation inference.
- No replacement of `carl trace` as the policy activation authority.
- No natural-language parsing of documentation or instruction-pack rules.
- No compiled policy intermediate representation.
- No repository-graph query language, visual UI, database, or remote service.
- No changes to `carl reconcile` output or managed-marker semantics.
- No new third-party dependencies.

## Approved scope

- `internal/repomap/repomap.go`
- `internal/repomap/graph.go` (new)
- `internal/repomap/repomap_test.go`
- `internal/repomap/graph_test.go` (new)
- Focused existing tests under `internal/reconcile/` if compatibility requires
- `.github/carl/current-pr-contract.md`
- `.github/carl/plans/pack-phase6-cognitive-repository-graph.md` (new)
- `.github/carl/repo-map.json`
- `.github/carl/repo-map.example.json`
- `README.md`, `CLI.md`, `ARCHITECTURE.md`, `GLOSSARY.md`, `ROADMAP.md`
- `.github/carl/memory.md`
- `embedded/assets/.github/carl/memory.md`
- `.github/carl/trust-boundaries.md`

## Forbidden scope

- No edits to harness adapters, instruction packs, CI, release workflows,
  dependency manifests, pack selection/profile/registry/provenance state, or
  `.github/carl/runtime.json`.
- No network access or execution of repository content.
- No following symlinks outside the repository.
- No inferred policy activation or precedence.
- No claims that static imports prove runtime data flow.
- No destructive repository operations.

## Architectural constraints

- Existing top-level repo-map inventory fields and JSON meanings remain
  backward compatible.
- The graph schema is version 1 and uses stable repository-relative node IDs.
- Nodes and edges are sorted deterministically; filesystem enumeration order
  never becomes graph order or dependency precedence.
- Go package dependencies come only from statically parsed repository-local
  import declarations.
- Direct change impact is the reverse of observed repository-local dependency
  edges; it is not a transitive or runtime impact guarantee.
- Policy definitions may be represented as graph nodes, while component and
  package nodes may be marked as policy attachment points. Active policy
  assignment remains the responsibility of `carl trace`.
- Every graph knowledge facet reports whether evidence is derived, partial,
  or unavailable.

## Security constraints

- Treat repository paths, Go source, and governance artefacts as untrusted
  local input.
- Parse Go imports without compiling or executing repository code.
- Emit only repository-relative paths and stable identifiers.
- Keep writes confined to `.github/carl/repo-map.json`.
- Do not expose secrets, environment values, prompts, session data, or hidden
  reasoning.

## Trust boundaries

- Current repository state is the authority for structural graph evidence.
- Source imports and filesystem paths are validated evidence, not proof of
  runtime execution or data flow.
- Generated graph output is derived orientation evidence, not a canonical
  governance or ownership authority.
- `carl trace` remains authoritative diagnostic evidence for active policy
  evaluation; graph policy attachment points do not override it.

## Expected files

- `.github/carl/current-pr-contract.md`
- `.github/carl/plans/pack-phase6-cognitive-repository-graph.md`
- `internal/repomap/repomap.go`
- `internal/repomap/graph.go`
- `internal/repomap/repomap_test.go`
- `internal/repomap/graph_test.go`
- `.github/carl/repo-map.json`
- `.github/carl/repo-map.example.json`
- `README.md`
- `CLI.md`
- `ARCHITECTURE.md`
- `GLOSSARY.md`
- `ROADMAP.md`
- `.github/carl/memory.md`
- `embedded/assets/.github/carl/memory.md`
- `.github/carl/trust-boundaries.md`

## Contract assertions

1. `carl map` preserves all existing inventory fields and adds
   `schema_version: 1` plus a deterministic graph of stable, unique,
   repository-relative nodes and edges.
2. Repository-local Go imports produce dependency edges, and direct reverse
   dependants produce explicit change-impact references without compiling or
   executing repository code.
3. Graph nodes classify criticality and trust boundaries, identify component
   policy attachment points, and provide agent context without claiming active
   policy assignment.
4. Graph coverage explicitly distinguishes derived, partial, and unavailable
   ownership, dependency, data-flow, policy, and change-impact evidence.
5. `carl map` remains local-only and idempotent, and existing `carl reconcile`
   behaviour remains compatible with the extended JSON.

## Validation plan

- `gofmt -w internal/repomap/repomap.go internal/repomap/graph.go internal/repomap/repomap_test.go internal/repomap/graph_test.go`
- `go test ./internal/repomap`
- `go test ./internal/reconcile`
- `go test ./...`
- `go vet ./...`
- `go build ./cmd/carl`
- `git diff --check`
- Manual repeated `carl map` smoke check and graph JSON inspection
- Manual `carl reconcile` compatibility smoke check without retaining
  unrelated generated changes

## cARL/docs update expectation

Expected.

Phase 6 changes the durable repo-map schema, command output, orientation model,
and graph-output trust boundary. CLI, architecture, roadmap, glossary, memory,
embedded memory, trust boundaries, and the repo-map example must be reconciled.

## Stop conditions

Stop and escalate if:

- useful graph generation requires executing repository code;
- ownership or runtime-flow claims require unsupported inference;
- backward compatibility requires removing or changing existing fields;
- implementation requires a graph database, new dependency, network access,
  or a new canonical governance authority;
- validation cannot prove deterministic IDs, edges, and read-only scanning.

## Escalation triggers

Escalate if:

- graph metadata needs to become authoritative policy state;
- `carl trace` semantics must change;
- a human-owned graph override artefact becomes necessary;
- CI, release, harness, runtime state, or dependency policy changes become
  necessary;
- repository graph output would expose absolute paths or sensitive data.

## Context reset notes

When complete, retain this contract as the delivered Phase 6 record until the
next task supersedes it. Promote only stable graph schema semantics,
evidence-quality boundaries, and `carl map` compatibility facts into durable
artefacts.

## Completion checklist

- [x] Implementation stayed inside approved scope.
- [x] Forbidden scope was not touched.
- [x] Contract assertions were validated directly.
- [x] Tests and manual smoke checks were run or gaps reported.
- [x] Documentation and cARL artefacts were reconciled.
- [x] This contract was marked complete.
