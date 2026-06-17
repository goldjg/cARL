# cARL CLI v1 Foundation Plan

## Task

Implement the first working version of the cARL CLI in Go.

This PR should create a minimal but production-oriented foundation.

## Goal

Implement three commands:

``` bash
carl init
carl repair
carl version
```

The implementation should be designed so additional commands can be
added later without architectural rewrites.

------------------------------------------------------------------------

## Design Principles

cARL is a governance runtime for coding agents.

The CLI manages repository-local cARL installations.

The CLI is responsible for:

-   installing managed cARL artefacts
-   repairing managed artefacts
-   reporting installed runtime state

The CLI is NOT responsible for:

-   executing coding agents
-   orchestrating agent workflows
-   enforcing invariants itself
-   parsing repository code

It only manages the cARL runtime artefacts.

------------------------------------------------------------------------

## Repository Layout

Create a structure similar to:

``` text
cmd/carl/
internal/runtime/
internal/install/
internal/repair/
internal/version/
internal/manifest/
embedded/
```

Avoid premature complexity.

------------------------------------------------------------------------

## Runtime Manifest

Introduce a runtime manifest file:

``` text
.github/carl/runtime.json
```

Example:

``` json
{
  "runtimeVersion": "1.0.0",
  "source": "goldjg/cARL",
  "sourceTag": "v1.0.0",
  "sourceCommit": "abcdef123456",
  "installedAt": "2026-06-17T00:00:00Z",
  "managedArtifacts": [
    ".github/carl/memory.md",
    ".github/carl/invariants.yml"
  ]
}
```

This becomes the source of truth.

Do not infer state from filesystem scans.

------------------------------------------------------------------------

## Command: init

Behaviour:

``` bash
carl init
```

Rules:

-   fail if any managed cARL artefact already exists
-   fail if runtime.json already exists
-   explain which files caused the failure

If no installation exists:

-   create required directories
-   install embedded artefacts
-   create runtime.json

Initial implementation should use embedded files.

Do NOT download from GitHub.

Future releases may support remote updates.

------------------------------------------------------------------------

## Embedded Assets

Use Go embed.

The CLI release must be self-contained.

All required cARL artefacts should be embedded in the binary.

No network dependency.

------------------------------------------------------------------------

## Command: repair

Behaviour:

``` bash
carl repair
```

Repair only managed artefacts.

Rules:

-   do not modify memory.md
-   do not modify runtime.json
-   do not modify repository-specific artefacts

Repair only canonical runtime files.

Examples:

``` text
copilot-instructions.md
instruction packs
templates
tool policy
```

Compare installed files against embedded runtime files.

If drift is detected:

-   report drift
-   restore embedded version

Provide dry-run output first.

Example:

``` text
Drift detected:
  .github/instructions/core/carl.instructions.md

Repairing...
Done.
```

------------------------------------------------------------------------

## Command: version

Behaviour:

``` bash
carl version
```

Output:

``` text
CLI Version:      1.0.0
Runtime Version:  1.0.0
Source:           goldjg/cARL
Tag:              v1.0.0
Commit:           abcdef123456

Installed Packs:
  core/carl
  core/memory-cache
  core/pr-contract

Runtime Status:
  Healthy
```

If no runtime is installed:

``` text
No cARL runtime installed.
```

------------------------------------------------------------------------

## Versioning Strategy

Use semantic version tags.

Do NOT use commit hashes as primary version identifiers.

Store both:

``` json
{
  "sourceTag": "v1.0.0",
  "sourceCommit": "abcdef"
}
```

Tags are authoritative.

Commits are informational.

------------------------------------------------------------------------

## Out of Scope

Do NOT implement:

-   upgrade
-   validate
-   doctor
-   remote downloads
-   GitHub API integration
-   pack installation
-   pack removal
-   invariant enforcement
-   repository scanning beyond runtime discovery

------------------------------------------------------------------------

## Acceptance Criteria

1.  `carl init` installs a runtime from embedded assets.
2.  Re-running `carl init` fails safely.
3.  `carl repair` restores modified managed artefacts.
4.  `memory.md` is never overwritten.
5.  `runtime.json` is never overwritten.
6.  `carl version` reports runtime state correctly.
7.  All runtime state comes from `runtime.json`.
8.  No network access is required.
9.  Architecture supports future commands without redesign.

------------------------------------------------------------------------

## Additional Design Note

Use:

``` text
.github/carl/runtime.json
```

as the authoritative runtime state file.

Future capabilities such as upgrades, validation, drift analysis, and
cARLy Gates should rely on this file as the canonical source of runtime
metadata.
