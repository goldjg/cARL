<!-- version: 1.0.0 -->
# cARL — Roadmap

This roadmap describes future evolution ideas. None of these items are implemented in the current PR. They are recorded here to preserve intent and prevent rediscovery.

---

## Guiding Principles for Roadmap Items

- **Preserve runtime semantics** — new capabilities should extend, not change, existing governance behaviour
- **One concern per pack** — new instruction packs should remain focused
- **Version-controlled artefacts** — new governance artefacts belong in `.github/carl/`
- **Backward compatibility first** — existing users should not need to change their setup

---

## Delivered

### cARL CLI v1 Foundation
**Status:** Delivered (PR #2)
**Commands:** `carl init`, `carl repair`, `carl version`
**Description:** Self-contained Go binary that manages repository-local cARL runtime
installations. All 32 artefacts are embedded in the binary (no network required).
`runtime.json` is the authoritative runtime state. `memory.md` and `runtime.json`
are protected from repair. Health status is content-based (byte-comparison against
embedded canonicals). Build-time version and commit injection via `-ldflags`.

### Release Workflow (CLI Binary Publishing)
**Status:** Delivered (PR #3)
**Workflow:** `.github/workflows/release.yml`
**Description:** GitHub Actions workflow triggered on `v*` semantic version tags.
Builds the cARL CLI for five platforms (linux/amd64, linux/arm64, darwin/amd64,
darwin/arm64, windows/amd64) using a matrix strategy. Injects the tag as
`cliVersion` and the commit SHA as `sourceCommit` via `-ldflags`. Uploads
per-platform binaries as GitHub Actions artifacts and attaches them to the
GitHub Release (creating the release if absent, or uploading to an existing one).
No new secrets or dependencies required — uses `GITHUB_TOKEN` with `contents: write`.

### `carl status` Command
**Status:** Delivered (PR #4)
**Command:** `carl status`
**Description:** Read-only health report command. Reads `runtime.json`, compares
managed repairable artefacts against embedded canonical versions, and outputs CLI
version, runtime version, source, tag, commit, installed packs, separate lists of
missing and drifted artefacts, and an overall status of Healthy, Drifted, or
Incomplete. `memory.md` and `runtime.json` are protected and never reported as
drift. Exports `repair.Inspect` for shared, tested drift classification.

### `carl doctor` Command
**Status:** Delivered (PR #5)
**Command:** `carl doctor`
**Description:** Diagnostic command. Reads `runtime.json`, inspects all managed
artefacts using `repair.Inspect`, and emits categorised findings (ERROR, WARNING,
INFO) with per-finding remediation actions. Missing artefacts produce an ERROR with
`carl repair` as the action; drifted artefacts produce a WARNING with `carl repair`;
missing manifest produces an ERROR with `carl init`. Always returns exit code 0 —
diagnostics complete successfully even when issues are present. Never modifies files.

### cARL Pack for Go
**Status:** Delivered (PR #6)
**Pack:** `.github/instructions/languages/go.instructions.md`
**Description:** Go-specific instruction pack following the same pattern as existing
language packs. Covers: error handling discipline, context propagation, goroutine
safety, standard-library preference, dependency hygiene with `go mod`, type safety
with interfaces, security (exec, path traversal, SSRF, template injection), and
testing with `go test`. Embedded in the binary under
`embedded/assets/.github/instructions/languages/go.instructions.md`.

### `carl map` Command
**Status:** Delivered (PR #7)
**Command:** `carl map`
**Description:** CLI command that derives a cognitive repository map from the
filesystem and writes `.github/carl/repo-map.json`. Detects programming languages
from source file extensions; identifies project entry points (`go.mod`,
`cmd/*/main.go`, `Makefile`, `package.json`, etc.); maps key directories (up to
3 levels deep) with purpose descriptions derived from Go package/command doc
comments or known-path heuristics (`.github/**`, common Go package paths);
lists GitHub Actions workflows, governance artefacts under `.github/carl/`, and
root-level documentation. Idempotent — re-running updates the file in place.
Excludes `.git/`, `node_modules/`, and `vendor/` from all scans.

### `carl plan` Command
**Status:** Delivered (PR #8)
**Command:** `carl plan`
**Description:** Read-only CLI command that discovers, validates, and summarises
plan files in `.github/carl/plans/`. For each `.md` file it extracts title (from
the first level-1 heading), status/lifecycle state (from the `Status:` field in
`## Plan metadata`), and purpose (from the first paragraph of `## Task summary`,
`## Task`, or `## Goal`, in that order). Validates each plan against the standard
template structure and emits inline warnings for: missing `## Plan metadata` section
and empty `Status:` field. Always exits 0 — read-only, never modifies files.

### Harness Adapter Support
**Status:** Delivered (PR #9)
**Commands:** `carl harness`, `carl harness list`, `carl harness status`
**Description:** Introduces the harness adapter concept: a bridge between cARL
canonical artefacts and AI coding agent context injection mechanisms. cARL artefacts
are the canonical source of truth; harness files are adapters, not authorities.
Supports GitHub Copilot as the first implemented adapter (detection via
`.github/copilot-instructions.md`). Registers Claude Code, Codex, Cursor, and
Antigravity as planned adapters for future implementation. `carl harness list` shows
all known adapters with support status (static, no filesystem check). `carl harness
status` detects which harnesses are active in the current repository. Both subcommands
are read-only and always exit 0. Designed for extensibility as new agents are supported.

### `carl harness sync` — Harness Adapter File Generation
**Status:** Delivered (PR #11)
**Command:** `carl harness sync [<harness-id>...]`
**Description:** Adds a `sync` subcommand to `carl harness` that generates adapter
files for all supported harnesses (or a named subset) from the canonical cARL
artefacts embedded in the CLI binary. Adapter files are treated as disposable
outputs — always regenerated from the canonical source and never edited manually.
All five harnesses (copilot, claude, codex, cursor, antigravity) are supported.
The `SourceFile` field is added to the `Adapter` struct to record which embedded
artefact provides the content for each harness. The `harness.Command` now accepts
an `Artifacts` dependency consistent with other write commands (`repair`, `doctor`,
`status`). Sync is idempotent and does not require `carl init` to have been run first.

### Harness Health Awareness
**Status:** Delivered
**Commands:** `carl harness status`, `carl doctor`, `carl status`
**Description:** Promotes harness adapters to managed disposable artefacts with
content-based health checks. `carl harness status` now reports detection-file
presence plus sync health (`Present`, `Missing`, `Drifted`, `Synced`) by
comparing adapter bytes against the embedded canonical source. `carl doctor`
surfaces missing or drifted harness adapters as `WARNING` findings with
`carl harness sync` remediation. `carl status` adds a dedicated harness summary
section (active, missing, drifted, healthy) without changing overall runtime
status semantics.

### `carl convert` Command (AADLC Migration)
**Status:** Delivered
**Command:** `carl convert aadlc [--dry-run | --apply]`
**Description:** Migrates durable governance knowledge from legacy AADLC
repositories into canonical cARL artefacts so adoption of cARL does not lose
accumulated context. Built around a converter framework: each source implements
a small `Converter` interface (`Discover` + `Classify`) while a shared,
converter-agnostic migration engine performs duplicate detection, conflict
detection, routing, and deterministic reporting. Additional converters
(`claude`, `copilot`, `repo`, ...) can be registered later without reworking the
engine. The AADLC converter discovers artefacts under `.aadlc/`,
`.github/aadlc/`, `aadlc/`, and `AADLC.md` (Markdown + YAML, recursive);
classifies content into invariants, durable memory, and governance rules by
section heading; and routes invariants to `.github/carl/invariants.yml` and
memory/governance entries to a managed block in `.github/carl/memory.md`
(`<!-- BEGIN/END GENERATED: convert aadlc -->`). Existing cARL knowledge is
never overwritten — duplicates are skipped and reported, and conflicts (e.g. a
migrated invariant whose generated `aadlc-` id collides with a different
existing invariant) are reported for human review and never written. AADLC
artefacts are never deleted or modified. `--dry-run` (default) produces the same
report as `--apply` without writing. Idempotent and deterministic — repeated
runs never duplicate content.

### `carl reconcile` Command
**Status:** Delivered
**Command:** `carl reconcile`
**Description:** Reconciles repository-specific durable artefacts so `memory.md`
reflects the current repository rather than only the upstream default runtime.
Reads `.github/carl/repo-map.json` and updates the generated snapshot section in
`.github/carl/memory.md` — covering languages, entry points, key directories,
workflows, governance artefacts, documentation files, and a last-reconciled date.
Human-authored content outside the generated block (delimited by
`<!-- BEGIN GENERATED: reconcile -->` / `<!-- END GENERATED: reconcile -->`) is
never overwritten. If the generated content is unchanged, reports
`No reconciliation needed.` without writing any files. Idempotent. Does not
modify `runtime.json`, harness adapter files, or any other managed artefact.
No network access required.

---

## Near-Term (Candidate Next Items)

### 1. Repo Map Population Tooling
**Status:** Delivered (PR #7) — see Delivered section above.

### 2. Multi-Repository Governance
**Status:** Not started  
**Description:** Guidance for adopting cARL across multiple repositories with shared governance packs. Includes patterns for: central pack repository, fork-and-override, and symlink or CI-copy strategies.  
**Value:** Teams with many repositories need a scalable adoption model.

### 3. cARL Pack for Rust
**Status:** Not started  
**Description:** Rust-specific instruction pack following the same pattern as existing language packs. Should cover: memory safety, unsafe block governance, dependency discipline, testing with `cargo test`, and `clippy` enforcement.

### 4. cARL Pack for Go
**Status:** Delivered (PR #6) — see Delivered section above.

### 5. cARL Pack for C# / .NET
**Status:** Not started  
**Description:** C#/.NET instruction pack. Should cover: nullable reference types, async/await discipline, Entity Framework safety, and .NET-specific secret management.

---

## Medium-Term

### 6. Memory Cache Schema
**Status:** Not started  
**Description:** Define a structured YAML or JSON schema for `memory.md` to enable programmatic reading and writing. Currently it is a freeform markdown document. A schema would support tooling, validation, and agent-driven updates.  
**Design question:** Should memory be YAML front-matter + markdown body, or fully structured JSON?

### 7. PR Contract Validation Tooling
**Status:** Not started  
**Description:** A lightweight CI check that verifies a PR contract exists and is in `active` status before allowing merge. Optionally validates that tests reference contract assertions.  
**Design question:** Should this be a GitHub Action or a standalone script?

### 8. Invariant Enforcement in CI
**Status:** Not started  
**Description:** Parse `invariants.yml` and run automated checks against a PR. For example: detect hardcoded secrets, detect broad rewrite patterns, or enforce plan-before-execute via PR comment presence.

### 9. cARL Adoption Guide
**Status:** Not started  
**Description:** Step-by-step guide for teams adopting cARL into an existing repository. Should cover: minimal adoption (root instructions only), partial adoption (core packs + carl/ artefacts), and full adoption (all packs + plans workflow).

### 10. cARL Pack Health Checks
**Status:** Not started  
**Description:** Tooling to detect stale packs (outdated versions), missing artefacts (memory.md not populated), or pack composition gaps (no cloud pack for a cloud-heavy repository).

---

## Long-Term / Exploratory

### 11. cARL Runtime Metrics
**Status:** Speculative  
**Description:** Capture structured metrics from agent sessions: correction loops consumed, mode switches, contract escalations, invariant violations. Useful for understanding agent behaviour patterns at scale.  
**Design question:** Where should metrics be stored? PR metadata? A dedicated artefact? A separate observability service?

### 12. cARL Marketplace
**Status:** Speculative  
**Description:** A curated, versioned pack registry where teams can discover and adopt community packs for additional languages, platforms, or cloud providers. Similar to GitHub Actions Marketplace.  
**Design question:** How are packs versioned and reviewed for quality and security?

### 13. Cross-Session Memory Persistence
**Status:** Speculative  
**Description:** Explore mechanisms for memory persistence that survive repository forks, renames, and migrations. Currently `memory.md` is tied to a single repository.

### 14. Agent Capability Profile
**Status:** Speculative  
**Description:** A machine-readable declaration of which cARL packs are active in a repository, enabling IDE tooling to surface relevant governance context to developers.

### 15. cARL for Non-Copilot Agents
**Status:** Delivered (PR #10, #11) — harness adapter framework, all five adapters, and adapter file generation all implemented.
**Description:** Adapt cARL governance artefacts for use with other AI coding agents (Cursor, Aider, Claude Code, etc.) that support system-prompt injection from repository files. Harness adapters bridge cARL canonical artefacts to each agent's context injection mechanism. All five adapters (copilot, claude, codex, cursor, antigravity) are supported with detection files and adapter file definitions. Detection: `CLAUDE.md` (Claude Code), `AGENTS.md` (Codex), `.cursorrules` (Cursor), `ANTIGRAVITY.md` (Antigravity), `.github/copilot-instructions.md` (Copilot). Adapter file content generation via `carl harness sync` was delivered in PR #11.
**Design question:** ~~Each agent has different context injection mechanisms. What is the minimal adaptation needed per agent?~~

---

## Open Design Questions

These questions should be resolved before implementing related roadmap items:

1. **Memory schema format** — Freeform markdown vs structured YAML/JSON for `memory.md`?
2. **Pack inheritance** — Should repositories be able to extend a base pack rather than copy it?
3. **Multi-repo governance** — Central pack repository, fork-and-override, or CI-copy model?
4. **CI integration depth** — How much should cARL enforce via CI vs rely on agent compliance?
5. **Community pack quality bar** — What review process should community packs go through before being recommended?
6. **Version pinning** — Should repositories pin specific pack versions or always use latest?
7. **Agent compatibility** — Which agent-specific features (e.g. Copilot instruction packs capability) should cARL depend on vs avoid for portability?

---

## Intentionally Deferred

The following were considered for this initial bootstrap PR and explicitly deferred:

- Structured memory schema (deferred — current freeform markdown is sufficient for v1)
- CI integration tooling (deferred — governance via agent compliance is the v1 model)
- Community pack registry (deferred — single-repository adoption first)
- Non-Copilot agent support (deferred — Copilot is the primary target for v1)
