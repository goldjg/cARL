<!-- version: 1.0.0 -->
# cARL v1.0.0-rc.1

This is a release candidate, not the final stable `v1.0.0` release. It is
intended for upgrade, installation, packaging, and coding-agent validation
before the v1 compatibility line is declared final.

## What cARL 1.0 is

cARL is a repository-local governance and instruction runtime for AI coding
agents. It commits durable memory, scoped PR contracts, trust boundaries,
tool policy, instruction packs, and harness adapters alongside the code they
govern. The CLI installs and manages those artefacts without adding a runtime
service or mandatory network dependency.

## Installation

Release archives and native Linux packages will be attached to the
[`v1.0.0-rc.1` GitHub prerelease](https://github.com/goldjg/cARL/releases/tag/v1.0.0-rc.1).
The release workflow also prepares the Homebrew cask and, when its configured
publisher token is available, a WinGet update submission.

After installing the CLI:

```sh
carl version
carl init
carl doctor
carl status
```

Use `carl init --adopt` instead of ordinary `init` when valid cARL artefacts
already exist but `.github/carl/runtime.json` is absent.

## Shared governance across supported coding agents

GitHub Copilot, Claude Code, and Codex use production-validated adapters that
route to the same `.github/copilot-instructions.md` loader. Cursor and
Antigravity have implemented detection, generation, drift, and sync support,
but remain theoretical until native-harness execution is validated.

Adapter presence and byte parity are local filesystem evidence. They do not
prove that a model followed every instruction.

## Lifecycle and repository adoption

- `carl init` performs a collision-safe fresh installation.
- `carl init --adopt` preserves existing files byte-for-byte, installs only
  missing bundled artefacts, and writes the runtime manifest last.
- `carl doctor`, `status`, and `repair` distinguish missing, drifted,
  repairable, and protected state.
- Runtime-owned assets can be repaired explicitly; user-owned policy,
  provenance, and protected memory are not silently replaced.

## Packs and policy composition

Packs have distinct present, selected, active, effective, and overridden
states. Selection is repository policy; profiles determine active seeds;
dependencies expand transitively; precedence is explicit and deterministic;
and overrides require explicit authority from both packs. Invalid or
conflicting state fails closed.

## Profiles, roles, and tasks

Schema-versioned profiles compose organisation defaults, repository defaults,
a named profile, and optional role/task overlays. The bundled inactive
`profiles.example.json` reproduces the profile-absent 24-pack baseline when
copied deliberately to `profiles.json`.

The fictional enterprise examples demonstrate a fail-safe adoption sequence:
copy the default-only bootstrap, select the 11 example packs, copy the full
catalogue, and then activate a profile explicitly. They are examples, not
certifications.

## Registries and verified installation

Registries are explicit and opt-in. HTTPS and repository-local indexes can
advertise versioned pack artefacts with SHA-256 digests. Installation validates
the complete dependency transaction before writing and records deterministic
provenance. A checksum proves integrity against the configured index; it does
not authenticate the publisher or establish a signing trust root.

## Explain and trace

`carl explain` and `carl trace` expose deterministic pack-level policy
provenance: selection, profile/default activation, dependencies, order,
constraints, overrides, and conflicts. They are read-only and network-free.
They do not expose prompts, hidden reasoning, or chain-of-thought.

## Cognitive repository graph

`carl map` produces a schema-versioned repository inventory and evidence-scoped
graph with stable nodes, structural and static Go-import edges, direct change
impact, trust-boundary classifications, policy attachment points, and explicit
coverage limitations. It does not infer owners or runtime data flow.

## Repair, reconcile, doctor, and status

- `doctor` reports actionable findings without changing files.
- `repair` restores only declared repairable runtime-owned assets.
- `map` refreshes repository orientation evidence deterministically.
- `reconcile` updates only the generated memory snapshot and is idempotent.
- `status` separates runtime and harness health.

## Version reporting

`carl version` reports the CLI executable, its bundled runtime payload, and the
repository runtime separately. CLI identity and bundled provenance are build
metadata and do not depend on a repository `runtime.json`.

## Distribution and notarisation

GoReleaser builds Linux amd64/arm64, macOS amd64/arm64, and Windows amd64
archives plus deb, rpm, and apk packages and SHA-256 checksums. macOS binaries
are configured for Developer ID signing, hardened runtime, and App Store
Connect notarisation. Release runs are serialised, fail closed on required
publisher credentials, retry only Apple-context notarisation throttling, and
verify all required GitHub assets before downstream publication.

## Upgrade from v0.4.3

The tested upgrade keeps existing artefacts under explicit control:

1. run the RC binary with `carl version`, `carl status`, and `carl doctor`;
2. back up and move `.github/carl/runtime.json` outside the managed runtime
   path;
3. run `carl init --adopt`;
4. review `carl doctor`;
5. run `carl repair` only if current canonical runtime-owned content is wanted;
6. run `carl status`, `carl harness status`, `carl pack effective`, `carl map`,
   and `carl reconcile`.

Adoption and repair preserve repository-specific memory and other user-owned
files. The separately requested `reconcile` step may update only memory's
declared generated snapshot section.

## Compatibility promise

The `1.x` line preserves documented commands and exits, schema-versioned JSON,
runtime/profile/pack/registry/provenance formats, map schema, pack composition,
ownership boundaries, and lifecycle semantics. Additive JSON fields are
compatible; field removal or meaning changes require a schema transition;
intentional public breaking changes require a new major version. Human-readable
formatting and implementation details are not byte-for-byte stable. See
[COMPATIBILITY.md](COMPATIBILITY.md).

## Known limitations

- The RC validation executed Windows amd64 binaries. GoReleaser cross-built
  every configured target, but Linux and macOS binaries were not executed on
  this host.
- RC-specific signing, notarisation, Gatekeeper, GitHub prerelease, Homebrew,
  and WinGet behaviour still requires evidence from the real tagged workflow.
- Cursor and Antigravity remain theoretical harnesses.
- Registry SHA-256 does not authenticate publishers.
- Policy availability and trace output do not prove model adherence.

## Release-candidate validation requested

Please test:

- fresh installation and `init --adopt` in disposable repositories;
- upgrade from v0.4.3 with user-owned files present;
- Linux and macOS archive/package installation;
- macOS Gatekeeper acceptance of the notarised binaries;
- GitHub Copilot, Claude Code, and Codex loading through the shared adapter;
- profile, role, task, and enterprise-example composition;
- Homebrew and WinGet installation once their channel updates are available.

Report the platform, architecture, command, exit code, and relevant
schema-versioned output for any failure.
