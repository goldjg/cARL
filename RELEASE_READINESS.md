<!-- version: 1.0.1 -->
# v1.0.0-rc.1 Release Readiness

## Decision

The repository is ready to create the `v1.0.0-rc.1` tag after this readiness
change is merged and its required checks pass. No tag or release was created
during this work.

Final `v1.0.0` remains gated on RC execution evidence: the real tag workflow,
signed/notarised macOS artefacts and Gatekeeper behaviour, prerelease asset
publication, Homebrew/WinGet channel updates, and native smoke tests outside
Windows amd64.

## Baseline

- Authoritative `main`: `5f24ebc7050e22d358a2e7351d2ea92449797786`.
- Host: Windows amd64.
- Repository toolchain: Go 1.24.0, matching `go.mod`.
- Release tool: GoReleaser v2.17.0, matching the workflows.
- Authoritative upgrade source: annotated tag `v0.4.3` at
  `3f879dceb6b3d43dab86f77ae71485e1dabfcdc2`.

## Compatibility promise

The `1.x` line keeps documented command names, semantics, and exits;
schema-versioned JSON and repository formats; pack metadata, selection,
activation, dependency, precedence, override, and conflict semantics;
runtime/user ownership boundaries; and documented lifecycle behaviour stable.
Additive JSON fields remain compatible, while removal or meaning changes
require a schema transition. Intentional breaks to these public contracts
require a major version. Human formatting, prose, compatible bundled policy
revisions, internals, and theoretical harness behaviour are not byte-for-byte
promises. The durable contract is [COMPATIBILITY.md](COMPATIBILITY.md).

## Release-equivalent host binary

The Windows amd64 RC binary was built with `CGO_ENABLED=0`, `-s -w`, and the
same metadata variables as `.goreleaser.yaml`.

| Field | Value |
|---|---|
| CLI version | `1.0.0-rc.1` |
| Bundled runtime | `1.0.0-rc.1` |
| Source | `goldjg/cARL` |
| Tag | `v1.0.0-rc.1` |
| Commit | `5f24ebc7050e22d358a2e7351d2ea92449797786` |
| SHA-256 | `506C85420AC3D09F639FCCE7E1E2254D7DD6B109BEADCC485F8C809237FD621A` |

Outside an initialised repository, `carl version` and
`carl version --components` reported the CLI and bundled runtime and handled
the absent repository runtime cleanly. Inside a current installation, CLI,
bundled runtime, repository runtime, and comparison status were distinct.

## Lifecycle matrix

| Scenario | Result | Evidence |
|---|---|---|
| Fresh install | PASS | `init`, `doctor`, `status`, harness list/status, pack list/effective, trace, map, and reconcile passed. The runtime installed 38 managed artefacts. |
| Reconcile idempotence | PASS | The second reconcile reported `No reconciliation needed`; the memory SHA-256 was unchanged. |
| Adoption recovery | PASS | Ordinary `init` stopped with adoption guidance. `init --adopt` preserved existing cARL files, installed one missing bundled asset, and created `runtime.json` after it. |
| Drift and repair | PASS | Doctor identified the deliberately drifted shared loader. Repair restored only the repairable asset; protected memory and user-owned profile state retained their hashes. |
| v0.4.3 upgrade | PASS | An executable built from the authoritative tag initialised a 37-artefact runtime. The RC reported `Upgrade available`, managed it safely, and the adopt/repair upgrade reached a healthy 38-artefact RC runtime without changing user-owned sentinel bytes. |
| Profile absent | PASS | 24 selected, 24 active, 24 effective; exact lexicographic IDs/order; zero conflicts; trace reported 24 policies from legacy selection. |
| Default profile | PASS | Copying `profiles.example.json` kept the exact 24-pack set/order, unset role/task, and zero conflicts. Deleting the user-owned profile restored profile-absent behaviour exactly. |
| Enterprise examples | PASS | Exact transitions are recorded below; every valid stage had zero conflicts. |
| Harness management | PASS | All five adapters detected, drifted when modified, and returned to byte-identical canonical state after sync. Each non-Copilot adapter remained a thin route to the shared loader. |
| Non-trivial repository map | PASS | Schema 1; 97 unique sorted nodes; 128 valid edges; no rooted/traversing evidence paths; repeated output hash was identical. |
| Malformed map input | PASS | Reconcile exited 1 and left memory unchanged. |
| Symlink input | LIMITED | Direct host fixture creation required Windows administrator privilege. The dedicated `internal/repomap` symlink case remains covered by the passing suite but can skip when the host cannot create symlinks. |

### Enterprise adoption transitions

| Stage | Selected | Active | Effective | Enterprise selected/active/effective |
|---|---:|---:|---:|---:|
| Initial | 24 | 24 | 24 | 0 / 0 / 0 |
| Bootstrap copied | 24 | 24 | 24 | 0 / 0 / 0 |
| Enterprise selected | 35 | 24 | 24 | 11 / 0 / 0 |
| Full catalogue copied | 35 | 24 | 24 | 11 / 0 / 0 |
| `brindleforge-process-pilot / implementer / sandbox-build` | 35 | 5 | 9 | 11 / 2 / 2 |
| Default restored | 35 | 24 | 24 | 11 / 0 / 0 |

The representative effective order was:

1. `enterprise/discipline-implementation`
2. `enterprise/brindleforge-process-pilot`
3. `core/carl`
4. `core/cognition-governance`
5. `core/pr-contract`
6. `core/tool-permission-tiers`
7. `languages/go`
8. `languages/python`
9. `platform/docker`

## Upgrade evidence

The v0.4.3 binary was built from the annotated tag using that tag's own
GoReleaser ldflags:

- CLI/runtime: `0.4.3`;
- source/tag: `goldjg/cARL` / `v0.4.3`;
- commit: `3f879dceb6b3d43dab86f77ae71485e1dabfcdc2`;
- Windows binary SHA-256:
  `A9916AFC7127201AABE37B23B2D5CD67A208D42EA3A8B4F9D77E2A1925E6D9B2`.

Before adoption, the RC correctly separated its own `1.0.0-rc.1` metadata from
the repository's `0.4.3` runtime and reported `Upgrade available`. The tested
upgrade explicitly moved the old manifest out of the active path, used
`init --adopt`, reviewed doctor output, and repaired current runtime-owned
content. Memory and an independent user-owned sentinel remained preserved.
The exact documented order was also repeated: inspect with
`version`/`status`/`doctor`, move the old manifest, adopt, inspect, repair,
validate harness and pack state, map, and reconcile twice. Adoption and repair
left memory byte-identical; the explicit first reconcile updated its generated
snapshot and the second was a no-op.

## Repository validation

| Check | Result |
|---|---|
| Canonical Git-blob `gofmt` verification | PASS — 0 files |
| `go test -count=1 ./...` with Go 1.24.0 | PASS |
| `go vet ./...` with Go 1.24.0 | PASS |
| `go build ./cmd/carl` with Go 1.24.0 | PASS |
| `go test -race -count=1 ./...` | PASS — native Ubuntu Linux amd64, Go 1.24.0, GCC 13.3.0, `CGO_ENABLED=1` |
| `git diff --check` | PASS |
| GoReleaser v2.17.0 `check` | PASS |
| GoReleaser v2.17.0 snapshot with Go 1.24.0 | PASS |
| Retry wrapper `bash -n` and regression suite | PASS |
| Release workflow YAML parsing with PyYAML 6.0.3 | PASS |
| Canonical/embedded byte parity | PASS — 38/38, zero drift |
| Harness shim routing | PASS |
| Generated-map consistency | PASS — map refreshed, reconcile updated once, second reconcile was a no-op |

The race suite was completed after merge on a dedicated native Ubuntu Linux
amd64 VM against the exact tagged commit
`4f6e30bbf3fd4de230ee60524d266a9533e6a224`. The official Go 1.24.0 archive
was verified against its published SHA-256 before use. The command exited zero;
all packages passed and the race detector reported no data races.

The GoReleaser snapshot cross-built Linux amd64/arm64, macOS amd64/arm64, and
Windows amd64; created five archives, six native Linux packages, checksums,
and a Homebrew cask. Only the Windows amd64 binaries were executed locally.

## Release pipeline classification

### Statically validated

- tag push and manual existing-tag resolution to the exact commit;
- repository-wide release serialization without cancellation;
- Linux amd64/arm64, macOS amd64/arm64, and Windows amd64 target matrix;
- CLI and bundled-runtime ldflags;
- five archives, six Linux packages, and SHA-256 checksums;
- required Apple and Homebrew credential preflight before publication;
- Developer ID signing, hardened runtime, and App Store Connect notarisation
  configuration;
- one outer retry only for Apple-context notarisation rate limits;
- preservation of the underlying GoReleaser failure exit status;
- existing-tag release-note preservation and same-name asset replacement;
- required asset assertion before WinGet;
- Homebrew cask publisher and optional WinGet update submission;
- no downstream WinGet publication after a failed release job.

### Previously production-proven

The public `v0.4.3` workflow run
[`30051807280`](https://github.com/goldjg/cARL/actions/runs/30051807280)
completed successfully. Its GoReleaser job, Apple credential gate,
GoReleaser publish step, 12-asset assertion, and WinGet submission all passed.
The public GitHub Release contains the expected five archives, six native
packages, and checksum file.

### Requires v1.0.0-rc.1 execution evidence

- RC prerelease classification and final asset set on GitHub;
- actual RC Developer ID signature, notarisation ticket, and Gatekeeper result;
- Homebrew cask update/install for the RC;
- WinGet acceptance and installation for the RC;
- Linux and macOS execution smoke tests;
- safe rerun against an actually partial RC publication, if that recovery path
  is needed.

## Findings

### BLOCKER — fixed in this change

1. Existing-tag reruns could not replace already-uploaded same-name assets.
   GoReleaser now preserves existing release notes and replaces those assets.
2. The retry wrapper treated a generic GitHub/Homebrew `429` or `RATE_LIMIT`
   as Apple notarisation throttling. Retry matching now requires
   Apple/notarisation context, with negative regression tests.
3. A missing required Homebrew publisher token could fail late after GitHub
   publication began. The workflow now checks it with the Apple credentials
   before GoReleaser runs.

### RC EXIT CRITERION

- Obtain successful real-tag evidence for signing, notarisation, Gatekeeper,
  GitHub prerelease publication, Homebrew, and WinGet.
- Execute smoke tests on Linux and macOS release archives before final
  `v1.0.0`.

### FOLLOW-UP

- Native-harness validation for Cursor and Antigravity.
- Profile TUI/clone UX, Pack Phase 7 publishing, richer adherence automation,
  additional sectors, and additional language packs.

### NON-ISSUE

- The profile-absent compatibility fallback and adopted default profile are
  exactly equivalent.
- Enterprise examples are present but inactive until deliberate adoption.
- Registry checksums are correctly limited to configured-index integrity.
- Human-readable formatting remains intentionally non-byte-stable; JSON is the
  automation contract.
- Local adapter sync does not promote a harness support tier.

## Documentation changes

- `COMPATIBILITY.md` defines the durable `1.x` public boundary.
- `RELEASE_NOTES_v1.0.0-rc.1.md` provides capability-oriented GitHub
  prerelease notes.
- README and CLI now lead with current install/adopt/upgrade usage, the pack
  state model, profile and enterprise adoption, support tiers, evidence
  limitations, and release troubleshooting.
- Architecture, distribution, and roadmap now agree on current runtime
  inventory, compatibility, release targets, credential gates, Apple-only
  retry behaviour, and same-tag recovery.
- The active contract, implementation plan, durable memory, embedded memory,
  and generated repository map were reconciled.

## Delivery

- Branch: `agent/prepare-v1.0.0-rc.1`
- Initial readiness commit: `95c5575`
- Draft PR: [#45](https://github.com/goldjg/cARL/pull/45)

No tag or release was created. No Homebrew or WinGet publication was
triggered.

## Commands

The validation used the release-equivalent binaries in isolated temporary Git
repositories. Principal commands:

```text
git status -sb
git log -5 --decorate --oneline
git show -s v0.4.3
go build -ldflags "<GoReleaser-equivalent RC metadata>" ./cmd/carl
carl version
carl version --components
carl init
carl init --adopt
carl doctor
carl repair
carl status
carl harness list
carl harness status
carl harness sync [copilot|claude|codex|cursor|antigravity]
carl pack list --json
carl pack effective --json
carl pack profile list --json
carl pack profile show <id> --json
carl pack profile activate <profile> [--role <role>] [--task <task>]
carl pack select <11 enterprise pack IDs>
carl trace --json
carl map
carl reconcile
gofmt -l <canonical Git-blob Go files>
go test -count=1 ./...
go test -race -count=1 ./...
go vet ./...
go build ./cmd/carl
git diff --check
goreleaser check
goreleaser release --snapshot --skip=publish --clean
bash -n scripts/release-with-retry.sh
bash -n scripts/release-with-retry.test.sh
bash scripts/release-with-retry.test.sh
python -c "<PyYAML parse of release workflows and GoReleaser config>"
```

Read-only public GitHub API calls verified the latest `main` commit, the
authoritative v0.4.3 release, its 12 assets, and the successful release
workflow/job/step conclusions. Those evidence lookups changed no external
state.
