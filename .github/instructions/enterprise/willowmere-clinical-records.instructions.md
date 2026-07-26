<!-- version: 1.0.0 -->
<!-- requires: core/baseline, core/dependency, core/identity, core/pr-contract, core/security, core/tool-permission-tiers -->
<!-- precedence-mode: immutable -->
<!-- priority: 90 -->
# Willowmere Health Network Clinical Records Pack

Fictional company: Willowmere Health Network, a healthcare provider. Use this
strict scenario for clinical records, care workflows, patient identity,
clinical integrations, or systems whose failure could affect patient safety.

## Adherence signal

At the first substantive response, include
`Enterprise scenario: Willowmere Health Network / clinical records / strict`.
In the final response, include `Willowmere clinical evidence`.

## Cross-discipline controls

- Minimise collection, use, display, retention, and disclosure of health and
  identity data.
- Treat patient safety, privacy, clinical continuity, data provenance, and
  authorized access as co-equal design constraints.
- Use synthetic patients and de-identified fixtures only; never place real
  patient or workforce records in repository content or prompts.
- Preserve a human-verifiable audit trail for access, correction, disclosure,
  consent, and clinically significant decisions.
- Do not let automation silently replace accountable clinical judgment.
- Fail safely: distinguish unavailable, incomplete, stale, conflicting, and
  unauthorized data.

## Design

Include patients, clinicians, carers, accessibility needs, consent, emergency
access, downtime workflows, alert fatigue, and the harm caused by confusing or
missing information.

## Architecture

Document clinical and administrative trust boundaries, provenance, consent,
record lifecycle, tenant separation, emergency access, interoperability,
recovery objectives, and safe degraded operation.

## Code

Enforce authorization at the patient/resource boundary, validate identifiers
and clinical code systems, retain provenance, avoid silent coercion, and test
wrong-patient, stale-data, partial-integration, and unauthorized-access paths.

## Implementation

Require representative synthetic validation, clinical or safety-owner review
for workflow changes, staged rollout, monitoring, downtime procedures, and a
tested rollback or containment plan.

## Review

Block on patient mix-up, unauthorized disclosure, missing provenance, unsafe
defaulting, lost clinical context, unreviewed workflow impact, or inadequate
downtime behaviour.

## Boundary

This pack is an engineering test policy, not medical, privacy, or regulatory
compliance certification.
