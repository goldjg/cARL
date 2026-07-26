# Opt-In Enterprise Pack and Profile Examples

> **Invariant: “Installing or merging these examples does not change agent behaviour until a user explicitly activates or adopts one.”**

The files under `.github/instructions/enterprise/` and
`.github/carl/profiles.enterprise.example.json` are fictional engineering
fixtures. They demonstrate cARL pack/profile composition without changing this
repository's default policy state.

They are not legal, regulatory, medical, security, cybersecurity, privacy, or
functional-safety certifications.

## Why the examples are inactive

Only `.github/carl/packs.json` is explicit selection state and only
`.github/carl/profiles.json` is active profile state. This example set commits
neither file.

The repository therefore keeps the same compatibility behavior as `main`:

- selection is derived from the existing `runtime.json` managed artifacts;
- all 24 selected built-in packs are active because no `profiles.json` exists;
- the enterprise pack definitions are present and discoverable but inactive;
- no fictional profile, role, or task affects an agent by default.

The dedicated fixture filename
`.github/carl/profiles.enterprise.example.json` has no evaluator semantics. It
is ordinary schema-version 1 data that is read only after a user deliberately
copies it to `.github/carl/profiles.json`.

## Pack state model

The fictional scenarios illustrate four distinct states:

| State | Meaning | Enterprise example before adoption |
|---|---|---|
| Present | A definition exists and can be discovered | All 11 enterprise packs are present |
| Selected | The pack ID is recorded in `packs.json`, or selected by the legacy compatibility derivation | Enterprise packs are not selected |
| Active | The pack is a default, active profile, active role/task overlay, or selected-as-active compatibility seed | No enterprise pack is active |
| Effective | The active seed or a required dependency survived validation and composition | No enterprise pack is effective |

Presence never implies selection. Selection does not imply activation when a
profile file exists. Activation does not bypass dependency, precedence,
override, or conflict validation.

## Profile composition model

When a copied profile fixture is active, cARL composes policy additively:

1. organisation defaults;
2. repository defaults;
3. the active profile's base packs;
4. the active role overlay;
5. the active task overlay;
6. every transitive required dependency.

The enterprise fixture leaves organisation and repository defaults empty so
the active `default` profile is the complete, explicit baseline. The default
profile lists the same 24 built-in packs, in the same order, as the repository
compatibility baseline.

The fictional profiles each contribute one company/scenario pack. Their role
overlays contribute one reusable discipline pack:

- `designer` → `enterprise/discipline-design`
- `architect` → `enterprise/discipline-architecture`
- `coder` → `enterprise/discipline-code`
- `implementer` → `enterprise/discipline-implementation`
- `reviewer` → `enterprise/discipline-review`

Task overlays add the language, platform, cloud, identity, or security packs
needed by that test context.

## Dependencies, precedence, and overrides

Dependencies declared with `requires:` are expanded transitively even when the
dependency was not itself an active profile seed. This is why a strict
scenario can activate its core baseline, security, identity, dependency,
PR-contract, and tool-permission requirements through one company pack.

Precedence is deterministic:

- explicit priority descending;
- then pack ID lexicographically.

The example priorities make company and discipline guidance appear before the
priority-`0` built-ins, without replacing them.

No enterprise example declares `overrides:`. An override would only be valid
when explicitly declared and the target declared
`precedence-mode: overridable`; existing built-in controls are not weakened by
these fixtures.

## Scenario catalogue

| Profile | Sector and posture | Intended boundary |
|---|---|---|
| `argentum-regulated-payments` | Argentum Financial Group; strict | Payments, ledgers, fraud, identity, and regulated customer-data systems |
| `argentum-innovation-lab` | Argentum Financial Group; bounded-lightweight | Synthetic-data discovery in disposable, isolated non-production sandboxes |
| `willowmere-clinical-records` | Willowmere Health Network; strict | Clinical records, patient identity, care workflows, and safety-sensitive systems |
| `willowmere-service-design` | Willowmere Health Network; balanced | Accessibility, workflow research, content design, and synthetic prototypes |
| `brindleforge-connected-factory` | Brindleforge Manufacturing; strict | OT, robotics, industrial control, safety interlocks, and plant deployment |
| `brindleforge-process-pilot` | Brindleforge Manufacturing; bounded-lightweight | Simulation, synthetic analytics, and offline disposable process pilots |

Strict examples add explicit controls for regulated, safety-sensitive, or
production-connected work.

Bounded-lightweight does not mean relaxed governance. It means the work is
allowed only inside a narrower safe boundary: synthetic data, disposable
resources, isolated non-production systems, and no live customer, clinical,
payment, production, or industrial-control impact. Crossing that boundary
requires stopping and activating the corresponding strict profile.

## Adopt the fixture safely

The fixture references all 24 built-ins and all 11 enterprise packs. Select
the enterprise packs first; `carl pack select` retains the 24 packs selected
through the existing compatibility baseline and adds these 11:

```powershell
go run ./cmd/carl pack select `
  enterprise/argentum-innovation-lab `
  enterprise/argentum-regulated-payments `
  enterprise/brindleforge-connected-factory `
  enterprise/brindleforge-process-pilot `
  enterprise/discipline-architecture `
  enterprise/discipline-code `
  enterprise/discipline-design `
  enterprise/discipline-implementation `
  enterprise/discipline-review `
  enterprise/willowmere-clinical-records `
  enterprise/willowmere-service-design
```

Then deliberately adopt the ordinary profile data:

```powershell
Copy-Item `
  .github/carl/profiles.enterprise.example.json `
  .github/carl/profiles.json
```

The copied fixture initially activates `default`, so adoption retains the
24-pack repository effective baseline until an enterprise profile is
explicitly activated.

## Activate an example

Strict financial review:

```powershell
go run ./cmd/carl pack profile activate argentum-regulated-payments `
  --role reviewer `
  --task security-assessment
```

Balanced health design:

```powershell
go run ./cmd/carl pack profile activate willowmere-service-design `
  --role designer `
  --task accessibility-research
```

Bounded-lightweight manufacturing implementation:

```powershell
go run ./cmd/carl pack profile activate brindleforge-process-pilot `
  --role implementer `
  --task sandbox-build
```

## Inspect composition and trace

Inspect the deterministic effective set:

```powershell
go run ./cmd/carl pack effective --json
```

Inspect activation, dependency, ordering, constraint, and conflict decisions:

```powershell
go run ./cmd/carl trace --json
```

The output proves profile and pack composition against repository state. It
does not prove that a model executed the instructions.

## Restore the repository default

After an example test, restore the fixture's explicit default:

```powershell
go run ./cmd/carl pack profile activate default
go run ./cmd/carl pack effective --json
```

The restored effective pack IDs and order match the profile-absent repository
baseline exactly. The enterprise packs remain selected but inactive until a
fictional profile is activated again.

To abandon adoption entirely, review and remove the user-owned
`.github/carl/profiles.json` and `.github/carl/packs.json`; the repository then
returns to profile-absent compatibility behavior.

## Composition validation versus agent testing

These fixtures support several different evidence levels:

- Profile schema and reference validation proves the fixture is structurally
  valid.
- Effective-set and trace validation proves deterministic pack composition,
  dependency expansion, precedence, and absence of conflicts.
- Observable marker fixtures define output text that an actual agent test can
  look for after activating a scenario and discipline.
- Actual agent-execution testing requires running a model in each activated
  context and evaluating its behavior against the instructions and task.
- Seeing a marker is useful test evidence, but it does not expose or prove
  hidden reasoning, complete instruction application, or full compliance.

PR #44 validates composition and representative contexts. It does not itself
execute a model across every scenario.

## Ownership

The enterprise instruction files and example profile document are
repository-local examples. They are not cARL built-ins and are not embedded
runtime assets.

Any copied `.github/carl/packs.json` or `.github/carl/profiles.json` is
user-owned active policy state and should be reviewed and committed
deliberately.
