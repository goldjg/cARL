<!-- version: 2.4.0 -->
# Durable Architectural Truth Cache

This cache stores durable project truths that should persist beyond a
single task. Update it only when a stable fact, decision, invariant, or
unresolved question should carry forward.

## Project purpose
cARL (Cognitive Agent Runtime Layer) is a reusable governance and
instruction layer for AI coding agents. It provides modular instruction
packs, durable memory artefacts, harness-specific shared-loader adapters, and
cARLv2 cognition governance for consistent, secure, maintainable, and governed
AI-assisted development. GitHub Copilot, Claude Code, and Codex are proven
production harnesses. Cursor and Antigravity have implemented and synchronised
adapters but remain theoretical until native-harness validation is performed.

Strategic direction: cARL is evolving towards policy-as-code for AI coding
agents (closest analogy: Open Policy Agent, applied to coding agents,
repository context, agent instructions, engineering constraints, validation,
and governed execution). The staged plan lives in ROADMAP.md under "Pack
Runtime Phase Plan" — when asked to "implement the next roadmap item" or
"implement pack phase N", treat that ROADMAP section plus this memory file as
the authoritative specification.

## Non-goals
cARL must not become a generic package manager, orchestration platform,
agent framework, or enterprise governance monolith. Preserve: deterministic
behaviour, repository-local operation, offline-first design, self-contained
Go binary, explicit committed artefacts, inspectable state, reproducibility,
agent-harness independence, strong validation and reconciliation. No runtime
network dependencies; access to an explicitly configured HTTPS or
repository-local pack registry stays optional. Never
infer policy order from filesystem order; never treat generated files as
canonical.

## Architecture summary

cARL artefacts are the canonical source of governance truth for this repository.

`.github/carl/` contains durable governance artefacts, repository memory, PR contracts, invariants, trust boundaries, tool policy, plans, runtime metadata, and generated repository maps.

`.github/instructions/` contains modular single-concern instruction packs used by supported harnesses.

Harness-specific files such as `.github/copilot-instructions.md`, `CLAUDE.md`, `AGENTS.md`, `.cursor/rules/carl.mdc`, and `.agents/rules/carl.md` are adapters/shims. They may load, summarise, or route agents toward cARL, but they are not the canonical governance authority.

`.github/copilot-instructions.md` is both the GitHub Copilot harness entrypoint and the **shared cARL adapter loader** for all other harness shims. It is located at that path for Copilot compatibility. All other harness entrypoints (CLAUDE.md, AGENTS.md, .cursor/rules/carl.mdc, .agents/rules/carl.md) are tiny shim files that tell the harness to read `.github/copilot-instructions.md` before any repository work. It should remain a thin, procedural loader that makes the cARL lifecycle explicit:

1. hydrate cARL and derive the effective, non-overridden instruction packs
   from repository-local runtime, selection, profile, and pack metadata before
   planning or implementation;
2. apply cARL governance during execution;
3. validate contract, implementation, and tests together;
4. reconcile documentation and durable cARL artefacts before final response;
5. report whether cARL/docs updates were required.

If prompt/session memory conflicts with cARL artefacts, trust cARL and report the conflict.

If `.github/carl/memory.md` conflicts with current repository state, current repository state wins and memory should be updated.

`runtime.json` is the installation manifest and legacy pack-selection fallback,
not the sole authority for all mutable runtime state. User-owned pack
selection, policy profiles, explicit registries, and registry installation
provenance live in `packs.json`, `profiles.json`, `registries.json`, and
`installed-packs.json` respectively. The current binary embeds 38 runtime
artefacts.

`carl init --adopt` is the explicit recovery path when embedded cARL artefacts
already exist but `runtime.json` is absent. Adoption preserves every existing
artefact, installs only missing bundled artefacts, and creates the manifest
last. Ordinary `carl init` retains collision-safe failure. Adoption establishes
future repair ownership, but repair continues to protect `memory.md` and
`runtime.json`.

## Release infrastructure

The cARL CLI release pipeline uses **GoReleaser** (`.goreleaser.yaml`) as the canonical packaging layer. Releases are tag-triggered (`v*`). GoReleaser produces:

- Platform archives (tar.gz for Linux/macOS, zip for Windows)
- Native Linux packages: deb, rpm, apk (via nfpm)
- SHA-256 checksums
- GitHub Release with all artefacts attached

The release job runs on **`macos-latest`**. GoReleaser cross-compiles Linux and Windows binaries on the same runner — no separate ubuntu job is needed. darwin binaries are signed and notarised through GoReleaser OSS `notarize.macos` (Developer ID Application + hardened runtime + App Store Connect API key) before archiving. The release workflow uses a single **`goreleaser release --clean`** which builds, signs, notarises, archives, checksums, and publishes the GitHub Release in one step.

Five Apple repository secrets are required for macOS signing/notarisation:
`MACOS_CERTIFICATE_P12_BASE64`, `MACOS_CERTIFICATE_PASSWORD`, `NOTARIZE_ISSUER_ID`, `NOTARIZE_KEY_ID`, and `NOTARIZE_KEY`. See DISTRIBUTION.md for setup.

Homebrew tap publishing is **enabled** via the `goldjg/homebrew-carl` tap. GoReleaser publishes the cask definition automatically on each tagged release; `HOMEBREW_TAP_GITHUB_TOKEN` must be set as a repository secret with `Contents: write` access to `goldjg/homebrew-carl`. Required Apple and Homebrew credentials are checked before publication. Same-tag GoReleaser recovery preserves existing release notes and replaces same-name GitHub Release assets; the outer retry remains limited to Apple-context 429/notary rate-limit failures and preserves the final failure status. WinGet submission is automated in the release workflow via `wingetcreate update` when `WINGETCREATE_TOKEN` is configured; otherwise manual submission remains available (see `DISTRIBUTION.md`). Enterprise mirroring into JFrog Artifactory or similar is documented in `DISTRIBUTION.md` but not automated in CI.

The stable public compatibility boundary for the `1.x` line is documented in
`COMPATIBILITY.md`. Documented CLI semantics and exit behaviour,
schema-versioned JSON and repository artefacts, pack metadata/composition
semantics, provenance, repository-map schema, ownership boundaries, and the
documented lifecycle commands are stable. Compatible evolution may add JSON
fields; changing or removing field meaning requires a schema transition.
User-owned policy state is never silently replaced, repair remains confined to
declared repairable assets, and intentional breaking changes to stable public
contracts require a major version. Human-readable formatting, prose,
compatible bundled policy revisions, internal implementation detail, and
explicitly theoretical harness behaviour are not byte-for-byte promises.

<!-- BEGIN GENERATED: reconcile -->
## Repository snapshot

This section is regenerated by `carl reconcile`. Do not edit manually.

**Languages:** Go, Shell
**Last reconciled:** 2026-07-26

### Entry points

- `go.mod` — Go module definition: github.com/goldjg/carl
- `cmd/carl/main.go` — carl CLI entry point

### Key directories

- `.agents`
- `.agents/rules`
- `.cursor`
- `.cursor/rules`
- `.github` — GitHub configuration and Copilot instruction root
- `.github/carl` — cARLv2 governance artefacts and templates
- `.github/carl/plans` — Prompt-as-code planning artefacts
- `.github/instructions` — Copilot instruction packs
- `.github/instructions/cloud` — Cloud guidance packs
- `.github/instructions/core` — Core governance packs
- `.github/instructions/enterprise`
- `.github/instructions/languages` — Language-specific guidance packs
- `.github/instructions/platform` — Platform guidance packs
- `.github/scripts`
- `.github/workflows` — GitHub Actions workflows
- `cmd` — CLI command entry points
- `cmd/carl` — Is the cARL CLI — a governance runtime manager for coding agents.
- `embedded` — Provides access to the cARL runtime assets bundled into the CLI binary.
- `embedded/assets` — Embedded runtime asset mirror
- `internal` — Internal implementation packages
- `internal/cmdutil` — Defines shared types for cARL CLI commands.
- `internal/convert` — Implements the `carl convert` command and its converter framework.
- `internal/doctor` — Implements the `carl doctor` command.
- `internal/harness` — Implements the `carl harness` command and its subcommands.
- `internal/install` — Implements the `carl init` command.
- `internal/manifest` — Manages the cARL runtime manifest (.github/carl/runtime.json).
- `internal/pack` — Implements the `carl pack` command and its subcommands.
- `internal/plan` — Implements the `carl plan` command.
- `internal/reconcile` — Implements the `carl reconcile` command.
- `internal/repair` — Implements the `carl repair` command.
- `internal/repomap` — Implements the `carl map` command.
- `internal/status` — Implements the `carl status` command.
- `internal/version` — Implements the `carl version` command.
- `scripts` — Build and automation scripts
- `session-records`

### Workflows

- `.github/workflows/goreleaser-check.yml` — goreleaser-check workflow
- `.github/workflows/release.yml` — release workflow

### Governance artefacts

- `.github/carl/current-pr-contract.md` — Active PR scope and constraints
- `.github/carl/current-pr-contract.template.md` — PR contract template
- `.github/carl/enterprise-profiles.md`
- `.github/carl/invariants.yml` — Runtime invariants enforced by all implementation PRs
- `.github/carl/memory.md` — Durable architectural truth cache
- `.github/carl/profiles.enterprise.example.json`
- `.github/carl/profiles.enterprise.scenarios.example.json`
- `.github/carl/profiles.example.json`
- `.github/carl/repo-map.example.json` — Repo map template example
- `.github/carl/repo-map.json` — Generated cognitive repository map
- `.github/carl/runtime.json`
- `.github/carl/tool-policy.yml` — Tool permission tier definitions
- `.github/carl/trust-boundaries.md` — Trust boundary documentation

### Documentation

- `AGENTS.md`
- `ARCHITECTURE.md` — Architecture documentation
- `CLAUDE.md`
- `CLI.md` — CLI command reference
- `COMPATIBILITY.md`
- `DISTRIBUTION.md`
- `GLOSSARY.md` — Terminology glossary
- `LICENSE` — Licence
- `README.md` — Repository overview and pack catalogue
- `RELEASE_NOTES_v1.0.0-rc.1.md`
- `RELEASE_READINESS.md`
- `ROADMAP.md` — Feature roadmap and backlog
- `VISION.md` — Project vision
<!-- END GENERATED: reconcile -->




## Command behaviour

`carl reconcile` refreshes the generated repository snapshot in `.github/carl/memory.md`. It is idempotent: when generated content is unchanged, it should perform no write. It does not modify `runtime.json`, harness adapter files, or other managed artefacts. It requires no network access and exits non-zero with an actionable message if `repo-map.json` or `memory.md` is missing.

`carl pack` (subcommands `list`, `show <pack-id>`, `select <pack-id>...`, `unselect <pack-id>...`, `profile`, `registry`, `install`, `update`, and `effective`, all leaf commands supporting `--json`) discovers, inspects, verifies, installs, selects, activates, and composes instruction packs. Pack IDs are `<category>/<name>`, derived from the canonical path `.github/instructions/<category>/<name>.instructions.md`. Discovery merges bundled, repository-local, registry-managed, and selected sources, sorted by pack ID — filesystem enumeration order is never policy order. Metadata is schema-versioned (`schemaVersion: 1`); the discovered set is validated (malformed/duplicate IDs, invalid versions, unknown schema versions, missing or cyclic dependencies, invalid owned artefacts, contradictory state, and registry-provenance drift) before output. JSON errors are structured payloads on stderr with non-zero exit (via `cmdutil.ExitError` with `SuppressPrefix`). Pack selection is persisted in `.github/carl/packs.json` (schema-versioned, deduplicated, sorted, user-owned); when absent, selection falls back to the legacy derivation from `runtime.json` managed artefacts, and malformed fallback state is an explicit error. Pack commands never write `runtime.json`. Composition metadata comes only from explicit pack file comment headers (`requires:`, `precedence-mode:`, `priority:`, `overrides:`) scanned in the first ten lines; absent headers default to no dependencies, additive mode, priority 0, no overrides; malformed headers are explicit errors. `carl pack effective` computes the effective pack set from active profile seeds plus transitive required dependencies (with explicit reasons), ordered by priority descending with pack-ID tie-breaks. Overrides apply only when explicitly declared and the target declares mode `overridable`; overridden packs stay visible in the set flagged `overriddenBy` for provenance, but their instruction definitions are not applied; conflicts (missing dependencies, override of a non-overridable pack, mutual overrides) exit non-zero. Existing discovery, selection, activation, and composition commands require no network access.

Pack Phase 3 policy profiles live in the schema-versioned, user-owned `.github/carl/profiles.json` artefact. The additive active seed set is organisation defaults plus repository defaults plus active-profile packs plus optional role/task overlays. Every profile reference must already be selected. `carl pack profile list`, `show`, `activate`, and `clear` inspect or update the explicit profile context; activation writes only `profiles.json`. `state.active` is profile-driven when the artefact exists; repositories without it retain selected-as-active behaviour as a compatibility fallback. Profile definitions, contexts, and references are strictly validated, and selection commands reject changes that would leave profile references unselected.

`.github/carl/profiles.example.json` is a bundled, managed, inactive reference
profile installed by `carl init`. Its ordinary schema-version 1 `default`
profile explicitly lists the complete shipped pack baseline and leaves
role/task context unset to preserve the role-neutral compatibility behaviour.
Only `.github/carl/profiles.json` is active profile state: users deliberately
copy and customise the example, every reference remains subject to ordinary
selection and validation, and the evaluator has no hidden default-profile
case. Future interactive profile creation or cloning must write this same
canonical model rather than introduce another state format.

Pack Phase 4 registries are configured explicitly in the schema-versioned, user-owned `.github/carl/registries.json` artefact; there is no built-in or inferred registry. Registry indexes may be HTTPS or repository-local and advertise releases by pack ID, canonical semantic version, relative artifact, and SHA-256. `carl pack registry list` is local-only; `registry search`, `install`, and `update` access only configured sources. Resolution chooses the highest semantic version or an exact requested version and rejects equal-version cross-registry ambiguity unless a registry is named. Install verifies bounded artifact bytes, SHA-256, declared pack version, composition metadata, required dependencies, ownership, symlink-safe paths, and the complete planned pack set before any write. It writes only canonical instruction-pack paths and deterministic provenance in `.github/carl/installed-packs.json`; it does not select or activate packs and never writes `runtime.json`. Updates use recorded registry provenance, reject local digest drift, changed registry location, and same-version digest mutation, and never downgrade. SHA-256 proves integrity against the configured index, not publisher identity or a signing trust chain.

Pack Phase 5 provides top-level `carl explain <pack-id>` and `carl trace` commands. They are deterministic, schema-versioned, local-only, network-free, and read-only views over existing validated discovery and `ComputeEffectiveSet` semantics. `explain` works for effective and inactive discoverable packs and reports source, repository-relative canonical definition, selection/profile/default/dependency activation, precedence, registry provenance, and override state. `trace` reports the complete effective pack evaluation in precedence order plus structured activation, dependency, ordering, pack-level constraint, resolved-override, and unresolved-conflict decisions; overridden entries remain visible but report `applied: false` and do not add constraints, while unresolved conflicts retain non-zero exits. The policy unit is an instruction pack: these commands do not parse natural-language prose into individual rules and do not expose prompts, hidden model reasoning, or chain-of-thought. Every explain/trace notice also states that instruction availability or loading does not prove model adherence. Explanation output is derived diagnostic evidence, not a canonical governance authority.

Pack Phase 6 extends `carl map` with repo-map schema version 1 and an additive cognitive `graph` while preserving all prior inventory fields consumed by `carl reconcile`. Stable repository-relative nodes cover the repository root, components, Go packages, entry points, workflows, governance artefacts, documentation, and instruction-pack definitions. Sorted `contains` edges describe structure; sorted `depends_on` edges come only from statically parsed repository-local Go imports and cite source-file evidence. Direct reverse dependencies populate bounded `change_impact`. Nodes include agent context, deterministic criticality heuristics, trust-boundary classifications, and component/package policy attachment points. Graph `coverage` records whether ownership, dependencies, data flows, trust boundaries, criticality, policy attachments, and impact evidence is derived, partial, or unavailable. The graph never guesses ownership or runtime data flow, never executes repository code, and does not claim policy activation; `carl trace` remains the active-policy provenance view. Graph output is derived orientation evidence, not canonical governance, ownership, runtime, or risk evidence.

`carl harness` manages and inspects harness adapters for AI coding agents. Its subcommands are `list`, `status`, and `sync`.

Harness adapters bridge cARL canonical artefacts to agent context injection mechanisms. cARL artefacts are the canonical source of truth; harness files are adapters, not authorities.

`carl harness list` shows all known adapters with support tier:

- `copilot`, `claude`, and `codex` — `production`;
- `cursor` and `antigravity` — `theoretical`: adapter is implemented and
  synchronised, but has not been validated end-to-end in its native harness.

`carl harness status` reports both detection-file presence and sync health by
comparing adapter file bytes against the canonical embedded source. It
aggregates presence as `detected`; neither `Present` nor `Synced` proves that a
native harness loaded or obeyed governance.

`carl harness sync [<harness-id>...]` generates adapter files for all adapters with defined adapter files, or only named harnesses when harness IDs are supplied. Syncing a shim harness writes both the shared loader (`.github/copilot-instructions.md`) and the harness-specific shim. The shared loader is written once even when syncing all harnesses. Adapter files are disposable and always overwritten. Sync works for all tiers regardless of support level. Sync is idempotent and does not require `carl init`.

`carl doctor` surfaces missing or drifted harness adapters as warning findings with `carl harness sync` remediation and reports the production harness IDs.

`carl status` includes a separate harness summary covering detected, missing,
drifted, healthy, and production harnesses without changing overall runtime
status semantics.

`carl version` reports three distinct version layers:

- CLI version (the executable);
- bundled runtime version and provenance (canonical payload embedded in the executable);
- repository runtime version from `.github/carl/runtime.json` when installed.

`carl version` and `carl version --components` show each harness support tier.
The component view compares bundled vs installed instruction packs and harness
shims, reporting bundled version, installed version, and sync/drift state
(`current`, `older`, `newer`, `missing`, `unknown`).

Detection files:

- Copilot: `.github/copilot-instructions.md`
- Claude: `CLAUDE.md`
- Codex: `AGENTS.md`
- Cursor: `.cursor/rules/carl.mdc`
- Antigravity: `.agents/rules/carl.md`

A shim harness is locally healthy only when both the shared loader
(`.github/copilot-instructions.md`) and the harness-specific shim are present
and synced. Production validation is a separate end-to-end evidence claim.

Maintainer field validation has proven the shared-loader shim workflow for
Copilot, Claude Code, and Codex. Claude's `CLAUDE.md` shim is sufficient for the
current production baseline; a `/carl` skill is not a prerequisite. Cursor and
Antigravity are the outstanding native-harness validation targets.

`harness.Command` accepts an `Artifacts` dependency using the same interface pattern as `repair`, `doctor`, and `status`.

The `repair` package exports `Inspect(rootDir, managed, arts)`, which returns separate missing and drifted slices while skipping protected paths. `repair.Command.detectDrift` delegates to `Inspect` internally.

`repair.CompareFile(rootDir, targetPath, canonicalPath, arts)` is the shared byte-comparison helper used by both runtime artefact inspection and harness adapter health checks.

The `repomap` package implements `carl map`. Its `Build(rootDir)` function derives the compatible inventory and schema-versioned cognitive graph from the filesystem using `filepath.WalkDir` and Go standard-library import parsing. It exports `RunInDir(rootDir)` for testability. `OutputFile` is `.github/carl/repo-map.json`.

The `convert` package implements `carl convert <source> [--dry-run | --apply]`, an AADLC-to-cARL governance migration command.

`convert` uses a converter framework: each source implements the `Converter` interface and is registered in the `converters` slice. A shared converter-agnostic engine performs duplicate detection, conflict detection, routing, and deterministic reporting. New converters can be added without changing the engine.

The AADLC converter discovers artefacts under `.aadlc/`, `.github/aadlc/`, `aadlc/`, and `AADLC.md`. It classifies Markdown and YAML bullet content by section-heading keywords into invariants, durable memory, and governance rules.

AADLC invariants are appended to `.github/carl/invariants.yml` using namespaced `aadlc-` IDs and `high` severity. Memory and governance entries go into a managed block in `.github/carl/memory.md`.

AADLC artefacts are never modified or deleted. Default mode is `--dry-run`. Conversion is idempotent and deterministic.

Malformed managed convert block markers cause conversion to fail before writing anything, rather than treating the block as absent and appending a second block.

## Core invariants

- cARL artefacts are the canonical source of durable governance truth.
- Harness-specific files are adapters/loaders, not authorities.
- Harness adapters must remain disposable and regenerable from canonical cARL assets.
- Instruction packs should remain modular and focused on a single concern.
- `.github/copilot-instructions.md` is both the Copilot harness entrypoint and the shared cARL adapter loader for all harness shims. It should remain a thin, procedural loader rather than duplicating the full operating model.
- The shared loader must derive effective instruction packs from canonical
  repository-local selection/profile/composition state. Presence is not
  selection, selection is not always activation, overridden definitions are
  not applied, filesystem order is not policy order, and invalid state fails
  closed without a load-all fallback.
- cARLv2 artefacts should reduce semantic rediscovery without becoming a per-turn session diary.
- Prompt-as-code should be used for substantial, long, nested, model-comparison, or boundary-sensitive agent tasks.
- Every implementation PR must make an explicit cARL/docs update decision before final response.

## Known sharp edges

- Instruction availability is not instruction adherence: a harness may load an instruction file, but different models vary in their ability to operationalise the full governance lifecycle without explicit checkpoints.
- Agents may over-anchor on completed PR contracts; distinguish durable invariants from historical PR constraints.
- Model availability and capability can vary; fallback models must preserve the active PR contract.
- Repeated corrective prompting is a failure signal; reset the session or switch model instead of continuing prompt ping-pong.
- Derived data support does not guarantee equivalent UX surface support.

## Canonical validation commands

- Build CLI: `go build ./cmd/carl`
- Run tests: `go test ./...`
- Build tagged release: `go build -ldflags "-X main.cliVersion=<tag> -X main.bundledRuntimeVersion=<runtime-version> -X main.bundledRuntimeSource=goldjg/cARL -X main.bundledRuntimeTag=<tag> -X main.bundledRuntimeCommit=$(git rev-parse HEAD)" ./cmd/carl`
- Validate GoReleaser config: `goreleaser check`
- GoReleaser snapshot dry-run (no publish): `goreleaser release --snapshot --skip=publish --clean`

## Current operating assumptions

Model availability and capability are not stable invariants. The PR contract remains the source of truth across model fallback.

Harness behaviour is not equivalent to model compliance. A harness may place instructions in context, but cARL must still make the required governance lifecycle explicit enough for weaker or cheaper models to follow.

The active authority order is:

1. current repository state;
2. active user instruction within approved scope;
3. `.github/carl/current-pr-contract.md`;
4. `.github/carl/invariants.yml`;
5. `.github/carl/trust-boundaries.md`;
6. `.github/carl/memory.md`;
7. relevant `.github/instructions/` packs;
8. harness adapter files;
9. stale prompt/session memory.

## Open questions

<!-- Populate with unresolved questions that should persist into future work. -->

## Last updated
2026-07-26 by v1.0.0-rc.1 release-readiness reconciliation
