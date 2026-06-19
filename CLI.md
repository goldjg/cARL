<!-- version: 1.0.0 -->
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
carl init
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
| `cARL artefacts already exist` | Individual managed files exist without a `runtime.json` | Remove the listed files or adopt them before running `init` |

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
| `no cARL runtime installed` | `runtime.json` not found | Run `carl init` first |

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
   - **INFO** — no issues found; runtime is healthy
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
```

**Output (missing and drifted artefacts)**

```
ERROR   .github/carl/invariants.yml — artefact is missing from disk
        Action: run `carl repair`
WARNING .github/copilot-instructions.md — artefact has drifted from its canonical version
        Action: run `carl repair`
WARNING claude (CLAUDE.md) — harness adapter file has drifted from its canonical version
        Action: run `carl harness sync`

1 error(s), 2 warning(s), 0 info(s) found.
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
4. Inspects harness adapters and reports a separate summary of active, missing,
   drifted, and healthy adapters.
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
  Active adapters:  5
  Missing adapters: 0
  Drifted adapters: 0
  Healthy adapters: 5

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
  Active adapters:  4
  Missing adapters: 1
  Drifted adapters: 1
  Healthy adapters: 3

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
structure from the filesystem.

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
8. Writes the result to `.github/carl/repo-map.json`.

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
```

**Generated file structure**

```json
{
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
  ]
}
```

**Notes**

- Directory purpose descriptions are derived from Go `// Package ...` or
  `// Command ...` doc comments, well-known path heuristics, or left blank.
- The generated file itself appears in the `governance` section on subsequent runs.
- Run `carl map` after adding new packages, workflows, or documentation to keep
  the map current.

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

**Notes**

- AADLC artefacts are never deleted or modified — the command only reads them.
- Migrated invariants are namespaced with an `aadlc-` id prefix and a derived
  name; severity defaults to `high`.
- Migrated memory and governance entries live in a managed block in
  `memory.md` delimited by `<!-- BEGIN GENERATED: convert aadlc -->` /
  `<!-- END GENERATED: convert aadlc -->`. Human-authored content outside the
  block is preserved.
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
| `sync` | Generate adapter files for supported harnesses from canonical cARL artefacts |

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
2. For each adapter shows: ID, display name, and support status (`supported` or `planned`).
3. Prints a summary line with the count of supported adapters.

This subcommand is purely informational — it does not check the filesystem.

**Output**

```
Harness Adapters:

  copilot       GitHub Copilot       supported
  claude        Claude Code          supported
  codex         Codex                supported
  cursor        Cursor               supported
  antigravity   Antigravity          supported

5 of 5 adapter(s) supported.
```

**Support status values**

| Status | Meaning |
|---|---|
| `supported` | Detection file and adapter files are defined; detection and status reporting are active |
| `planned` | Adapter is declared for discoverability; not yet implemented |

> **Note:** Content generation and sync (populating adapter files from cARL artefacts) is now available for all supported adapters via `carl harness sync`.

---

### `carl harness status`

Reports the detection and sync status of all known harness adapters in the current repository.

**Usage**

```
carl harness status
```

**What it does**

1. For each known adapter, checks whether its detection file is present in the repository.
2. For supported adapters, compares adapter file bytes against the canonical embedded source.
3. Reports presence as `Present` or `Missing`, and sync health as `Synced`, `Drifted`, `Missing`, or `-`.
4. Prints a summary line with active, missing, drifted, and healthy adapter counts.

**Output (Copilot synced)**

```
Harness Adapter Status:

  copilot       GitHub Copilot       supported    Present  Synced
  claude        Claude Code          supported    Missing  Missing
  codex         Codex                supported    Missing  Missing
  cursor        Cursor               supported    Missing  Missing
  antigravity   Antigravity          supported    Missing  Missing

1 active, 4 missing, 0 drifted, 1 healthy.
```

**Output (drifted adapter)**

```
Harness Adapter Status:

  copilot       GitHub Copilot       supported    Present  Synced
  claude        Claude Code          supported    Present  Drifted
  codex         Codex                supported    Missing  Missing
  cursor        Cursor               supported    Missing  Missing
  antigravity   Antigravity          supported    Missing  Missing

2 active, 3 missing, 1 drifted, 1 healthy.
```

**Presence and sync values**

| Status | Meaning |
|---|---|
| `Present` | Detection file is present; harness is active in the repository |
| `Missing` | Detection file or managed adapter file is absent |
| `Drifted` | Adapter file exists but differs from the canonical embedded source |
| `Synced` | Adapter file exists and matches the canonical embedded source |
| `-` | No presence or sync check is available for this adapter |

**Detection file by adapter**

| Adapter | Detection file |
|---|---|
| `copilot` | `.github/copilot-instructions.md` |
| `claude` | `CLAUDE.md` |
| `codex` | `AGENTS.md` |
| `cursor` | `.cursorrules` |
| `antigravity` | `ANTIGRAVITY.md` |

---

### `carl harness sync`

Generates adapter files for supported harnesses from the canonical cARL artefacts
embedded in the CLI binary. Adapter files are disposable — they are always
regenerated from the canonical source and should not be edited manually.

**Usage**

```
carl harness sync [<harness-id>...]
```

**What it does**

1. Resolves the set of target harnesses: all supported harnesses if no IDs are
   given, or only the named harnesses if one or more IDs are provided.
2. For each target harness, reads the canonical content from the embedded
   artefacts (`.github/copilot-instructions.md`).
3. Writes the content to each harness's adapter file(s), creating parent
   directories as needed. Existing files are overwritten.
4. Reports each file written and a summary count.

**Output (sync all harnesses)**

```
Syncing harness adapters...

  copilot        .github/copilot-instructions.md
  claude         CLAUDE.md
  codex          AGENTS.md
  cursor         .cursorrules
  antigravity    ANTIGRAVITY.md

5 adapter file(s) synced.
```

**Output (sync a specific harness)**

```
Syncing harness adapters...

  claude         CLAUDE.md

1 adapter file(s) synced.
```

**Errors**

| Error | Cause | Resolution |
|---|---|---|
| `unknown harness "<id>"` | The given harness ID is not in the registry | Run `carl harness list` to see valid IDs |

**Notes**

- Sync is idempotent — running it multiple times produces the same result.
- The command does not require `carl init` to have been run first.
- To activate a harness after sync, simply commit the generated adapter file.

---

## carl version

Shows CLI and installed runtime version information, including installed packs
and a health status.

**Usage**

```
carl version
```

Aliases: `carl --version`, `carl -v`

**What it does**

1. Reads `runtime.json` for the installed runtime version, source, and artefact list.
2. Derives installed pack names from the managed artefact paths.
3. Performs a content-based health check: compares each managed artefact against
   its embedded canonical version and reports whether the runtime is healthy or
   has drifted.

**Output (runtime installed)**

```
CLI Version:      1.0.0
Runtime Version:  1.0.0
Source:           goldjg/cARL
Tag:              v1.0.0
Commit:           abc1234

Installed Packs:
  cloud/azure
  cloud/entra
  cloud/gcp
  cloud/microsoft-graph
  cloud/netlify
  core/baseline
  core/carl
  core/cognition-governance
  core/dependency
  core/identity
  core/memory-cache
  core/pr-contract
  core/security
  core/tool-permission-tiers
  languages/html
  languages/javascript
  languages/powershell
  languages/python
  languages/terraform
  languages/typescript
  platform/cicd
  platform/docker
  platform/kubernetes

Runtime Status:
  Healthy
```

**Output (no runtime installed)**

```
No cARL runtime installed.
```

**Runtime Status values**

| Status | Meaning |
|---|---|
| `Healthy` | All managed artefacts are present and match their canonical versions |
| `Drift detected (N artefact(s) modified or missing)` | One or more artefacts differ; run `carl repair` |

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
