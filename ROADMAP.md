<!-- version: 1.1.0 -->
# cARL — Roadmap

This roadmap describes the strategic direction and future evolution of cARL. None of the items marked as not started are implemented in the current codebase. Items are recorded here to preserve intent and prevent rediscovery.

---

## Strategic Direction

cARL is evolving from a GitHub Copilot-focused governance system into a **harness-agnostic governance runtime with harness-specific bootloaders**.

The goal is to provide consistent governance — memory, contracts, policies, and operating modes — across heterogeneous coding agents. cARL does not aim to replace agent runtimes (that is a separate concern). It provides the governance layer that any agent runtime can consume.

**Canonical principle:** Governance lives in cARL artefacts. Harness files are adapters that consume governance, not alternate sources of truth.

---

## Guiding Principles for Roadmap Items

- **Preserve runtime semantics** — new capabilities should extend, not change, existing governance behaviour
- **One concern per pack** — new instruction packs should remain focused
- **Version-controlled artefacts** — new governance artefacts belong in `.github/carl/`
- **Backward compatibility first** — existing users should not need to change their setup
- **Harnesses consume, not own** — Harness-specific files are adapter artefacts intended to be generated; the canonical source lives in cARL.
- **Measure activation, not presence** — governance files existing is not the same as governance being active

---

## Architectural Direction: Multi-Harness Governance Runtime

Recent field testing revealed that different coding agents consume governance differently. GitHub Copilot is effectively solved through `.github/copilot-instructions.md`. Claude Code, however, does not reliably operate under cARL governance through `CLAUDE.md` alone — successful operation required a dedicated `/carl` skill that explicitly discovers, loads, and activates governance before work begins.

This experience defines the architectural model for all future harness support.

### Canonical Governance (Harness-Independent)

cARL governance artefacts are the single source of truth, independent of any coding agent:

- `.github/carl/memory.md`
- `.github/carl/current-pr-contract.md`
- `.github/carl/tool-policy.yml`
- `.github/carl/invariants.yml`
- `.github/carl/repo-map.json`
- `.github/instructions/` packs
- Future governance artefacts

Harnesses must **consume** these artefacts. They must not become alternate sources of truth or duplicate governance content in agent-specific files.

### Harness Adapters (Generated Bootloaders)

Harness-specific files are adapters and bootloaders — generated from cARL canonical artefacts and treated as implementation outputs rather than primary governance sources:

- `.github/copilot-instructions.md` (Copilot — shared loader and Copilot entrypoint)
- `CLAUDE.md` (Claude Code shim)
- `.claude/skills/carl.md` (Claude cARL skill)
- `AGENTS.md` (Codex shim)
- `.cursor/rules/carl.mdc` (Cursor shim)
- `.agents/rules/carl.md` (Antigravity shim)
- Future harness adapter files

`.github/copilot-instructions.md` is both the Copilot harness entrypoint and the **shared cARL adapter loader**. All other harness shim files are tiny files that direct agents to read `.github/copilot-instructions.md` before any repository work. Canonical governance remains under `.github/carl/`.

Adapters should be generated via `carl harness sync` and never manually edited. Drift between an adapter and its canonical source is a health issue, not a design choice.

> **Current state:** The shim model is implemented. `carl harness sync` writes the shared loader once plus the harness-specific shim for each synced harness. A shim harness is healthy only when both the shared loader and the shim are present and synced.

### Runtime Activation Lifecycle

Governance file presence is not the same as governance activation. A harness adapter is not considered successful merely because governance files exist. Every harness must complete the following lifecycle:

1. **Bootstrap** — the harness-specific bootloader or skill runs before any work begins
2. **Governance discovery** — the agent locates canonical cARL artefacts in the repository
3. **Governance loading** — the agent reads and internalises the governance content
4. **Governance verification** — the agent confirms its operating mode, active contract, and constraints
5. **Governed execution** — all subsequent work operates under the loaded governance context

Implementation details differ per harness (skills, instruction files, rules files, etc.), but the lifecycle is invariant.

### Verification Over Assumption

Future tooling should measure governance activation rather than assume it. This includes:

- Harness readiness validation (is the adapter present, current, and bootstrapped?)
- Governance bootstrap confirmation signal (did the agent emit a structured acknowledgement that governance was loaded? — note: self-reported; not independently verified)
- Adapter health reporting (drift detection between generated adapter and canonical source)
- Cross-harness lifecycle conformance checks

---

## Claude Code Support

Claude Code is the primary post-Copilot validation target. Field testing has confirmed that `CLAUDE.md` alone is insufficient for reliable governance activation. The following work items formalise the Claude harness as a first-class supported adapter.

### Claude Bootstrap Model
**Status:** Not started  
**Description:** Implement the tooling and workflow for generating and managing Claude-specific governance bootstrap artefacts:

- `CLAUDE.md` generation via `carl harness sync`
- cARL skill generation (`.claude/skills/carl.md`)
- Skill installation workflow (how to install the skill into a Claude Code project)
- Skill update workflow (how to update the skill when cARL canonical content changes)
- Skill versioning (version header in the skill file; drift detection against the embedded canonical)

The skill must be treated as a generated adapter output, regenerated by `carl harness sync`, and health-checked by `carl harness status` and `carl doctor`.

### Claude Governance Loader (the `/carl` Skill)
**Status:** Not started  
**Description:** Formalise the `/carl` skill concept as the canonical Claude Code governance bootloader. The skill must implement the full runtime activation lifecycle:

1. Locate canonical cARL governance artefacts in the repository
2. Load and summarise the governance content (memory, tool policy, instruction packs)
3. Report the active operating mode (Plan-only, Assisted implementation, Automatic)
4. Report the active PR contract state (active, draft, none)
5. Report memory cache and tool policy status
6. Explicitly confirm governed operating mode before any work begins

The skill content should be generated from a canonical template embedded in the cARL CLI binary and served via `carl harness sync claude`.

### Claude Harness Validation
**Status:** Not started  
**Description:** Define the validation criteria for determining whether a Claude Code installation is governance-ready:

- Skill installed (`.claude/skills/carl.md` present)
- Skill version matches the current canonical (not drifted)
- Governance artefacts present (memory, tool policy, instructions)
- Bootstrap operational (skill can be invoked and loads governance)
- Governance successfully loaded (operating mode confirmed, not just artefacts present)

Surface this as a `carl harness status --verbose claude` report and as `carl doctor` findings.

### Claude Support Tier
**Current tier:** Experimental  
**Target tier:** Production  
**Description:** Define the criteria for promoting Claude Code from experimental to production support. Production tier requires:

- Reliable governance activation (bootstrap lifecycle fully operational)
- Adapter health checks passing
- End-to-end validation documented
- Skill generation and update workflows delivered

**Support tiers:**
| Tier | Meaning |
|---|---|
| Production | Tested, validated, governance reliably activates end-to-end |
| Experimental | Adapter exists, partial validation, governance loading under investigation |
| Theoretical | Adapter exists, no end-to-end validation performed |

---

## Cross-Harness Governance Lifecycle Pattern

Every future harness must implement the same five-stage lifecycle, regardless of the mechanism used:

| Stage | Description |
|---|---|
| 1. Bootstrap | Harness-specific bootloader or skill runs before any task begins |
| 2. Governance discovery | Agent locates cARL canonical artefacts in the repository |
| 3. Governance loading | Agent reads and internalises memory, contracts, policies, and instruction packs |
| 4. Governance verification | Agent confirms operating mode, active contract, and active constraints |
| 5. Governed execution | All subsequent work operates under the loaded governance context |

Implementation mechanisms differ across harnesses. The lifecycle does not:

| Harness | Bootstrap mechanism |
|---|---|
| GitHub Copilot | `.github/copilot-instructions.md` instruction file |
| Claude Code | `.claude/skills/carl.md` skill invoked via `/carl` |
| Cursor | `.cursor/rules/carl.mdc` rules file |
| Codex | `AGENTS.md` agent instructions file |
| Future harnesses | TBD — mechanism differs, lifecycle is invariant |

When adding support for a new harness, the first question is: **how does this harness complete all five lifecycle stages?** Adapter file presence alone is not sufficient. If a harness cannot reliably complete stages 2–4, it remains in the Theoretical tier.

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
**Status:** Delivered (PR #3); migrated to GoReleaser; macOS signing and notarisation configured from v0.4.2
**Workflow:** `.github/workflows/release.yml`
**Description:** GitHub Actions workflow triggered on `v*` semantic version tags.
Originally used a hand-rolled build matrix. Now uses GoReleaser to build the cARL CLI
for five platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64,
windows/amd64), produce platform archives (tar.gz/zip), generate native Linux
packages (deb/rpm/apk via nfpm), compute SHA-256 checksums, and publish a GitHub
Release. Build-time `cliVersion` and `sourceCommit` are injected via `-ldflags`.
Homebrew tap publishing is **enabled** via the `goldjg/homebrew-carl` tap.
GoReleaser publishes the cask definition automatically on each tagged release
using `HOMEBREW_TAP_GITHUB_TOKEN`.
WinGet submission is automated in the release workflow via `wingetcreate update`
when `WINGETCREATE_TOKEN` is configured; manual submission remains a fallback.
The release job runs on `macos-latest` so that `codesign` is available.
GoReleaser cross-compiles Linux and Windows binaries on the same
runner. darwin binaries are signed inline by a GoReleaser post-hook
(`.github/scripts/codesign-darwin.sh`) immediately after each darwin binary is
built. GoReleaser then runs a single `goreleaser release --clean` which builds,
signs (via hook), archives, checksums, and publishes the GitHub Release in one
step. darwin artefacts are codesigned (Developer ID Application, hardened
runtime) but not notarised; Gatekeeper may prompt on first run. Full
notarisation requires switching to App Store Connect API key auth and configuring
`notarize.macos` — see DISTRIBUTION.md. Uses OSS GoReleaser only —
no GoReleaser Pro features. Two Apple repository secrets are required for macOS
signing (see DISTRIBUTION.md).

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
Supports GitHub Copilot as the production-validated primary adapter (detection via
`.github/copilot-instructions.md`). Registers Claude Code as experimental (partial
validation, governance loading under investigation) and Codex, Cursor, and Antigravity
as theoretical (adapters exist but not yet validated end-to-end). `carl harness list`
shows all known adapters with support tier. `carl harness status` detects which
harnesses are active in the current repository. Both subcommands are read-only and
always exit 0. Designed for extensibility as new agents are validated.

### `carl harness sync` — Harness Adapter File Generation
**Status:** Delivered (PR #11)
**Command:** `carl harness sync [<harness-id>...]`
**Description:** Adds a `sync` subcommand to `carl harness` that generates adapter
files for all harnesses with defined adapter files (or a named subset) from the
canonical cARL artefacts embedded in the CLI binary. Adapter files are treated as
disposable outputs — always regenerated from the canonical source and never edited
manually. All five harnesses (copilot, claude, codex, cursor, antigravity) have
adapter files; sync works for all tiers (production, experimental, theoretical).
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
report as `--apply` without writing. If the managed convert block's markers in
`memory.md` are malformed (begin without end, end without begin, or end before
begin) the command fails with a non-zero exit and writes nothing rather than
appending a second generated block — mirroring `carl reconcile`'s marker safety.
Idempotent and deterministic — repeated runs never duplicate content.

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

### 6. Harness Adapter Generation — Skill Support
**Status:** Not started  
**Description:** Extend `carl harness sync` to generate not only flat adapter files (e.g. `CLAUDE.md`) but also structured skill artefacts (e.g. `.claude/skills/carl.md`). The skill template should be embedded in the CLI binary alongside other canonical artefacts, health-checked by `carl doctor`, and regenerated by `carl harness sync claude`.

### 7. Harness Readiness Validation
**Status:** Not started  
**Description:** Add a `carl harness validate [<harness-id>]` command (or extend `carl harness status --verbose`) that reports whether a harness has completed all five lifecycle stages, not just whether adapter files are present. For Claude Code this means verifying the skill is installed, current, and can load governance. Surface failures as actionable `carl doctor` findings with specific remediation steps.

### 8. Governance Bootstrap Confirmation Signal (Exploratory)
**Status:** Not started  
**Description:** Explore a machine-readable governance bootstrap report format. When an agent completes the runtime activation lifecycle, it could emit a *structured confirmation signal* (e.g. a YAML or JSON artefact) recording: operating mode confirmed, PR contract state, memory loaded, tool policy loaded, timestamp. This is intentionally framed as exploratory: an agent can only self-report, and self-reporting is not proof. Treat any signal as a useful hint, not a guarantee that governance was active. Stronger assurance would require CLI-observed checks or CI evidence (see item 7).

---

## Medium-Term

### 9. Memory Cache Schema
**Status:** Not started  
**Description:** Define a structured YAML or JSON schema for `memory.md` to enable programmatic reading and writing. Currently it is a freeform markdown document. A schema would support tooling, validation, and agent-driven updates.  
**Design question:** Should memory be YAML front-matter + markdown body, or fully structured JSON?

### 10. PR Contract Validation Tooling
**Status:** Not started  
**Description:** A lightweight CI check that verifies a PR contract exists and is in `active` status before allowing merge. Optionally validates that tests reference contract assertions.  
**Design question:** Should this be a GitHub Action or a standalone script?

### 11. Invariant Enforcement in CI
**Status:** Not started  
**Description:** Parse `invariants.yml` and run automated checks against a PR. For example: detect hardcoded secrets, detect broad rewrite patterns, or enforce plan-before-execute via PR comment presence.

### 12. cARL Adoption Guide
**Status:** Not started  
**Description:** Step-by-step guide for teams adopting cARL into an existing repository. Should cover: minimal adoption (root instructions only), partial adoption (core packs + carl/ artefacts), and full adoption (all packs + plans workflow).

### 13. cARL Pack Health Checks
**Status:** Not started  
**Description:** Tooling to detect stale packs (outdated versions), missing artefacts (memory.md not populated), or pack composition gaps (no cloud pack for a cloud-heavy repository).

### 14. Adapter Drift Detection in CI
**Status:** Not started  
**Description:** A CI check that detects drift between generated harness adapter files and their canonical embedded sources. Fails if an adapter file has been manually edited or if `carl harness sync` has not been re-run after a cARL upgrade. Prevents governance divergence from going unnoticed between releases.

---

## Long-Term / Exploratory

### 15. Multi-Harness Governance Runtime
**Status:** In progress — harness framework and adapter file generation delivered (PR #9, #11); Claude Code bootstrap model and cross-harness validation pending  
**Description:** Implements the multi-harness governance runtime architecture defined above. Current state:

- All five harness adapters are implemented with detection files and adapter file definitions (copilot, claude, codex, cursor, antigravity)
- `carl harness sync` generates adapter files from canonical embedded artefacts
- `carl harness status` and `carl doctor` surface adapter health
- Copilot is production-validated; Claude Code is experimental (governance loading requires dedicated `/carl` skill); Codex, Cursor, and Antigravity are theoretical

Next milestones (see Claude Code Support and Cross-Harness sections above):

- Claude bootstrap model (skill generation, installation, update, versioning)
- Claude governance loader (formalised `/carl` skill implementing the full activation lifecycle)
- Claude harness validation (skill health checks, governance activation verification)
- Cross-harness lifecycle conformance tooling
- Claude promotion from Experimental to Production tier

### 16. cARL Runtime Metrics
**Status:** Speculative  
**Description:** Capture structured metrics from agent sessions: correction loops consumed, mode switches, contract escalations, invariant violations. Useful for understanding agent behaviour patterns at scale.  
**Design question:** Where should metrics be stored? PR metadata? A dedicated artefact? A separate observability service?

### 17. cARL Marketplace
**Status:** Speculative  
**Description:** A curated, versioned pack registry where teams can discover and adopt community packs for additional languages, platforms, or cloud providers. Similar to GitHub Actions Marketplace.  
**Design question:** How are packs versioned and reviewed for quality and security?

### 18. Cross-Session Memory Persistence
**Status:** Speculative  
**Description:** Explore mechanisms for memory persistence that survive repository forks, renames, and migrations. Currently `memory.md` is tied to a single repository.

### 19. Agent Capability Profile
**Status:** Speculative  
**Description:** A machine-readable declaration of which cARL packs are active in a repository, enabling IDE tooling to surface relevant governance context to developers.

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
8. **Bootstrap confirmation signal format** — If an agent self-reports governance activation, what is the right format for that structured confirmation signal? YAML artefact, PR comment, or structured log? How do we distinguish signal from proof (see item 8 in Near-Term)?
9. **Skill versioning** — Should the `/carl` skill embed a version header, and should `carl doctor` detect version mismatches vs the current CLI binary?
10. **Cross-harness lifecycle conformance** — How should `carl harness validate` determine that governance loading (not just discovery) succeeded for a given harness?

---

## Intentionally Deferred

The following were considered for the initial bootstrap PR and explicitly deferred:

- Structured memory schema (deferred — current freeform markdown is sufficient for v1)
- CI integration tooling (deferred — governance via agent compliance is the v1 model)
- Community pack registry (deferred — single-repository adoption first)
- Non-Copilot agent support (deferred — Copilot is the primary target for v1; superseded by roadmap item 15, Multi-Harness Governance Runtime)
