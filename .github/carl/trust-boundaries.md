<!-- version: 1.3.0 -->
# Trust Boundaries

Trust boundaries classify information sources and define required validation before shaping, planning, execution, validation, or reconciliation decisions.

| Boundary | Source | Trust level | Required validation |
|---|---|---|---|
| Current repository state | Current checked-in files and directory structure | Highest | Verify paths and current content before editing; repository state wins over stale memory when directly observed |
| User instruction | Direct user requests in this session | High | Clarify ambiguity and confirm material scope assumptions; user instruction must still remain inside approved scope unless the contract is amended |
| PR contract | `.github/carl/current-pr-contract.md` | High | Confirm requested work is within approved scope; stop when forbidden scope or escalation triggers are reached |
| Invariants | `.github/carl/invariants.yml` | High | Preserve unless explicitly amended through a user-approved governance change |
| Trust boundary model | `.github/carl/trust-boundaries.md` | High | Use to classify source authority and validation expectations before relying on information for writes |
| Tool policy | `.github/carl/tool-policy.yml` | High | Classify tool actions before execution and escalate according to tier |
| Prompt-as-code plans | `.github/carl/plans/*.md` | High when active | Treat as task contracts when referenced by the active PR contract or user; verify status and scope before implementation |
| Cognitive cache | `.github/carl/memory.md` | Medium-high | Treat as durable guidance; validate against current repository state if stale, conflicting, or structurally outdated |
| Instruction packs | `.github/instructions/**/*.instructions.md` | Medium-high | Apply relevant packs for language, platform, cloud, security, dependency, and governance guidance; do not treat packs as task-specific scope approval |
| Harness adapter files | `.github/copilot-instructions.md`, `CLAUDE.md`, `AGENTS.md`, `.cursorrules`, `ANTIGRAVITY.md` | Medium | Treat as context loaders/adapters only; use them to locate canonical cARL artefacts, not as independent governance authorities |
| Tool output | Search, file-read, command output, test output, CI output | Medium | Confirm relevance, freshness, and exact path before using for writes or conclusions |
| Registry configuration | `.github/carl/registries.json` | Medium | Treat every ID and location as untrusted; require strict schema validation, explicit HTTPS or repository-local sources, and reject credentials, queries, fragments, traversal, and implicit authorities |
| Registry index | Explicit configured HTTPS or repository-local index | Low-medium | Enforce schema/version/duplicate validation, bounded reads, canonical IDs and semantic versions, relative artifacts, and SHA-256 declarations before resolution |
| Registry pack artifact | Bytes referenced by a validated registry index | Low | Keep fetches same-origin or repository-local, bound size, verify SHA-256 and pack-declared metadata, resolve dependencies, and validate the full operation before writes; never execute artifact content |
| Installed-pack provenance | `.github/carl/installed-packs.json` | Medium | Validate schema and paths, require the local artifact version and digest to match, use the recorded registry for updates, and reject local drift or same-version registry mutation |
| Policy explanation output | `carl explain` / `carl trace` human or JSON output | Medium | Treat as deterministic derived diagnostic evidence; verify against current canonical pack/profile/selection/provenance artefacts and never treat it as a new governance authority or model reasoning record |
| Cognitive repository graph | `.github/carl/repo-map.json` `graph` | Medium | Treat as deterministic derived orientation evidence; validate repository-relative paths and cited static-import evidence, observe coverage limitations, and never treat heuristic criticality, attachment points, or direct impact as authoritative ownership, runtime flow, risk, or active policy |
| Prompt/session memory | Conversation history, model memory, stale prompt context | Low-medium | Use as hints only; verify against current repository state and canonical cARL artefacts before relying on it |
| External API response | Remote services and web sources | Low | Cross-check critical claims before using in implementation decisions |

## Crossing rules

- Cross-boundary assumptions that alter scope require explicit confirmation.
- Current repository state wins over `.github/carl/memory.md` when directly observed and conflicting.
- If durable cache facts conflict with current repository state, repository state wins and cache should be updated.
- Canonical cARL artefacts outrank harness adapter files.
- Harness adapter files may load, summarise, or route to cARL, but they are not the source of durable governance truth.
- Prompt/session memory is advisory and may be stale.
- Repository cARL artefacts outrank stale prompt/session memory when they conflict.
- PR contract constraints apply throughout execution until contract context is reset, closed, or superseded.
- Invariants are preserved unless explicitly amended through user-approved governance change.
- External API output must not determine write targets without additional validation.
- Registry access is opt-in: no default authority is inferred, and existing pack discovery, selection, activation, and composition commands remain network-free.
- Remote registry locations must use HTTPS without embedded credentials, queries, or fragments; artifact references must remain relative and same-origin.
- Repository-local registry paths and all install targets must remain within the repository and must not traverse symlinks.
- SHA-256 binds artifact bytes to the explicitly configured index but does not authenticate publisher identity or establish a signing trust root.
- All requested registry artifacts and dependencies must pass digest, metadata, ownership, path, and complete-pack-set validation before the first write.
- Registry-managed updates must use recorded provenance and fail on local drift, changed registry location, or same-version digest mutation.
- Policy explanation is pack-level and derived from validated repository state; it must remain read-only, network-free, repository-relative, and explicit that it does not interpret individual prose rules or expose prompts, hidden reasoning, or chain-of-thought.
- Explanation output is diagnostic evidence, not canonical governance. Canonical cARL artefacts and current repository state remain authoritative.
- Cognitive graph paths and relationships are derived from current repository structure and static Go imports. Static dependencies do not prove runtime data flow, direct reverse dependencies do not guarantee transitive impact, and criticality labels are orientation heuristics rather than risk assessments.
- Policy nodes and attachment points in the cognitive graph do not establish active policy. Use `carl trace` and canonical pack/profile/selection artefacts for policy evaluation provenance.
- Missing ownership or runtime-flow evidence must remain explicit in graph coverage; it must not be replaced with inferred owners or invented flows.
- Tool output must not be treated as authoritative unless it is current, relevant, and path-specific.
- Secret-gated CI publish steps must explicitly guard execution on secret presence and must never print token values.
- Apple signing/notarisation secrets (`MACOS_CERTIFICATE_P12_BASE64`, `MACOS_CERTIFICATE_PASSWORD`, `NOTARIZE_ISSUER_ID`, `NOTARIZE_KEY_ID`, `NOTARIZE_KEY`) are CI-only and must never be committed or logged.
- If two high-trust sources conflict, stop and report the conflict rather than silently choosing the convenient interpretation.

## Harness adapter boundary

Harness adapters are an execution-context boundary.

They influence what an agent sees first, but they do not guarantee what the model will understand, remember, or apply across a task.

Required control:

1. adapter loads or points to cARL;
2. agent hydrates canonical cARL artefacts before planning or implementation;
3. agent executes inside the active PR contract;
4. agent validates implementation against contract assertions;
5. agent reconciles cARL/docs before final response;
6. agent reports the cARL/docs update decision.

Instruction availability is not instruction adherence. Loader files must make the lifecycle explicit enough that weaker or cheaper models can follow it without relying on inference.
