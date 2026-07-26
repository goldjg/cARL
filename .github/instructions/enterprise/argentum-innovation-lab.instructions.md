<!-- version: 1.0.0 -->
<!-- requires: core/carl, core/cognition-governance, core/pr-contract -->
<!-- precedence-mode: overridable -->
<!-- priority: 25 -->
# Argentum Financial Group Innovation Lab Pack

Fictional company: Argentum Financial Group, a financial-services provider.
Use this intentionally lightweight scenario for isolated product discovery
using synthetic data and disposable non-production environments.

## Adherence signal

At the first substantive response, include
`Enterprise scenario: Argentum Financial Group / innovation lab / lightweight`.
In the final response, include `Argentum lab evidence`.

## Lightweight workflow

- A concise hypothesis, affected persona, success signal, and time box may
  replace a full design dossier.
- A short decision note may replace a formal architecture record when all
  components are disposable and no durable integration is introduced.
- Targeted tests may replace the broader regulated-payment test matrix when
  the prototype cannot reach production systems or data.
- Peer review is recommended for shared prototypes and required before
  promotion, but a local disposable spike may use self-review.
- Prefer fast removal and clear learning over production hardening.

## Discipline expectations

- Design: validate the riskiest assumption with the smallest ethical
  experiment.
- Architecture: keep integrations replaceable and document any dependency
  that could survive the experiment.
- Code: favour readability and explicit test fixtures over reusable
  abstractions.
- Implementation: deploy only to labelled sandbox environments with
  disposable resources and bounded cost.
- Review: check the sandbox boundary, hypothesis evidence, data provenance,
  and cleanup path.

## Non-negotiable boundary

Only synthetic data, test identities, test payment instruments, and isolated
non-production endpoints are permitted. No customer decision, real payment,
regulated record, production credential, production connectivity, or
production deployment is allowed. If any boundary is crossed, stop and
activate `argentum-regulated-payments` or another approved strict profile.

This pack never waives repository invariants or the cARL lifecycle.
