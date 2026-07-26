<!-- version: 1.0.0 -->
# Enterprise Profile Adherence Scenarios

## Status

Completed

## Goal

Create executable policy-profile scenarios for fictional financial-services,
health, and manufacturing enterprises so agent adherence can be exercised
across design, architecture, code, implementation, and review work.

## Affected files

- `.github/carl/current-pr-contract.md`
- `.github/carl/packs.json`
- `.github/carl/profiles.json`
- `.github/carl/enterprise-profiles.md`
- `.github/instructions/enterprise/*.instructions.md`

## Step-by-step changes

1. Add reusable discipline packs for design, architecture, code,
   implementation, and review.
2. Add fictional-company scenario packs for financial services, health, and
   manufacturing, including strict and bounded lightweight/balanced variants.
3. Select the existing built-in packs and the new repository-local packs in
   user-owned `packs.json`, without changing any built-in definition.
4. Define six profiles with role and task overlays in user-owned
   `profiles.json`, with one safe non-production implementation context active.
5. Document activation examples and the expected adherence evidence.
6. Validate schema, discovery, profile references, effective composition, and
   profile switching with the cARL CLI.

## Test strategy

- `go run ./cmd/carl pack list --json`
- `go run ./cmd/carl pack profile list --json`
- `go run ./cmd/carl pack profile show <profile-id> --json` for every profile
- `go run ./cmd/carl pack effective --json` for representative strict,
  balanced, and lightweight contexts
- `go test ./...`
- `go vet ./...`
- `go build ./cmd/carl`
- `git diff --check`
- Verify built-in pack files are unchanged.

## Risks

- A lightweight profile could be mistaken for production authority. Every
  lightweight pack therefore has an explicit non-production boundary and
  escalation rule.
- Profile references can become invalid if selected packs drift. CLI
  validation must cover every inactive profile as well as the active one.
- Activating a test profile changes subsequent agent hydration. The committed
  active context must be intentional and documented.

## cARL/docs update expectation

Expected. The repository gains durable user-owned selection/profile state and
repository-local enterprise test packs, so a dedicated guide should record
their purpose, activation, and boundaries without duplicating mutable active
state into durable memory.
