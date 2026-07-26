<!-- version: 1.0.0 -->
<!-- requires: core/baseline, core/dependency, core/identity, core/pr-contract, core/security, core/tool-permission-tiers -->
<!-- precedence-mode: immutable -->
<!-- priority: 90 -->
# Argentum Financial Group Regulated Payments Pack

Fictional company: Argentum Financial Group, a financial-services provider.
Use this strict scenario for payment, account, ledger, fraud, identity, or
regulated customer-data systems.

## Adherence signal

At the first substantive response, include
`Enterprise scenario: Argentum Financial Group / regulated payments / strict`.
In the final response, include `Argentum control evidence`.

## Cross-discipline controls

- Treat payment data, credentials, account data, fraud signals, and customer
  identifiers as sensitive even when examples are synthetic.
- Maintain separation of duties between request, approval, implementation,
  deployment, and review for material production changes.
- Require traceability from business control to design decision, code change,
  test evidence, and rollout approval.
- Preserve immutable financial events. Corrections use compensating entries;
  do not silently rewrite ledger history.
- Use least privilege, strong identity validation, encryption in transit and
  at rest, and auditable access.
- Never use live customer, card, account, credential, or production endpoint
  data in examples or tests.

## Design

Identify customer harm, financial loss, dispute, accessibility, fraud, and
operational-support outcomes. Define degraded and failure-state behaviour
before happy-path interaction details.

## Architecture

Document data classification, trust boundaries, transaction idempotency,
double-entry or reconciliation invariants, recovery objectives, key
management, audit retention, and third-party payment boundaries.

## Code

Use exact monetary types, deterministic rounding, idempotency keys, replay
protection, authorization at the resource boundary, and negative tests for
duplicate, stale, unauthorized, and partially failed operations.

## Implementation

Require staged rollout, reconciliation checks, rollback or compensating-action
plans, monitoring, and named approval before production-affecting mutation.

## Review

Block on authorization bypass, money or ledger integrity loss, secret
exposure, missing audit evidence, unsafe migration, or unbounded third-party
failure. Findings must map to the relevant financial control or invariant.

## Boundary

This pack is an engineering test policy, not a statement of legal or
regulatory compliance.
