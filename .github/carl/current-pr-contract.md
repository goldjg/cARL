<!-- version: 1.7.0 -->
# Current PR Contract

## Goal

Implement Pack Phase 4: Registry and Installation:

1. Add schema-versioned, committed registry configuration and installed-pack
   provenance models.
2. Support explicit HTTPS and repository-local registry indexes.
3. Add `carl pack registry list`, `carl pack registry search`,
   `carl pack install`, and `carl pack update`.
4. Resolve the highest available semantic version deterministically, with an
   exact-version option for installation.
5. Verify every downloaded artifact against its declared SHA-256 digest
   before writing it.
6. Preserve offline-first behaviour: existing pack commands never access the
   network, and repository-local registries support fully offline workflows.

## Contract status

active

## Non-goals

- No public or default registry.
- No registry publishing workflow.
- No certificate or signing-key infrastructure.
- No claim that a SHA-256 digest proves publisher identity.
- No semantic-version ranges, lockfile solver, or profile inheritance.
- No removal/uninstall command.
- No automatic pack selection or profile activation.
- No policy provenance/explanation command (Pack Phase 5).
- No CI, release workflow, harness, or repo-map changes.

## Approved scope

- `cmd/carl/main.go` (pack synopsis only, if required)
- `internal/pack/*`
- Relevant tests under `internal/pack/`
- `.github/carl/registries.json` as a user-owned artefact (schema only; not
  committed in this repository)
- `.github/carl/installed-packs.json` as a user-owned artefact (schema only;
  not committed in this repository)
- `.github/carl/plans/pack-phase4-registry-installation.md`
- `README.md`, `CLI.md`, `ARCHITECTURE.md`, `GLOSSARY.md`, `ROADMAP.md`
- `.github/carl/memory.md`
- `embedded/assets/.github/carl/memory.md`
- `.github/carl/trust-boundaries.md`
- This contract file

## Forbidden scope

- No edits to CI or release workflows.
- No writes to `.github/carl/runtime.json`.
- No changes to install or repair ownership semantics.
- No automatic registry discovery.
- No arbitrary or cross-origin artifact URLs.
- No HTTP plaintext registry transport.
- No execution of downloaded content.
- No destructive repository operations outside rollback of files created by a
  failed pack install/update transaction.
- No new third-party dependencies.

## Architectural constraints

- Existing commands remain deterministic and network-free.
- Registries are configured explicitly in
  `.github/carl/registries.json`; absence means no registry capability is
  configured, not an error for existing commands.
- Registry indexes and provenance use schema version 1 and strict validation.
- Remote registry locations use HTTPS without credentials, query strings, or
  fragments. Local locations are repository-relative and may not escape the
  repository root.
- Artifact locations are relative to their registry index, cannot traverse,
  and therefore remain on the same HTTPS origin or within the repository.
- Resolution chooses the highest semantic version. Equal winning versions
  from multiple registries are an explicit ambiguity unless a registry is
  named.
- Pack files install only under
  `.github/instructions/<category>/<name>.instructions.md`.
- Installation never overwrites a repository-local pack unless it is already
  owned by `.github/carl/installed-packs.json`.
- Update uses the recorded registry and refuses to overwrite a locally drifted
  installed pack.
- Required pack dependencies are resolved and verified before any write.
- Pack selection, activation, dependency composition, and precedence semantics
  remain unchanged.

## Security constraints

- Never commit or log secrets.
- Treat registry configuration, indexes, URLs, artifact bytes, provenance, and
  pack headers as untrusted input.
- Reject credential-bearing URLs, query strings, fragments, plaintext HTTP,
  absolute artifact locations, path traversal, cross-origin redirects,
  oversized responses, malformed JSON, unsupported schemas, duplicate
  versions, invalid identifiers/versions/digests, and digest mismatches.
- SHA-256 verifies artifact integrity against the configured index only; it
  does not authenticate a publisher.
- Validate pack-declared version and composition metadata before writing.
- Refuse symlink traversal for install targets.
- Keep every write inside the repository root and make multi-pack operations
  rollback on failure.

## Files expected to change

- `.github/carl/current-pr-contract.md`
- `.github/carl/plans/pack-phase4-registry-installation.md`
- `internal/pack/pack.go`
- `internal/pack/registry.go` (new)
- `internal/pack/registry_command.go` (new)
- `internal/pack/registry_test.go` (new)
- Relevant existing tests under `internal/pack/`
- `README.md`, `CLI.md`, `ARCHITECTURE.md`, `GLOSSARY.md`, `ROADMAP.md`
- `.github/carl/memory.md`
- `embedded/assets/.github/carl/memory.md`
- `.github/carl/trust-boundaries.md`

## Contract assertions

1. Registry configuration, indexes, and provenance are strict,
   schema-versioned, deterministic models; malformed or ambiguous data fails
   explicitly with structured JSON errors when requested.
2. Existing pack commands perform no network access; registry search/install/
   update access only explicitly configured sources, and local registry
   locations work without a network.
3. Resolution is deterministic, selects the highest semantic version, supports
   exact-version installation, and rejects equal-version cross-registry
   ambiguity unless the caller selects a registry.
4. Installation verifies SHA-256 and pack metadata, resolves required
   dependencies, validates the complete operation, and performs no writes
   before all inputs are verified.
5. Installs write only validated instruction-pack paths plus
   `installed-packs.json`; they do not select/activate packs or write
   `runtime.json`, and they do not overwrite unowned local packs.
6. Updates use recorded provenance, reject drift, remain idempotent when no
   newer version exists, and accurately report integrity verification without
   claiming publisher authentication.

## Tests / validation

- `gofmt -w internal/pack/*.go cmd/carl/main.go`
- `go test ./internal/pack`
- `go test ./...`
- `go vet ./...`
- `go build ./cmd/carl`

## Stop conditions

Stop and escalate if:

- safe installation requires writing `runtime.json`;
- registry artifacts require arbitrary/cross-origin URLs or credential
  persistence;
- existing repository-local packs must be silently overwritten;
- dependency resolution cannot validate the complete write set before mutation.

## Escalation triggers

Escalate if:

- publisher authentication requires designing signing-key trust roots;
- remote discovery requires an implicit/default external authority;
- compatibility requires changing existing selection, activation, precedence,
  or repair ownership semantics;
- documentation updates conflict with canonical architecture claims.

## Context reset notes

When complete, retain this contract as the delivered Phase 4 record until the
next task supersedes it.
