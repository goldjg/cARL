<!-- version: 1.11.0 -->
# cARL — Roadmap

This roadmap describes the strategic direction and future evolution of cARL. None of the items marked as not started are implemented in the current codebase. Items are recorded here to preserve intent and prevent rediscovery.

---

## Strategic Direction

cARL is evolving from a GitHub Copilot-focused governance system into a **harness-agnostic governance runtime with harness-specific bootloaders**.

The goal is to provide consistent governance — memory, contracts, policies, and operating modes — across heterogeneous coding agents. cARL does not aim to replace agent runtimes (that is a separate concern). It provides the governance layer that any agent runtime can consume.

**Canonical principle:** Governance lives in cARL artefacts. Harness files are adapters that consume governance, not alternate sources of truth.

Beyond harness independence, cARL is evolving towards **policy-as-code for AI coding agents** — the closest conceptual comparison is Open Policy Agent, applied to coding agents, repository context, agent instructions, engineering constraints, validation, and governed execution. cARL should provide a coherent policy layer between repositories, engineering teams, organisational policy, coding-agent harnesses, agent roles, development workflows, and validation/evidence. This direction must not turn cARL into a generic package manager, orchestration platform, agent framework, or enterprise governance monolith, and must preserve its existing strengths: deterministic behaviour, repository-local operation, offline-first design, a self-contained Go binary, explicit committed artefacts, inspectable state, reproducibility, agent-harness independence, and strong validation/reconciliation.

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

The repository implements the same shared-loader architecture for five coding
agent harnesses. Maintainer validation has proven GitHub Copilot, Claude Code,
and Codex end-to-end on this baseline. Cursor and Antigravity have complete,
synchronised harness shims but have not yet been tested in their native
harnesses, so they remain in the `theoretical` tier.

This evidence defines the current architectural model: support starts with a
small harness-native entrypoint that loads the shared cARL adapter, while
end-to-end testing determines whether the support tier can be promoted.

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
- `AGENTS.md` (Codex shim)
- `.cursor/rules/carl.mdc` (Cursor shim)
- `.agents/rules/carl.md` (Antigravity shim)
- Future harness adapter files

`.github/copilot-instructions.md` is both the Copilot harness entrypoint and the **shared cARL adapter loader**. All other harness shim files are tiny files that direct agents to read `.github/copilot-instructions.md` before any repository work. Canonical governance remains under `.github/carl/`.

Adapters should be generated via `carl harness sync` and never manually edited. Drift between an adapter and its canonical source is a health issue, not a design choice.

> **Current state:** The shim model is implemented for all five harnesses.
> `carl harness sync` writes the shared loader once plus the harness-specific
> shim for each synced harness. A shim harness is locally healthy only when
> both the shared loader and the shim are present and synced. Local health does
> not prove that the native harness loaded or obeyed governance.

### Runtime Activation Lifecycle

Governance file presence is not the same as governance activation. Local
adapter detection and byte-for-byte sync are diagnostic evidence, not a claim
about agent behavior. End-to-end harness validation evaluates the following
lifecycle:

1. **Bootstrap** — the harness-specific entrypoint runs before any work begins
2. **Governance discovery** — the agent locates canonical cARL artefacts in the repository
3. **Governance loading** — the agent reads and internalises the governance content
4. **Governance verification** — the agent confirms its operating mode, active contract, and constraints
5. **Governed execution** — all subsequent work operates under the loaded governance context

Implementation details differ per harness (instruction files, rules files, or
future native extensions), but the lifecycle is invariant. A harness-native
skill is an optional future mechanism, not a prerequisite for Claude Code
production support.

### Verification Over Assumption

Future tooling should measure governance activation rather than assume it. This includes:

- Harness readiness validation (is the adapter present, current, and tested in
  the native harness?)
- Governance bootstrap confirmation signal (did the agent emit a structured acknowledgement that governance was loaded? — note: self-reported; not independently verified)
- Adapter health reporting (drift detection between generated adapter and canonical source)
- Cross-harness lifecycle conformance checks

---

## Harness Support Evidence

Support tiers describe validation evidence, not implementation completeness.
All five adapters have generation, detection, and sync-health support. The
maintainer has additionally proven the full shared-loader workflow in Copilot,
Claude Code, and Codex. Cursor and Antigravity are waiting for the equivalent
native-harness test.

**Support tiers:**

| Tier | Meaning |
|---|---|
| Production | Tested, validated, governance reliably activates end-to-end |
| Experimental | Adapter exists, partial validation, governance loading under investigation |
| Theoretical | Adapter is implemented, but no end-to-end native-harness validation has been performed |

| Harness | Entrypoint | Current tier | Evidence |
|---|---|---|---|
| GitHub Copilot | `.github/copilot-instructions.md` | Production | Proven end-to-end |
| Claude Code | `CLAUDE.md` | Production | Proven end-to-end |
| Codex | `AGENTS.md` | Production | Proven end-to-end |
| Cursor | `.cursor/rules/carl.mdc` | Theoretical | Shim implemented and synced; native harness not yet tested |
| Antigravity | `.agents/rules/carl.md` | Theoretical | Shim implemented and synced; native harness not yet tested |

---

## Cross-Harness Governance Lifecycle Pattern

Every future harness must implement the same five-stage lifecycle, regardless of the mechanism used:

| Stage | Description |
|---|---|
| 1. Bootstrap | Harness-specific entrypoint runs before any task begins |
| 2. Governance discovery | Agent locates cARL canonical artefacts in the repository |
| 3. Governance loading | Agent reads and internalises memory, contracts, policies, and instruction packs |
| 4. Governance verification | Agent confirms operating mode, active contract, and active constraints |
| 5. Governed execution | All subsequent work operates under the loaded governance context |

Implementation mechanisms differ across harnesses. The lifecycle does not:

| Harness | Bootstrap mechanism |
|---|---|
| GitHub Copilot | `.github/copilot-instructions.md` instruction file |
| Claude Code | `CLAUDE.md` shim to the shared loader |
| Codex | `AGENTS.md` agent instructions file |
| Cursor | `.cursor/rules/carl.mdc` rules file |
| Antigravity | `.agents/rules/carl.md` rules file |
| Future harnesses | Mechanism differs; lifecycle is invariant |

When adding support for a new harness, the first question is: **how does this harness complete all five lifecycle stages?** Adapter file presence alone is not sufficient. If a harness cannot reliably complete stages 2–4, it remains in the Theoretical tier.

---

## Delivered

### cARL CLI v1 Foundation
**Status:** Delivered (PR #2)
**Commands:** `carl init`, `carl repair`, `carl version`
**Description:** Self-contained Go binary that manages repository-local cARL runtime
installations. All 37 current runtime artefacts are embedded in the binary (no
network required). `runtime.json` is the installation manifest and legacy pack-
selection fallback; user-owned pack selection, profiles, registries, and installed
provenance live in their separate schema-versioned artefacts. `memory.md` and
`runtime.json` are protected from repair. Health status is content-based
(byte-comparison against embedded canonicals). Build-time version and commit
injection via `-ldflags`. Repositories containing cARL artefacts without a
manifest can use explicit non-destructive `carl init --adopt`: existing files
are preserved, missing bundled files are installed, and `runtime.json` is
created last so the normal doctor/repair lifecycle becomes available.

### Release Workflow (CLI Binary Publishing)
**Status:** Delivered (PR #3); migrated to GoReleaser; macOS signing configured from v0.4.2
**Workflow:** `.github/workflows/release.yml`
**Description:** GitHub Actions workflow triggered on `v*` semantic version tags.
Originally used a hand-rolled build matrix. Now uses GoReleaser to build the cARL CLI
for five platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64,
windows/amd64), produce platform archives (tar.gz/zip), generate native Linux
packages (deb/rpm/apk via nfpm), compute SHA-256 checksums, and publish a GitHub
Release. Build-time `cliVersion` plus bundled runtime provenance fields
(`bundledRuntimeVersion`, `bundledRuntimeSource`, `bundledRuntimeTag`,
`bundledRuntimeCommit`) are injected via `-ldflags`.
Homebrew tap publishing is **enabled** via the `goldjg/homebrew-carl` tap.
GoReleaser publishes the cask definition automatically on each tagged release
using `HOMEBREW_TAP_GITHUB_TOKEN`.
WinGet submission is automated in the release workflow via `wingetcreate update`
when `WINGETCREATE_TOKEN` is configured; manual submission remains a fallback.
The release job runs on `macos-latest`; GoReleaser cross-compiles Linux and
Windows binaries on the same runner. GoReleaser OSS `notarize.macos` signs
darwin binaries with a Developer ID Application certificate and hardened
runtime, notarises them through App Store Connect API credentials, then
archives, checksums, and publishes in one `goreleaser release --clean` flow.
The workflow invokes that flow through `scripts/release-with-retry.sh`, which
allows one bounded retry for Apple 429 rate limits, and asserts the expected
release assets before downstream WinGet publication. Five Apple repository
secrets are required: the certificate and password plus the notarisation
issuer ID, key ID, and key (see DISTRIBUTION.md).

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
Supports GitHub Copilot, Claude Code, and Codex as production adapters.
Cursor and Antigravity remain theoretical (adapters exist but have not been
validated end-to-end). `carl harness list`
shows all known adapters with support tier. `carl harness status` detects which
adapter entrypoints are present in the current repository and reports canonical
sync health; it does not claim that the native harness activated governance.
Both subcommands are read-only and always exit 0. Designed for extensibility as
new agents are validated.

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
section (detected, missing, drifted, healthy) without changing overall runtime
status semantics.

### Shared-Loader Effective Pack Hydration
**Status:** Delivered
**Surface:** `.github/copilot-instructions.md`
**Description:** Every supported harness routes through one normative,
repository-local hydration procedure. The loader distinguishes pack presence,
selection, activation, effective dependency/precedence composition, and
override state; uses `packs.json` or the legacy `runtime.json` fallback; applies
profile defaults and active overlays; and hydrates only effective,
non-overridden definitions. Invalid or unresolved state fails closed without a
load-all or filesystem-order fallback. The CLI remains optional management and
diagnostic tooling, and loader/trace availability does not prove model
adherence.

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

### `carl pack list` / `carl pack show` — Pack Discovery & Metadata
**Status:** Delivered
**Command:** `carl pack list [--json]`, `carl pack show <pack-id> [--json]`
**Description:** First vertical slice of the versioned pack runtime. Introduces
a schema-versioned pack metadata model (`schemaVersion: 1`) with stable
`<category>/<name>` IDs, semantic versions, source (`bundled` /
`repository-local`), state (bundled / installed / selected / active),
owned artefacts, and dependency declarations. Discovery merges bundled,
repository-local, and `runtime.json`-selected packs deterministically (sorted
by ID — never filesystem order) and validates the set (malformed or duplicate
IDs, invalid versions, unknown schema versions, missing or cyclic
dependencies, invalid owned artefacts, contradictory state). Human-readable
and `--json` output; structured JSON errors with non-zero exit codes. Works
inside and outside an initialised repository. Explicitly out of scope (future
phases): pack selection/activation commands, priority and override semantics,
a compiled policy intermediate representation, and any remote registry or
dependency downloading.

---

## Pack Runtime Phase Plan (Policy-as-Code Direction)

Planned evolution of the versioned pack runtime. Phases 1 through 6 are
delivered (see above and the Phase 2/3 entries below). Each later phase is a
candidate vertical slice; constraints noted per
phase are binding until explicitly revisited. Each entry is intended to be a
self-sufficient specification: together with `.github/carl/memory.md`, the
existing pack model (`internal/pack`), and the cross-phase constraints below,
"implement pack phase N" should be a complete task prompt. Every slice must
ship with tests, documentation updates (CLI.md, ARCHITECTURE.md, README.md,
GLOSSARY.md as applicable), a refreshed PR contract, and a memory.md
reconciliation decision.

### Pack Phase 2: Pack Composition
**Status:** Delivered
**Commands:** `carl pack select`, `carl pack unselect`, `carl pack effective`
**Description:** Repository pack selection commands, dependency expansion,
computation of the effective pack set, conflict detection, precedence rules,
and explicit override handling. Selection is persisted in a schema-versioned,
user-owned committed artefact (`.github/carl/packs.json`, deduplicated and
sorted; `runtime.json` remains init-only, with legacy managed-artefact
derivation as the fallback when `packs.json` is absent). Composition metadata
comes only from explicit pack file headers (`requires:`, `precedence-mode:`,
`priority:`, `overrides:`); absent metadata defaults to no dependencies,
additive mode, priority 0, and no overrides. `carl pack effective` expands
required dependencies transitively with explicit per-pack reasons, orders by
precedence (priority descending, pack-ID tie-break — never load order), and
detects conflicts (missing dependencies, overriding a non-overridable pack,
mutual overrides) with non-zero exit. Composition remains conservative:
non-overridden effective packs add constraints; overridden packs stay visible
in the evaluation flagged `overriddenBy` for provenance but are not applied.
Override authority is explicit metadata (target must declare
`precedence-mode: overridable`), never inferred from load order.

### Pack Phase 3: Profiles and Agent Roles
**Status:** Delivered
**Commands:** `carl pack profile list`, `show`, `activate`, `clear`
**Description:** Named policy profiles, role-specific context, task-specific
pack selection, repository and organisation defaults, and controlled
customisation. Profile policy is stored in the schema-versioned, user-owned
`.github/carl/profiles.json` artefact. Organisation defaults, repository
defaults, active-profile packs, and active role/task overlays compose
additively and provide explicit activation reasons. Every referenced pack
must already be selected; profile activation never changes selection,
precedence, or override authority. `state.active` and `carl pack effective`
are profile-driven when the artefact exists, with selected-as-active retained
only as a compatibility fallback for repositories without `profiles.json`.
Profile commands provide human-readable and schema-versioned JSON output,
validate all identifiers and references, and write only `profiles.json`;
`runtime.json` remains init-only.

Fresh installations ship `.github/carl/profiles.example.json` as an inactive,
cloneable `default` baseline with explicit references to the complete bundled
pack set and the existing role-neutral context. It is ordinary profile data,
not evaluator magic, and repositories without `profiles.json` retain the
compatibility fallback. A later menu/TUI workflow may create, clone, and edit
profiles on this same canonical model; a shape such as
`carl pack profile clone default my-team` is non-binding direction, not a
currently implemented command or a new state format.

### Pack Phase 4: Registry and Installation
**Status:** Delivered
**Commands:** `carl pack registry list`, `registry search`, `install`, `update`
**Description:** Explicit schema-versioned registry configuration supports
optional HTTPS and repository-local indexes without adding an implicit remote
authority. Releases declare semantic versions, relative artifacts, and
SHA-256 digests. Resolution selects the highest version deterministically or
an exact requested version and rejects equal-version cross-registry
ambiguity. Installation verifies all artifacts and required dependencies
before mutation, confines writes to canonical instruction-pack paths, records
deterministic provenance in `.github/carl/installed-packs.json`, and never
selects/activates packs or writes `runtime.json`. Updates use recorded
provenance, reject local drift and same-version registry mutation, and never
downgrade. Existing pack commands remain network-free and local registries
support fully offline workflows. SHA-256 is accurately scoped as integrity
against the configured index, not publisher authentication or signing.

### Pack Phase 5: Policy Provenance and Explanation
**Status:** Delivered
**Commands:** `carl explain <pack-id>`, `carl trace`
**Description:** Deterministic, schema-versioned, read-only commands report
pack-level policy provenance: which packs are effective, where their canonical
definitions came from, which selection/profile/default/dependency activated
them, their precedence, whether they add constraints, and which explicit
overrides resolved or remained conflicts. Trace order reuses the existing
effective-pack evaluator and unresolved conflicts keep their non-zero exit.
Both commands are local-only and network-free. The policy unit is an
instruction pack; the commands do not interpret individual natural-language
rules or claim to expose prompts, hidden model reasoning, or chain-of-thought.
Overridden entries remain visible for provenance with `applied: false` and a
structured non-application decision.

### Pack Phase 6: Cognitive Repository Graph
**Status:** Delivered
**Description:** `carl map` preserves its existing inventory fields and adds a
schema-versioned deterministic graph of repository components, Go packages,
entry points, workflows, governance artefacts, documentation, and policy
definitions. Structural containment and repository-local Go imports produce
evidence-backed edges; direct reverse imports produce bounded change-impact
references. Nodes classify criticality and trust boundaries, identify policy
attachment points, and provide agent context. Coverage metadata states whether
ownership, dependencies, runtime data flows, policy attachment, and impact
knowledge is derived, partial, or unavailable. The graph never guesses owners
or runtime flows and does not replace `carl trace` as the active-policy
provenance authority.

### Pack Phase 7: Publishing
**Status:** Not started — do not implement prematurely
**Description:** Internal pack distribution, organisation policy catalogues,
signed releases, compatibility metadata, and reusable profiles.

**Cross-phase constraints:** remain offline-capable with no runtime network
dependencies; self-contained Go binary; avoid unnecessary third-party
dependencies; preserve initialised repositories and existing CLI/exit-code
conventions; explicit schema versioning; deterministic output; handle absent
new metadata gracefully; no harness-specific core architecture. The effective
rule set and material policy meaning must remain equivalent across refactors.

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

### 6. Harness-Native Extensions
**Status:** Exploratory
**Description:** Add harness-native skills or commands only where native-harness
testing shows that the shared-loader shim is insufficient or a native extension
adds clear value. Such extensions must remain generated adapters rather than
canonical governance sources. No harness-native extension is currently required
for production support.

### 7. End-to-End Harness Conformance Validation
**Status:** Not started  
**Description:** Define a repeatable native-harness validation protocol for the
five lifecycle stages and add a `carl harness validate [<harness-id>]` command
(or equivalent evidence surface) that distinguishes local adapter health from
end-to-end conformance. Cursor and Antigravity are the first unvalidated
targets. Any self-reported agent signal must remain explicitly weaker evidence
than an externally observed test.

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
**Status:** Partially delivered
**Description:** Current discovery and composition already reject malformed or
unsupported schemas, invalid semantic versions, missing/cyclic dependencies,
contradictory state, invalid ownership, unresolved conflicts, and installed-
provenance drift. Remaining health work is advisory rather than foundational:
detecting that newer pack versions are available, identifying unpopulated
durable artefacts, and recommending composition gaps such as a missing cloud
pack for a cloud-heavy repository.

### 14. Adapter Drift Detection in CI
**Status:** Not started  
**Description:** A CI check that detects drift between generated harness adapter files and their canonical embedded sources. Fails if an adapter file has been manually edited or if `carl harness sync` has not been re-run after a cARL upgrade. Prevents governance divergence from going unnoticed between releases.

Local adapter drift detection is already delivered through `carl harness
status`, `carl status`, and `carl doctor`; this item is specifically the missing
CI enforcement layer.

---

## Long-Term / Exploratory

### 15. Multi-Harness Governance Runtime
**Status:** Delivered baseline (PR #9, #11, #37); automated cross-harness conformance remains not started
**Description:** Implements the multi-harness governance runtime architecture
defined above. Delivered state:

- All five harness adapters are implemented with detection files and adapter file definitions (copilot, claude, codex, cursor, antigravity)
- `carl harness sync` generates adapter files from canonical embedded artefacts
- `carl harness status` and `carl doctor` surface adapter health
- Copilot, Claude Code, and Codex are proven and production-supported
- Cursor and Antigravity are implemented but remain theoretical pending native-harness testing

Remaining generic milestones (see Cross-Harness and Near-Term sections above):

- Repeatable native-harness evidence capture for Cursor and Antigravity
- A clear distinction between locally detected/synced adapters and validated activation
- Cross-harness lifecycle conformance tooling

### 16. cARL Runtime Metrics
**Status:** Speculative  
**Description:** Capture structured metrics from agent sessions: correction loops consumed, mode switches, contract escalations, invariant violations. Useful for understanding agent behaviour patterns at scale.  
**Design question:** Where should metrics be stored? PR metadata? A dedicated artefact? A separate observability service?

### 17. cARL Marketplace
**Status:** Speculative  
**Description:** A curated public discovery and quality layer where teams can
find community packs for additional languages, platforms, or cloud providers.
The explicit registry, verified installation, update, and provenance plumbing
is already delivered in Pack Phase 4; this item is the curated catalogue,
review, and publisher-trust experience above that plumbing. Similar to GitHub
Actions Marketplace.
**Design question:** How are packs versioned and reviewed for quality and security?

### 18. Cross-Session Memory Persistence
**Status:** Speculative  
**Description:** Explore mechanisms for memory persistence that survive repository forks, renames, and migrations. Currently `memory.md` is tied to a single repository.

### 19. Agent Capability Profile
**Status:** Delivered runtime surface; IDE integration not started
**Description:** `.github/carl/packs.json` and `.github/carl/profiles.json`
provide machine-readable selection and active profile context, while
`carl pack effective --json`, `carl explain --json`, and `carl trace --json`
report the evaluated capability/policy set and provenance. IDE tooling that
surfaces this context to developers remains future work.

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
9. **Harness-native extension versioning** — If a future harness requires a native skill or command beyond its shim, how should that generated extension be versioned and health-checked?
10. **Cross-harness lifecycle conformance** — How should `carl harness validate` determine that governance loading (not just discovery) succeeded for a given harness?

---

## Intentionally Deferred

The following were considered for the initial bootstrap PR and explicitly deferred:

- Structured memory schema (deferred — current freeform markdown is sufficient for v1)
- CI integration tooling (deferred — governance via agent compliance is the v1 model)
- Curated public pack marketplace (deferred; explicit registry plumbing is delivered)
- Additional harness validation (the multi-harness baseline is delivered; Cursor and Antigravity native-harness testing remains)
