# Pack Phase 4: Registry and Installation

## Plan metadata

Status: delivered

## Task summary

Deliver the smallest secure vertical slice for explicit pack registries,
version resolution, verified installation, provenance, and updates while
preserving cARL's deterministic, repository-local, offline-first runtime.

## Authority and scope

The active authority is `.github/carl/current-pr-contract.md`. The Pack Phase
4 entry in `ROADMAP.md`, `.github/carl/memory.md`, existing `internal/pack`
semantics, invariants, trust boundaries, and the Go/security/dependency
instruction packs constrain this plan.

No CI, release, harness, repair-ownership, runtime-manifest, policy-explanation,
publishing, or signing-key work is included.

## Data model

### Repository registry configuration

`.github/carl/registries.json`:

- schema version 1;
- unique lower-case kebab-case registry IDs;
- explicit `location` per registry;
- location is either HTTPS or a repository-relative JSON file;
- no implicit or built-in registry.

### Registry index

Each index contains schema version 1 and pack releases with:

- pack ID;
- semantic version;
- relative artifact location;
- lower-case SHA-256 digest;
- optional title and description.

Pack ID plus version is unique within an index.

### Installed provenance

`.github/carl/installed-packs.json` records:

- schema version 1;
- pack ID and installed version;
- source registry ID and configured registry location;
- relative artifact location;
- verified SHA-256 digest;
- repository-relative installed path.

Entries are unique by pack ID and sorted deterministically.

## CLI surface

- `carl pack registry list [--json]`
- `carl pack registry search [<query>] [--registry <id>] [--json]`
- `carl pack install <pack-id> [--version <version>] [--registry <id>] [--json]`
- `carl pack update [<pack-id>...] [--json]`

Search and install fetch only explicit configured registries. Update fetches
only the source registry recorded for each installed pack.

## Resolution and installation flow

1. Strictly load registry configuration and the requested indexes.
2. Validate every registry/index field before candidate selection.
3. Select the highest semantic version, or the exact requested version.
4. Reject an equal winning version from multiple registries unless a registry
   was explicitly selected.
5. Fetch the relative artifact from the same registry base.
6. Enforce response size limits and verify SHA-256.
7. Parse the pack header and verify its declared version.
8. Resolve required dependencies recursively; locally available dependencies
   need no installation.
9. Reject cycles, missing releases, conflicting planned versions, unowned
   existing targets, symlink traversal, and update drift.
10. Validate every artifact and target before writing.
11. Write pack files and the deterministic provenance manifest, rolling back
    files if a write fails.

Install and update do not alter pack selection or profile activation.

## Security model

SHA-256 binds artifact bytes to an explicitly configured registry index. It is
an integrity check, not publisher authentication.

Remote registries require HTTPS. URLs cannot contain credentials, queries, or
fragments. Artifacts must use relative locations, and redirects cannot cross
origin. Local paths and install targets must remain inside the repository.
Symlink traversal is rejected. Registry indexes and pack artifacts have fixed
size limits.

## Contract test mapping

1. Strict deterministic models: configuration/index/provenance round-trip,
   malformed schemas, duplicates, invalid values, and JSON error tests.
2. Optional network boundary: fake fetcher tests prove existing discovery does
   not fetch; local registry install proves offline operation.
3. Resolution: semantic-version ordering, exact version, registry filter, and
   ambiguity tests.
4. Verified transaction: digest mismatch, metadata mismatch, dependency
   expansion, missing dependency, and no-write-on-validation-failure tests.
5. Ownership/path safety: unowned overwrite and symlink traversal rejection;
   selection/runtime files remain absent.
6. Update: recorded-registry use, drift rejection, upgrade, and no-op tests.

## Validation

- format changed Go files;
- run `go test ./internal/pack`;
- run `go test ./...`;
- run `go vet ./...`;
- run `go build ./cmd/carl`;
- inspect `git diff --check` and final repository status.

## Documentation and reconciliation

Update `README.md`, `CLI.md`, `ARCHITECTURE.md`, `GLOSSARY.md`, and
`ROADMAP.md`. Record stable Phase 4 behaviour in `.github/carl/memory.md` and
its embedded mirror. Amend `.github/carl/trust-boundaries.md` for explicit
registry, index, artifact, digest, and provenance trust.
