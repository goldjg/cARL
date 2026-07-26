<!-- version: 1.9.0 -->
# cARL CLI Reference

The `carl` CLI installs and manages the cARL governance runtime inside a repository.
It is a self-contained binary with no runtime dependencies — all governance artefacts
are embedded at build time.

---

## Installation

### Download a pre-built binary

Pre-built binaries for Linux, macOS, and Windows are attached to every
[GitHub Release](https://github.com/goldjg/cARL/releases/latest).
Replace `v1.0.0` in the commands below with the desired release tag.

**Linux (amd64)**
```sh
curl -L https://github.com/goldjg/cARL/releases/download/v1.0.0/carl-v1.0.0-linux-amd64 \
  -o carl && chmod +x carl && sudo mv carl /usr/local/bin/carl
```

**macOS (Apple Silicon)**
```sh
curl -L https://github.com/goldjg/cARL/releases/download/v1.0.0/carl-v1.0.0-darwin-arm64 \
  -o carl && chmod +x carl && sudo mv carl /usr/local/bin/carl
```

**macOS (Intel)**
```sh
curl -L https://github.com/goldjg/cARL/releases/download/v1.0.0/carl-v1.0.0-darwin-amd64 \
  -o carl && chmod +x carl && sudo mv carl /usr/local/bin/carl
```

**Windows (amd64)**

Download `carl-v1.0.0-windows-amd64.exe` from the
[releases page](https://github.com/goldjg/cARL/releases/latest) and add it to your `PATH`.

### Build from source

Requires Go 1.24 or later.

```sh
git clone https://github.com/goldjg/cARL.git
cd cARL
go build -ldflags "-X main.cliVersion=$(git describe --tags --always)" -o carl ./cmd/carl
```

---

## Commands

### `carl init`

Installs the cARL runtime into the current repository.

**Usage**

```
carl init [--adopt]
```

**What it does**

1. Checks that the runtime is not already installed (no `.github/carl/runtime.json`).
2. Checks that none of the managed artefact paths already exist. If any do, it
   lists the conflicts and exits without writing any files.
3. Writes all governance artefacts into `.github/` (instruction packs, `carl/`
   governance files).
4. Creates `.github/carl/runtime.json` — the authoritative runtime manifest
   recording the installed version, source tag, commit, timestamp, and list of
   managed artefacts.

**Adopting existing artefacts**

Use `carl init --adopt` when cARL artefacts already exist but
`.github/carl/runtime.json` does not. Adoption preserves every existing
artefact byte-for-byte, installs only missing bundled artefacts, and creates
the runtime manifest last. Run `carl doctor` afterward to inspect drift and
invoke `carl repair` separately if canonical repairable content is wanted.
`memory.md` remains protected from repair.

**Output (success)**

```
cARL runtime installed successfully.
  Runtime version:  1.0.0
  Source:           goldjg/cARL @ v1.0.0
  Artefacts:        32 files installed
```

**Errors**

| Error | Cause | Resolution |
|---|---|---|
| `cARL runtime already installed` | `runtime.json` already exists | Run `carl repair` to restore drift, or remove the runtime manually |
| `cARL artefacts already exist` | Individual managed files exist without a `runtime.json` | Run `carl init --adopt` to preserve and adopt them, or remove the listed files for a clean installation |

---

### `carl repair`

Restores modified managed cARL artefacts to their canonical state.

**Usage**

```
carl repair
```

**What it does**

1. Reads `runtime.json` to discover the list of managed artefacts.
2. Compares each managed artefact against its embedded canonical version
   (byte-for-byte).
3. Reports any files that differ or are missing (drift).
4. Overwrites drifted files with their canonical versions.

**Protected files** — the following are never overwritten by `repair`:

- `.github/carl/memory.md` — per-repository state managed by humans and agents
- `.github/carl/runtime.json` — managed exclusively by `carl init`

**Output (no drift)**

```
No drift detected.
```

**Output (drift found)**

```
Drift detected:
  .github/instructions/core/security.instructions.md
  .github/copilot-instructions.md

Repairing...
Done.
```

**Errors**

| Error | Cause | Resolution |
|---|---|---|
| `no cARL runtime installed` | `runtime.json` not found | Run `carl init`, or `carl init --adopt` when cARL artefacts already exist |

---

### `carl doctor`

Diagnoses runtime issues and provides actionable remediation guidance.

**Usage**

```
carl doctor
```

**What it does**

1. Reads `runtime.json` to discover the installed runtime state. If the manifest
   is absent, reports it as an ERROR and suggests `carl init`.
2. Detects and categorises findings as ERROR, WARNING, or INFO:
   - **ERROR** — missing runtime manifest, unreadable manifest, artefact absent from disk
   - **WARNING** — artefact content differs from its canonical version (drifted), or a harness adapter is missing/drifted
   - **INFO** — runtime health and production harness support
3. For each finding, provides a suggested remediation action.
4. Exits with code `0` regardless of whether issues are found — the command is
   diagnostic only and never modifies any files.

**Protected files** — the following are never inspected for drift:

- `.github/carl/memory.md` — per-repository state managed by humans and agents
- `.github/carl/runtime.json` — managed exclusively by `carl init`

**Output (no runtime installed)**

```
ERROR   missing runtime manifest (.github/carl/runtime.json)
        Action: run `carl init`

1 error(s), 0 warning(s), 0 info(s) found.
```

**Output (healthy runtime)**

```
INFO    runtime is healthy — all managed artefacts are present and canonical
INFO    production harnesses: copilot, claude, codex
```

**Output (missing and drifted artefacts)**

```
ERROR   .github/carl/invariants.yml — artefact is missing from disk
        Action: run `carl repair`
WARNING .github/copilot-instructions.md — artefact has drifted from its canonical version
        Action: run `carl repair`
WARNING claude (CLAUDE.md) — harness adapter file has drifted from its canonical version
        Action: run `carl harness sync`
INFO    production harnesses: copilot, claude, codex

1 error(s), 2 warning(s), 1 info(s) found.
```

**Finding levels**

| Level | Meaning |
|---|---|
| `ERROR` | Condition that prevents normal operation; immediate action required |
| `WARNING` | Condition that should be addressed; runtime still functional |
| `INFO` | Neutral observation; no action required |

---

Reports whether the installed cARL runtime is healthy, missing, or drifted.

**Usage**

```
carl status
```

**What it does**

1. Reads `runtime.json` to discover the installed runtime state.
2. Derives installed pack names from the managed artefact paths.
3. Compares each managed repairable artefact against its embedded canonical
   version (byte-for-byte), classifying files as missing (absent from disk) or
   drifted (present but content differs).
4. Inspects harness adapters and reports a separate summary of detected, missing,
   drifted, healthy, and production adapters.
5. Reports overall runtime status: `Healthy`, `Drifted`, or `Incomplete`.

**Protected files** — the following are never reported as missing or drifted:

- `.github/carl/memory.md` — per-repository state managed by humans and agents
- `.github/carl/runtime.json` — managed exclusively by `carl init`

**Output (runtime installed, healthy)**

```
CLI Version:      1.0.0
Runtime Version:  1.0.0
Source:           goldjg/cARL
Tag:              v1.0.0
Commit:           abc1234

Installed Packs:
  cloud/azure
  core/carl
  ...

Missing Artefacts:
  none

Drifted Artefacts:
  none

Harness Summary:
  Detected adapters: 5
  Missing adapters: 0
  Drifted adapters: 0
  Healthy adapters: 5
  Production:       copilot, claude, codex

Status:           Healthy
```

**Output (missing and drifted artefacts)**

```
...
Missing Artefacts:
  .github/carl/invariants.yml

Drifted Artefacts:
  .github/copilot-instructions.md

Harness Summary:
  Detected adapters: 4
  Missing adapters: 1
  Drifted adapters: 1
  Healthy adapters: 3
  Production:       copilot, claude, codex

Status:           Incomplete
```

**Output (no runtime installed)**

```
No cARL runtime installed.
```

**Status values**

| Status | Meaning |
|---|---|
| `Healthy` | All managed repairable artefacts are present and match their canonical versions |
| `Drifted` | All artefacts present, but one or more differ from their canonical versions; run `carl repair` |
| `Incomplete` | One or more managed artefacts are absent from disk; run `carl repair` |

---

### `carl map`

Generates and updates `.github/carl/repo-map.json` by deriving the repository
inventory and an evidence-scoped cognitive graph from the filesystem.

**Usage**

```
carl map
```

**What it does**

1. Walks the repository filesystem from the current working directory.
2. Detects programming languages from source file extensions.
3. Identifies project entry points (`go.mod`, `cmd/*/main.go`, `Makefile`, etc.).
4. Maps key directories (up to three levels deep) with human-readable purpose
   descriptions derived from Go package doc comments or known-path heuristics.
5. Lists GitHub Actions workflows from `.github/workflows/`.
6. Lists governance artefacts from `.github/carl/`.
7. Lists root-level documentation files.
8. Discovers instruction-pack definitions as policy nodes.
9. Parses repository-local Go imports without compiling or executing code.
10. Builds stable component and artefact nodes, containment and dependency
    edges, direct reverse-dependency change impact, criticality and
    trust-boundary classifications, and policy attachment points.
11. Reports evidence coverage for ownership, dependencies, data flows, trust
    boundaries, criticality, policy attachments, and change impact.
12. Writes the result to `.github/carl/repo-map.json`.

The command is idempotent — running it again updates the file in place.
`.git/`, `node_modules/`, and `vendor/` are always excluded from the scan.

**Output**

```
Repo map updated: .github/carl/repo-map.json
  Languages:     Go
  Entry points:  2
  Directories:   22
  Workflows:     1
  Governance:    8
  Documentation: 7
  Graph nodes:   74
  Graph edges:   91
```

**Generated file structure**

```json
{
  "schema_version": 1,
  "_note": "Repository map derived by `carl map`. Re-run to update after structural changes.",
  "generated_by": "carl map",
  "last_updated": "2026-06-18",
  "languages": ["Go"],
  "entry_points": [
    { "path": "go.mod", "purpose": "Go module definition: github.com/org/repo" },
    { "path": "cmd/myapp/main.go", "purpose": "myapp CLI entry point" }
  ],
  "directories": {
    ".github/carl": "cARLv2 governance artefacts and templates",
    "internal/mylib": "Implements the mylib subsystem."
  },
  "workflows": [
    { "path": ".github/workflows/release.yml", "purpose": "release workflow" }
  ],
  "governance": [
    { "path": ".github/carl/invariants.yml", "purpose": "Runtime invariants enforced by all implementation PRs" }
  ],
  "documentation": [
    { "path": "README.md", "purpose": "Repository overview and pack catalogue" }
  ],
  "graph": {
    "nodes": [
      {
        "id": "component:internal/mylib",
        "kind": "package",
        "path": "internal/mylib",
        "purpose": "Implements the mylib subsystem.",
        "criticality": "medium",
        "trust_boundary": "repository",
        "policy_attachment_point": true,
        "agent_context": "Implements the mylib subsystem.",
        "change_impact": ["component:cmd/myapp"]
      }
    ],
    "edges": [
      {
        "from": "component:cmd/myapp",
        "to": "component:internal/mylib",
        "type": "depends_on",
        "evidence": ["cmd/myapp/main.go"]
      }
    ],
    "coverage": {
      "ownership": {
        "status": "unavailable",
        "detail": "Ownership is not inferred. No owner is reported without a supported authoritative ownership source."
      },
      "dependencies": {
        "status": "partial",
        "detail": "Repository-local Go import declarations are derived statically; dependencies in other languages and dynamic dependencies are not inferred."
      }
    }
  }
}
```

**Notes**

- Directory purpose descriptions are derived from Go `// Package ...` or
  `// Command ...` doc comments, well-known path heuristics, or left blank.
- Graph node IDs and edges are sorted deterministically and contain only
  repository-relative paths.
- `contains` edges describe repository structure. `depends_on` edges describe
  observed repository-local Go imports.
- `change_impact` contains direct reverse Go-import dependants only. It is not
  transitive and does not guarantee runtime impact.
- Static imports are not treated as runtime data flow. Ownership is not
  guessed. `coverage` makes these evidence limits explicit.
- Policy nodes identify definitions and component/package nodes identify
  possible attachment points. They do not claim that a policy is active; use
  `carl trace` for active policy provenance.
- The generated file itself appears in the `governance` section on subsequent runs.
- Run `carl map` after adding new packages, workflows, or documentation to keep
  the map current.

---

### `carl pack`

Discovers, inspects, verifies, installs, selects, and composes instruction
packs. Works inside an initialised repository (merging bundled,
repository-local, registry-managed, and selected packs) and outside one
(bundled packs only).

**Usage**

```
carl pack list [--json]
carl pack show <pack-id> [--json]
carl pack select <pack-id>... [--json]
carl pack unselect <pack-id>... [--json]
carl pack profile list [--json]
carl pack profile show <profile-id> [--json]
carl pack profile activate <profile-id> [--role <role-id>] [--task <task-id>] [--json]
carl pack profile clear [--json]
carl pack registry list [--json]
carl pack registry search [<query>] [--registry <registry-id>] [--json]
carl pack install <pack-id> [--version <version>] [--registry <registry-id>] [--json]
carl pack update [<pack-id>...] [--json]
carl pack effective [--json]
```

**Subcommands**

| Subcommand | Purpose |
|---|---|
| `list` | List every discoverable pack with version, category, source, state, and description |
| `show <pack-id>` | Show full metadata for a single pack (e.g. `carl pack show core/security`) |
| `select <pack-id>...` | Add packs to the repository selection in `.github/carl/packs.json` |
| `unselect <pack-id>...` | Remove packs from the repository selection |
| `profile list` | List named policy profiles and the active profile/role/task context |
| `profile show <profile-id>` | Show one profile's base packs and role/task overlays |
| `profile activate <profile-id>` | Activate a profile with optional `--role` and `--task` overlays |
| `profile clear` | Clear the active profile and overlays; organisation/repository defaults remain active |
| `registry list` | List explicitly configured registries without fetching them |
| `registry search [<query>]` | Fetch configured registry indexes and list matching releases |
| `install <pack-id>` | Resolve, verify, and install the highest release (or exact `--version`) plus unavailable required dependencies |
| `update [<pack-id>...]` | Update named or all registry-managed packs from their recorded source registries |
| `effective` | Compute and print the effective pack set (active profile seeds + required dependencies, precedence order, overrides, conflicts) |

**Behaviour**

- Packs are identified by `<category>/<name>` IDs derived from their canonical
  path `.github/instructions/<category>/<name>.instructions.md` — never from
  filesystem enumeration order.
- Discovery merges three sources deterministically (sorted by pack ID):
  - **bundled** — packs embedded in the `carl` binary,
  - **repository-local** — packs present under `.github/instructions/` in the
    current repository (their metadata takes precedence over bundled copies),
  - **selected** — packs recorded in `.github/carl/packs.json` (written by
    `carl pack select` / `unselect`); when that file is absent, selection
    falls back to the legacy derivation from `.github/carl/runtime.json`
    (`managedArtifacts`) written by `carl init`.
- Pack metadata (version, title, description) is parsed from the
  `<!-- version: X.Y.Z -->` header, first `#` heading, and first paragraph of
  each pack file.
- Composition metadata is parsed from optional explicit comment headers in the
  first ten lines of a pack file — absent headers default to no dependencies,
  `additive` mode, priority `0`, and no overrides:
  - `<!-- requires: <pack-id>[, <pack-id>...] -->` — required dependencies,
  - `<!-- precedence-mode: additive|overridable|restrictable-only|immutable -->`,
  - `<!-- priority: <non-negative integer> -->`,
  - `<!-- overrides: <pack-id>[, <pack-id>...] -->` — explicit override
    declarations.
  Malformed headers are explicit errors; a repository-local pack's composition
  headers take precedence over the bundled copy's when it declares any.
- The metadata model is versioned: every payload carries `"schemaVersion": 1`.
- Registry-managed packs expose `"source": "registry:<registry-id>"` plus
  registry location, relative artifact, and verified SHA-256 provenance.
- When `.github/carl/profiles.json` exists, `state.active` is driven by the
  additive union of organisation defaults, repository defaults, the active
  profile, and its active role/task overlays. Every profile pack reference
  must already be selected. When `profiles.json` is absent, selected packs
  remain active as a compatibility fallback.
- The discovered pack set is validated before output: malformed IDs, duplicate
  IDs, invalid versions, unknown schema versions, missing or cyclic
  dependencies, invalid owned-artefact paths, and contradictory states are
  reported as errors.

**Output**

`carl pack list` prints a deterministic table; with `--json` it prints:

```json
{
  "schemaVersion": 1,
  "packs": [
    {
      "schemaVersion": 1,
      "id": "core/security",
      "version": "1.2.0",
      "title": "Security Pack",
      "description": "...",
      "category": "core",
      "source": "bundled+repository-local",
      "state": {
        "bundled": true,
        "installed": true,
        "selected": false,
        "active": false
      },
      "ownedArtifacts": [".github/instructions/core/security.instructions.md"],
      "dependencies": []
    }
  ]
}
```

`carl pack show <pack-id>` prints the same pack object under a `"pack"` key
with `--json`, or a human-readable detail view without it.

**Registry configuration and installation**

Registries are opt-in. There is no built-in or automatically discovered
registry. `.github/carl/registries.json` is a user-owned committed artefact:

```json
{
  "schemaVersion": 1,
  "registries": [
    {
      "id": "team",
      "location": "https://packs.example.com/carl/index.json"
    },
    {
      "id": "offline",
      "location": ".carl-registry/index.json"
    }
  ]
}
```

Remote locations must use HTTPS and cannot contain credentials, query strings,
or fragments. Local locations are repository-relative JSON paths and cannot
escape the repository. Existing commands (`list`, `show`, `select`, profiles,
and `effective`) never fetch registries. Only `registry search`, `install`,
and `update` access configured sources.

A registry index uses schema version 1:

```json
{
  "schemaVersion": 1,
  "packs": [
    {
      "id": "languages/rust",
      "version": "1.1.0",
      "artifact": "packs/languages-rust-1.1.0.instructions.md",
      "sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
      "title": "Rust Pack",
      "description": "Rust governance guidance."
    }
  ]
}
```

Artifacts are always relative to the index. A remote artifact therefore stays
on the same HTTPS origin, while a local artifact stays within the repository.
Cross-origin redirects, absolute artifacts, traversal, credential-bearing
URLs, oversized responses, malformed schemas, duplicate releases, and
invalid IDs, versions, or digests are rejected.

Resolution chooses the highest semantic version deterministically.
`--version` requests an exact version and `--registry` restricts authority to
one configured registry. If the winning version exists in multiple registries,
the command fails as ambiguous until `--registry` is supplied.

Install validates every artifact before any write:

1. fetch the selected relative artifact;
2. verify its SHA-256 against the configured index;
3. verify its declared version and composition headers;
4. resolve and verify unavailable required dependencies from the same
   registry;
5. validate the complete resulting pack set and every target path;
6. write instruction packs and provenance as one rollback-capable operation.

Registry installation never overwrites an unowned repository-local or bundled
pack. It never writes `runtime.json`, `packs.json`, or `profiles.json`, so the
pack remains unselected and inactive until explicitly selected.

Installed provenance is deterministic and committed in
`.github/carl/installed-packs.json`:

```json
{
  "schemaVersion": 1,
  "packs": [
    {
      "id": "languages/rust",
      "version": "1.1.0",
      "registry": "team",
      "registryLocation": "https://packs.example.com/carl/index.json",
      "artifact": "packs/languages-rust-1.1.0.instructions.md",
      "sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
      "installedPath": ".github/instructions/languages/rust.instructions.md"
    }
  ]
}
```

`carl pack update` uses each pack's recorded registry and refuses to overwrite
an installed file whose digest has drifted. It reports `unchanged` when no
newer version exists and rejects a registry that changes the digest of an
already recorded version. SHA-256 proves that artifact bytes match the
configured index; it does **not** prove publisher identity or provide a
signature trust chain.

`carl pack select` / `carl pack unselect` persist the repository pack
selection deterministically (deduplicated, sorted by pack ID) in
`.github/carl/packs.json`:

```json
{
  "schemaVersion": 1,
  "selected": ["core/baseline", "languages/go"]
}
```

`packs.json` is a user-owned committed artefact; pack commands never write
`.github/carl/runtime.json`. Selecting validates that every named pack is
discoverable. Unselecting a pack that other selected packs still require
leaves it in the *effective* set as a dependency (with its reason reported by
`carl pack effective`). A pack referenced by any configured profile/default/
role/task cannot be unselected until that profile reference is removed.

`profiles.json` is also user-owned and committed. Profile definitions are
edited as policy-as-code; the CLI validates and activates them:

```json
{
  "schemaVersion": 1,
  "defaults": {
    "organization": ["core/security"],
    "repository": ["core/baseline", "languages/go"]
  },
  "profiles": [
    {
      "id": "developer",
      "description": "Default implementation context.",
      "packs": ["core/carl"],
      "roles": {
        "reviewer": ["core/pr-contract"]
      },
      "tasks": {
        "security-review": ["core/identity"]
      }
    }
  ],
  "active": {
    "profile": "developer",
    "role": "reviewer",
    "task": "security-review"
  }
}
```

`carl init` installs `.github/carl/profiles.example.json` as an inactive,
cloneable reference baseline. Its `default` profile explicitly names the
complete pack set shipped with cARL and leaves the optional role/task context
unset, matching the existing role-neutral compatibility context. To adopt it,
copy it to `.github/carl/profiles.json`, ensure every referenced pack is
selected, and then edit the ordinary schema-version 1 data as required. The
example filename is never read as active profile state and has no special
evaluation semantics.

Profile, role, and task IDs use lower-case kebab-case. Defaults and overlays
compose additively; they never remove a selected pack or imply override
authority. `carl pack profile activate` and `clear` write only
`profiles.json`, deterministically. Unknown profiles, roles, tasks, duplicate
profile IDs, malformed fields, and references to unselected packs are errors.

`carl pack effective` computes the effective pack set:

1. Start from profile-driven active packs: organisation/repository defaults,
   active-profile packs, and active role/task overlays. Without
   `profiles.json`, start from selected packs for compatibility.
2. Expand required dependencies transitively; every entry carries explicit
   reasons (`organization default`, `profile <id>`,
   `role <id> in profile <id>`, `task <id> in profile <id>`,
   `selected` for legacy fallback, or `dependency of <id>`).
3. Apply explicit override declarations: an override is honoured only when
   the overriding pack declares it in metadata **and** the target pack
   declares `precedence-mode: overridable`. Overridden packs remain in the
   evaluation, flagged with `overriddenBy` for provenance, but their
   instruction definitions are not applied.
4. Order by precedence: priority descending, ties broken by pack ID — never
   filesystem or load order.

Conflicts (missing dependencies, overriding a non-overridable pack, mutual
overrides) are reported and cause a non-zero exit. With `--json`:

```json
{
  "schemaVersion": 1,
  "packs": [
    {
      "id": "languages/go",
      "version": "1.0.0",
      "priority": 0,
      "mode": "additive",
      "reasons": ["selected"]
    },
    {
      "id": "core/baseline",
      "version": "1.2.0",
      "priority": 0,
      "mode": "additive",
      "reasons": ["dependency of languages/go"]
    }
  ],
  "conflicts": [
    { "code": "override_not_permitted", "message": "..." }
  ]
}
```

**Errors**

| Error | Cause | Resolution |
|---|---|---|
| `unknown pack "<id>"` | The pack ID does not match any discoverable pack | Run `carl pack list` to see valid IDs |
| `pack validation failed: ...` | Discovered pack set contains invalid metadata | Fix the reported pack file or manifest entry |
| `parse .github/carl/packs.json: ...` | Selection artefact is malformed | Fix or delete `.github/carl/packs.json` |
| `read legacy pack selection from .github/carl/runtime.json: ...` | `packs.json` is absent and the legacy manifest is malformed | Repair the manifest or create an explicit valid `packs.json`; do not infer selection from directory contents |
| `.github/carl/profiles.json validation failed: ...` | Profile schema, context, or references are invalid | Fix the reported profile/default/overlay entry |
| `unknown profile "<id>"` | The requested profile is not configured | Run `carl pack profile list` |
| `.github/carl/registries.json does not configure any registries` | Registry search/install was requested without an explicit source | Define a trusted HTTPS or repository-local registry |
| `... is ambiguous across registries ...` | The same winning pack version is advertised by multiple authorities | Repeat with `--registry <id>` |
| `SHA-256 mismatch` | Artifact bytes do not match the configured index | Do not install; investigate the registry or transport |
| `... already exists as an unowned repository-local pack` | Install would overwrite a file not owned by registry provenance | Choose another pack ID or explicitly reconcile ownership |
| `installed pack ... has drifted` | A registry-managed file changed after installation | Review the local changes before updating |
| `pack composition conflicts detected` | `carl pack effective` found override or dependency conflicts | Fix the conflicting pack metadata or selection |

With `--json`, errors are emitted as a structured payload on stderr with a
non-zero exit code:

```json
{
  "schemaVersion": 1,
  "error": { "code": "pack_not_found", "message": "unknown pack \"x/y\"" }
}
```

**Notes**

- `state.selected` identifies the repository's eligible policy packs;
  `state.active` identifies profile-driven active seeds. Required dependencies
  appear in the effective set even when they are not active seeds.
- Pack *selection* (which packs are in play) is distinct from *priority*
  (ordering among selected packs) and *override authority* (whether a pack may
  relax another pack's rules). All three are modelled: selection lives in
  `packs.json`, priority and override authority come only from explicit pack
  metadata headers — never from load order.
- Composition is conservative: non-overridden effective packs add
  constraints; overridden packs remain visible for provenance but are not
  applied, and no override authority is inferred.
- Registry integrity and policy activation are separate: installing a verified
  artifact records provenance but does not select or activate it.

---

### `carl explain` / `carl trace`

Reports pack-level policy provenance and the observable decisions made by the
existing effective-pack evaluator.

**Usage**

```
carl explain <pack-id> [--json]
carl trace [--json]
```

`carl explain` works for every discoverable pack, including an inactive one.
It reports:

- whether the pack is applied, effective, overridden, or inactive; overridden
  entries report `applied: false` and `addsConstraints: false`;
- its version, source, and repository-relative canonical definition;
- registry provenance when it is registry-managed;
- whether it is selected or an active seed;
- structured selection, organisation/repository default, profile, role, task,
  and dependency activation steps with their source artefact;
- effective order, priority, and precedence mode;
- required and requiring packs;
- whether it adds constraints, declares or resolves overrides, or is
  overridden by another effective pack.

`carl trace` reports the complete effective policy evaluation in the same
precedence order as `carl pack effective`, including overridden entries for
provenance, followed by structured decisions:

- active seed and dependency inclusion;
- precedence ordering (priority descending, then pack-ID tie-break);
- conservative pack-level constraint strengthening;
- explicit non-application of overridden definitions;
- permitted overrides, including why they resolve (explicit declaration plus
  an `overridable` target);
- unresolved composition conflicts.

Both commands are read-only, local-only, and network-free. They reuse strict
pack, selection, profile, and installed-provenance validation and never fetch
configured registries. An unresolved composition conflict is printed and
causes a non-zero exit in human and JSON modes.

JSON output uses schema version 1. A shortened
`carl explain core/baseline --json` example:

```json
{
  "schemaVersion": 1,
  "notice": "Pack-level policy provenance only. ...",
  "context": {
    "mode": "profiles",
    "source": ".github/carl/profiles.json",
    "profile": "developer"
  },
  "policy": {
    "id": "core/baseline",
    "version": "1.2.0",
    "applied": true,
    "status": "effective",
    "source": "bundled+repository-local",
    "canonicalDefinition": ".github/instructions/core/baseline.instructions.md",
    "order": 1,
    "activation": [
      {
        "kind": "dependency",
        "description": "dependency of languages/go",
        "source": ".github/instructions/languages/go.instructions.md",
        "relatedPack": "languages/go"
      }
    ],
    "effect": {
      "addsConstraints": true,
      "declaredOverrides": [],
      "resolvedOverrides": [],
      "overriddenBy": []
    }
  },
  "decisions": [],
  "conflicts": []
}
```

Every human and JSON response includes an explicit boundary notice:
explanation is limited to pack-level policy provenance. The commands do not
interpret individual natural-language rules and do not expose prompts, hidden
model reasoning, or chain-of-thought. Instruction availability or loading does
not prove model adherence.

With `--json`, unknown packs use the existing `pack_not_found` structured
error. Invalid repository policy inputs use `policy_evaluation_failed`.

---

### `carl reconcile`

Updates repository-specific memory sections in `.github/carl/memory.md` using
data from `.github/carl/repo-map.json`. Human-authored content is preserved;
only the generated snapshot section is updated.

**Usage**

```
carl reconcile
```

**What it does**

1. Reads `.github/carl/repo-map.json` (generated by `carl map`).
2. Reads `.github/carl/memory.md`.
3. Builds a repository snapshot from the repo-map data: languages, entry
   points, key directories, workflows, governance artefacts, documentation
   files, and a last-reconciled date.
4. Updates the generated section in `memory.md` (delimited by
   `<!-- BEGIN GENERATED: reconcile -->` and `<!-- END GENERATED: reconcile -->`),
   leaving all human-authored content outside those markers untouched.
5. If the generated content is identical to what is already in `memory.md`,
   no file is written.

The command does not modify `runtime.json`, harness adapter files, or any
other managed artefact.

**Output (no changes needed)**

```
No reconciliation needed.
```

**Output (changes made)**

```
Reconciled durable artefacts.
  .github/carl/memory.md
```

**Errors**

| Error | Cause | Resolution |
|---|---|---|
| `repo map not found` | `.github/carl/repo-map.json` does not exist | Run `carl map` first |
| `memory.md not found` | `.github/carl/memory.md` does not exist | Run `carl init` first |

**Notes**

- Run `carl map` before `carl reconcile` to ensure the repo-map reflects the
  current repository structure.
- `carl reconcile` is idempotent — running it twice on the same repo-map
  produces the same `memory.md`.
- The generated section uses HTML comment markers so it is invisible when
  rendered as Markdown.
- Human-authored sections (project purpose, architecture, invariants, field
  findings, etc.) are never overwritten.

---

### `carl convert`

Migrates durable governance knowledge from a legacy or foreign source into
canonical cARL artefacts. The first supported source is **AADLC**
(`carl convert aadlc`). cARL is the productised form of AADLC, so many
repositories already carry durable invariants, lessons, and governance rules
under AADLC paths that should be preserved when cARL is adopted.

The command is built around a converter framework: each source implements a
small `Converter` interface (discover + classify) while a shared,
converter-agnostic engine performs duplicate detection, conflict detection,
routing, and reporting. Additional converters (e.g. `claude`, `copilot`,
`repo`) can be added later without reworking the migration engine.

**Usage**

```
carl convert <source> [--dry-run | --apply]
```

Sources:

| Source | Description |
|---|---|
| `aadlc` | Migrate durable knowledge from legacy AADLC artefacts |

Flags:

| Flag | Description |
|---|---|
| `--dry-run` | Analyse and report migration opportunities without writing (default) |
| `--apply` | Perform the migration and update cARL artefacts |

`--dry-run` and `--apply` are mutually exclusive. With no flag, the command
defaults to `--dry-run`.

**Discovery**

`carl convert aadlc` searches conventional AADLC locations:

```
.aadlc/
.github/aadlc/
aadlc/
AADLC.md
```

Directories are scanned recursively; only Markdown (`.md`) and YAML
(`.yml`/`.yaml`) files are considered. Missing locations are skipped silently.

**Classification**

Discovered content is classified into three categories, each routed to a
canonical cARL destination:

| Category | Examples | Destination |
|---|---|---|
| Invariants | Repository constraints and assumptions | `.github/carl/invariants.yml` |
| Durable memory | Architectural decisions, lessons learned, known limitations, historical context | `.github/carl/memory.md` |
| Governance rules | PR contract, planning, and approval requirements | `.github/carl/memory.md` |

YAML files using the cARL `invariants:` schema contribute their rules as
invariants. Markdown files are scanned section by section: a heading's text
selects the category, and the bullet-list items beneath it become migration
items. Content that cannot be confidently classified is ignored — the command
prefers safety over speculative migration.

**Conflict handling**

Existing cARL knowledge is never overwritten:

- **Duplicate** — an item already present in the destination is skipped and
  reported.
- **Conflict** — an item that collides with, but differs from, existing cARL
  content (e.g. a migrated invariant whose generated id matches an existing
  invariant with different wording) is reported and left for human review;
  it is never written.

**Migration report**

Both modes print the same deterministic report (only the destination heading
and the trailing note differ):

```
AADLC Migration Report

Discovered:
  2 artefact(s)
    .aadlc/invariants.yml
    AADLC.md

Convertible:
  3 invariant(s)
  2 memory entry(ies)
  2 governance rule(s)

Skipped:
  1 duplicate(s)

Conflicts:
  0 item(s) requiring review

Updated:
  .github/carl/invariants.yml
  .github/carl/memory.md

Migration applied.
```

Under `--dry-run` the destination heading reads `Would update:` and the report
ends with `Dry run — no changes written. Re-run with --apply to migrate.`

**Errors**

| Error | Cause | Resolution |
|---|---|---|
| `unknown convert source "<id>"` | The source is not registered | Run `carl convert --help` to see valid sources |
| `--apply and --dry-run are mutually exclusive` | Both flags were passed | Pass at most one mode flag |
| `invariants.yml not found` / `memory.md not found` | A destination is missing while applying | Run `carl init` first |
| `the convert-generated section markers are malformed` | `memory.md` has a begin marker without an end marker, an end marker without a begin marker, or an end marker before the begin marker | Repair the markers manually or restore from a clean `memory.md` before re-running |

**Notes**

- AADLC artefacts are never deleted or modified — the command only reads them.
- Migrated invariants are namespaced with an `aadlc-` id prefix and a derived
  name; severity defaults to `high`.
- Migrated memory and governance entries live in a managed block in
  `memory.md` delimited by `<!-- BEGIN GENERATED: convert aadlc -->` /
  `<!-- END GENERATED: convert aadlc -->`. Human-authored content outside the
  block is preserved. If these markers are malformed (begin without end, end
  without begin, or end before begin), the command fails with a non-zero exit
  and writes nothing rather than appending a second generated block.
- The command is idempotent and produces deterministic output — running it
  repeatedly never duplicates content.

---

### `carl plan`

Discovers, validates, and summarises plan files in `.github/carl/plans/`.

**Usage**

```
carl plan
```

**What it does**

1. Scans `.github/carl/plans/` for `.md` files.
2. Parses each file to extract:
   - **Title** — the first level-1 heading (`# …`).
   - **Status** (lifecycle state) — the `Status:` field in the `## Plan metadata`
     list item (e.g. Draft, Active, Completed, Archived).
   - **Purpose** — the first paragraph of `## Task summary`, `## Task`, or `## Goal`
     (tried in that order).
3. Validates each plan against the standard plan template structure:
   - Missing `## Plan metadata` section → Warning.
   - Empty `Status:` in `## Plan metadata` → Warning.
4. Prints a summary for each plan and a total count.
5. Exits with code `0` regardless of validation warnings — the command is
   read-only and never modifies any files.

**Output (no plans directory or no .md files)**

```
No plans found.
```

**Output (plans found)**

```
Plans in .github/carl/plans/

  my-feature.md
    Title:    My Feature Plan
    Status:   Active
    Purpose:  Add the widget subsystem.

  draft.md
    Title:    Draft Plan
    Status:   (not set)
    Purpose:  Starting soon.
    Warning:  Status not set in ## Plan metadata

2 plan(s) found. 1 warning(s).
```

**Fields**

| Field | Source | Notes |
|---|---|---|
| Title | First `# heading` in the file | Falls back to `(not set)` |
| Status | `- Status:` in `## Plan metadata` | Lifecycle state: Draft, Active, Completed, Archived |
| Purpose | First paragraph of `## Task summary`, `## Task`, or `## Goal` | Falls back to `(not set)` |

---

### `carl harness`

Manages and inspects harness adapters for AI coding agents.

Harness adapters bridge cARL canonical artefacts to the context injection
mechanisms of specific AI coding agents. cARL artefacts (`.github/carl/`) are
the canonical source of truth; harness files are adapters, not authorities.

**Usage**

```
carl harness <subcommand> [arguments]
```

**Subcommands**

| Subcommand | Description |
|---|---|
| `list` | List known harness adapters and their support status |
| `status` | Report harness adapter presence and sync health in the current repository |
| `sync` | Generate adapter files for all harnesses from canonical cARL artefacts |

Run `carl harness --help` to print subcommand usage.

---

### `carl harness list`

Lists all known harness adapters and their support status.

**Usage**

```
carl harness list
```

**What it does**

1. Prints the canonical adapter registry — all harnesses cARL knows about.
2. For each adapter shows: ID, display name, and support status (`production`, `experimental`, or `theoretical`).
3. Prints a summary line with counts by tier.

This subcommand is purely informational — it does not check the filesystem.

**Output**

```
Harness Adapters:

  copilot       GitHub Copilot       production
  claude        Claude Code          production
  codex         Codex                production
  cursor        Cursor               theoretical
  antigravity   Antigravity          theoretical

3 production, 0 experimental, 2 theoretical (5 total).
```

**Support status values**

| Status | Meaning |
|---|---|
| `production` | Tested and validated end-to-end in the native harness |
| `experimental` | Partial validation; governance loading under investigation |
| `theoretical` | Adapter is implemented; not yet validated end-to-end in the native harness |

> **Note:** Content generation and sync are available for all five adapters.
> Copilot, Claude Code, and Codex are proven production harnesses. Cursor and
> Antigravity have implemented shims but have not yet been tested in their
> native harnesses.

---

### `carl harness status`

Reports the detection and sync status of all known harness adapters in the current repository.

**Usage**

```
carl harness status
```

**What it does**

1. For each known adapter, checks whether its detection file is present in the repository.
2. For adapters with defined adapter files, compares adapter file bytes against the canonical embedded source.
3. Reports presence as `Present` or `Missing`, and sync health as `Synced`, `Drifted`, `Missing`, or `-`.
4. Prints a summary line with detected, missing, drifted, and healthy adapter counts.

**Output (Copilot synced)**

```
Harness Adapter Status:

  copilot       GitHub Copilot       production    Present  Synced
  claude        Claude Code          production    Missing  Missing
  codex         Codex                production    Missing  Missing
  cursor        Cursor               theoretical   Missing  Missing
  antigravity   Antigravity          theoretical   Missing  Missing

1 detected, 4 missing, 0 drifted, 1 healthy.
```

**Output (drifted adapter)**

```
Harness Adapter Status:

  copilot       GitHub Copilot       production    Present  Synced
  claude        Claude Code          production    Present  Drifted
  codex         Codex                production    Missing  Missing
  cursor        Cursor               theoretical   Missing  Missing
  antigravity   Antigravity          theoretical   Missing  Missing

2 detected, 3 missing, 1 drifted, 1 healthy.
```

**Presence and sync values**

| Status | Meaning |
|---|---|
| `Present` | Detection file is present; this does not prove governance activation |
| `Missing` | Detection file or managed adapter file is absent |
| `Drifted` | Adapter file exists but differs from the canonical embedded source |
| `Synced` | Managed adapter files exist and match their canonical embedded sources; this does not prove activation |
| `-` | No presence or sync check is available for this adapter |

**Detection file by adapter**

| Adapter | Detection file |
|---|---|
| `copilot` | `.github/copilot-instructions.md` |
| `claude` | `CLAUDE.md` |
| `codex` | `AGENTS.md` |
| `cursor` | `.cursor/rules/carl.mdc` |
| `antigravity` | `.agents/rules/carl.md` |

---

### `carl harness sync`

Generates adapter files for all harnesses from the canonical cARL artefacts
embedded in the CLI binary. Adapter files are disposable — they are always
regenerated from the canonical source and should not be edited manually.

**Usage**

```
carl harness sync [<harness-id>...]
```

**What it does**

1. Resolves the set of target harnesses: all harnesses with defined adapter files if no IDs are
   given, or only the named harnesses if one or more IDs are provided.
2. For each target harness, writes all required adapter files (shared loader plus harness-specific
   shim) from embedded artefacts. The shared loader (`.github/copilot-instructions.md`) is written
   once even when syncing multiple harnesses.
3. Creates parent directories as needed. Existing files are overwritten.
4. Reports each file written and a summary count.

**Adapter model**

`.github/copilot-instructions.md` is the shared cARL adapter loader. Every harness sync writes this
file. The harness-specific shim files (CLAUDE.md, AGENTS.md, etc.) are tiny files that tell the
harness to read `.github/copilot-instructions.md` before any repository work. Canonical governance
remains under `.github/carl/`.

**Output (sync all harnesses)**

```
Syncing harness adapters...

  copilot        .github/copilot-instructions.md
  claude         CLAUDE.md
  codex          AGENTS.md
  cursor         .cursor/rules/carl.mdc
  antigravity    .agents/rules/carl.md

5 adapter file(s) synced.
```

**Output (sync a specific harness)**

```
Syncing harness adapters...

  .github/copilot-instructions.md
  claude         CLAUDE.md

2 adapter file(s) synced.
```

**Errors**

| Error | Cause | Resolution |
|---|---|---|
| `unknown harness "<id>"` | The given harness ID is not in the registry | Run `carl harness list` to see valid IDs |

**Notes**

- Sync is idempotent — running it multiple times produces the same result.
- The command does not require `carl init` to have been run first.
- To activate a harness after sync, simply commit the generated adapter files.

---

## carl version

Shows three distinct version layers:

1. **CLI version** — the `carl` executable version.
2. **Bundled runtime version** — the canonical cARL governance payload embedded
   in the executable (with source/tag/commit provenance).
3. **Repository runtime version** — the runtime installed in the current
   repository (`.github/carl/runtime.json`), when present.

**Usage**

```
carl version [--components]
```

Aliases: `carl --version`, `carl -v`

**What it does**

1. Prints CLI and bundled runtime metadata (always, even outside an initialised repository).
2. Reads repository runtime metadata from `runtime.json` only when it exists.
3. Compares repository runtime version against bundled runtime version:
   - `Current`
   - `Upgrade available`
   - `Repository runtime is newer`
   - `Unknown` (non-semver comparison)
4. When runtime is installed, prints installed instruction packs and installed versions
   derived from each installed pack file metadata header: `<!-- version: X.Y.Z -->`.
5. Prints harness support tiers and shim versions from installed detection files.
6. With `--components`, prints support tiers plus bundled vs installed component
   versions and drift state.

**Output (runtime not installed)**

```
cARL CLI:
  Version:          1.2.0
Bundled Runtime:
  Version:          1.1.0
  Source:           goldjg/cARL
  Tag:              v1.2.0
  Commit:           98f680b3...
Repository Runtime:
  Not installed in the current repository.
Harness Shims:
  Harness       Support      File                                Version
  copilot       production   .github/copilot-instructions.md     2.1.1
  claude        production   CLAUDE.md                           unknown
  codex         production   AGENTS.md                           not installed
  cursor        theoretical  .cursor/rules/carl.mdc              not installed
  antigravity   theoretical  .agents/rules/carl.md               not installed
```

**Output (runtime installed)**

```
cARL CLI:
  Version:          1.2.0
Bundled Runtime:
  Version:          1.1.0
  Source:           goldjg/cARL
  Tag:              v1.2.0
  Commit:           98f680b3...
Repository Runtime:
  Version:          1.0.0
  Source:           goldjg/cARL
  Tag:              v1.0.0
  Commit:           742ac661...
  Status:           Upgrade available

Installed Packs:
  cloud/azure                       1.0.1
  core/baseline                     1.1.0
  core/carl                         2.0.0

Harness Shims:
  Harness       Support      File                                Version
  copilot       production   .github/copilot-instructions.md     2.1.1
  claude        production   CLAUDE.md                           1.0.0
  codex         production   AGENTS.md                           unknown
  cursor        theoretical  .cursor/rules/carl.mdc              not installed
  antigravity   theoretical  .agents/rules/carl.md               not installed
```

**`--components` output**

```
Instruction Packs:
  Pack                              Bundled   Installed  State
  core/baseline                     1.1.0     1.0.0      older
  core/carl                         2.0.0     2.0.0      current
  cloud/azure                       1.0.1     missing    missing

Harness Shims:
  Harness       Support      File                              Bundled   Installed  State
  copilot       production   .github/copilot-instructions.md   2.1.1     1.0.0      older
  claude        production   CLAUDE.md                         unknown   unknown    unknown
  codex         production   AGENTS.md                         unknown   missing    missing
```

**Repository Runtime status values**

| Status | Meaning |
|---|---|
| `Current` | Repository runtime version equals bundled runtime version |
| `Upgrade available` | Bundled runtime version is newer than repository runtime version |
| `Repository runtime is newer` | Repository runtime version is newer than bundled runtime version |
| `Unknown` | Either bundled or repository runtime version is not valid semantic versioning |

---

## Global options

| Flag | Alias | Effect |
|---|---|---|
| `--help` | `-h` | Print usage and available commands, then exit |
| `--version` | `-v` | Alias for `carl version` |

Run `carl` with no arguments to print usage.

---

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Command completed successfully |
| `1` | An error occurred; details are printed to stderr |

---

## Runtime manifest

`carl init` writes `.github/carl/runtime.json` — a JSON file that is the
authoritative source of truth for the installed runtime state.

```json
{
  "runtimeVersion": "1.0.0",
  "source": "goldjg/cARL",
  "sourceTag": "v1.0.0",
  "sourceCommit": "abc1234...",
  "installedAt": "2025-01-01T00:00:00Z",
  "managedArtifacts": [
    ".github/carl/invariants.yml",
    ".github/copilot-instructions.md",
    "..."
  ]
}
```

This file must not be edited manually and is never overwritten by `carl repair`.
