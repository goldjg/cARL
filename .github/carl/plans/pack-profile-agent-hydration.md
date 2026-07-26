<!-- version: 1.1.0 -->
# Pack/Profile Agent Hydration Plan

## Status

Completed

## Goal

Make the shared cARL loader a complete, CLI-optional contract for deriving and
applying only effective, non-overridden instruction packs from canonical
repository-local policy state.

## Affected surfaces

- Shared loader and embedded mirror
- Existing core cARL instruction pack and embedded mirror
- Focused source/embedded parity and loader-invariant tests
- Pack selection/explanation code only for discovered fail-closed or
  overridden-application inconsistencies
- Durable memory, trust boundaries, and public documentation where stable
  wording changes
- Bundled inactive default-profile example and exact baseline tests

## Implementation

1. Add a normative hydration section to the shared loader before the general
   lifecycle, with explicit state distinctions and fail-closed evaluation.
2. Put the same semantic contract in the existing core cARL pack. Do not add a
   dedicated hydration pack because the loader must bootstrap evaluation
   before any pack can be known effective and the existing core pack already
   owns cARL lifecycle governance.
3. Reconcile control-plane edge cases found during implementation review:
   malformed legacy manifests must not silently select nothing, and
   overridden policy definitions must be reported but not marked applied.
4. Add tests that map directly to the current PR contract assertions,
   including byte parity and shim routing.
5. Update durable cARL/public documentation only where current language is
   contradictory or incomplete.
6. Ship `.github/carl/profiles.example.json` as an inactive schema-version 1
   baseline installed by `carl init`; require deliberate adoption as
   `profiles.json`, with no evaluator special case.
7. Test that its explicit 24-pack, role-neutral context resolves identically
   to the profile-absent legacy baseline and preserves ordinary composition.

## Validation

- Run focused package tests while iterating.
- Run full test, vet, and build commands using the installed Go 1.26.4
  toolchain, which satisfies the repository's Go 1.24 minimum.
- Compare canonical and embedded assets byte-for-byte.
- Search all governance and documentation text for contradictory hydration
  language, mandatory CLI wording, load-all fallbacks, filesystem ordering,
  and overridden packs described as applied.

## Risks

- Prose can drift from evaluator behaviour; mitigate with exact invariant
  tests and durable wording anchored to canonical fields/metadata.
- Retaining overridden entries for provenance can be confused with applying
  them; make diagnostic visibility versus instruction application explicit.
- Harness loading and model adherence remain externally unproven; preserve
  that evidence limitation in loader, trace documentation, and final report.
