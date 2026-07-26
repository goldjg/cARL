<!-- version: 1.0.0 -->
# cARL 1.x Compatibility Policy

This policy defines the public compatibility promise for the cARL `1.x`
series. It covers documented product contracts, not every implementation
detail.

## Stable throughout 1.x

The following interfaces are stable for compatible `1.x` releases:

- command names and documented command semantics;
- documented exit behaviour;
- schema-versioned JSON command output;
- `.github/carl/runtime.json`;
- `.github/carl/packs.json`;
- `.github/carl/profiles.json`;
- `.github/carl/registries.json`;
- installed-pack provenance in `.github/carl/installed-packs.json`;
- repository-map schema;
- instruction-pack metadata headers;
- pack selection, activation, dependency, precedence, override, and conflict
  semantics;
- runtime-owned versus user-owned artefact boundaries;
- the documented lifecycle behaviour of `init`, `init --adopt`, `doctor`,
  `repair`, `status`, `map`, `reconcile`, and harness management.

Stable means that a compatible `1.x` release will preserve the documented
meaning and safety properties of these interfaces. It does not mean that every
implementation, diagnostic sentence, or line of output is frozen.

## Compatible evolution

- Additive fields may be added to schema-versioned JSON unless a specific
  schema contract explicitly forbids them. Consumers should ignore fields
  they do not understand.
- Removing a field or changing the meaning of an existing field requires a
  schema-version transition.
- User-owned policy and provenance files are never silently replaced.
- Runtime repair remains limited to declared repairable runtime-owned assets.
  Protected memory and runtime-manifest state retain their documented
  boundaries.
- Compatible bundled policy-pack revisions may clarify or strengthen guidance
  without changing pack identity or composition semantics.
- Intentional breaking changes to stable public contracts require a new major
  cARL version.

## Not byte-for-byte stable

The following may evolve compatibly within `1.x`:

- human-readable CLI formatting;
- wording or ordering that is not declared machine-readable;
- documentation prose and examples;
- bundled policy-pack text where a compatible revision is possible;
- internal package structure, helper APIs, algorithms, and generated build
  details that are not documented public contracts;
- experimental behaviour and explicitly theoretical harness behaviour.

Automation should consume documented JSON output and schema versions rather
than parse human-readable tables or prose.

## Ownership boundary

Runtime-owned artefacts may be compared, synchronised, or repaired only as
documented. User-owned files such as `packs.json`, `profiles.json`,
`registries.json`, installed-pack provenance, and repository-specific memory
must remain under explicit user control. `init --adopt` preserves existing
files and installs only missing bundled artefacts before creating
`runtime.json` last.

## Evidence boundary

Compatibility of cARL files and commands does not prove that a coding agent
loaded, understood, or obeyed every instruction. Harness support tiers and
policy provenance describe observable integration evidence, not hidden model
reasoning or perfect instruction compliance.
