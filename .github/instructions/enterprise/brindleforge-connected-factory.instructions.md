<!-- version: 1.0.0 -->
<!-- requires: core/baseline, core/dependency, core/identity, core/pr-contract, core/security, core/tool-permission-tiers -->
<!-- precedence-mode: immutable -->
<!-- priority: 90 -->
# Brindleforge Manufacturing Connected Factory Pack

Fictional company: Brindleforge Manufacturing, an industrial manufacturer.
Use this strict scenario for operational technology, industrial control,
robotics, safety interlocks, plant connectivity, or production scheduling.

## Observable marker fixture

When this pack is active, include
`Enterprise scenario: Brindleforge Manufacturing / connected factory / strict`.
In the final response, include `Brindleforge factory evidence`.

These markers are observable hooks for an actual agent-execution test. Pack
presence or effective-set validation does not execute a model, and seeing the
markers does not prove hidden reasoning or complete instruction compliance.

## Cross-discipline controls

- Protect human safety, machine safety, product quality, production continuity,
  and recoverability before optimisation.
- Treat IT, OT, safety systems, vendor remote access, engineering workstations,
  and cloud services as distinct trust zones.
- Never test against live controllers, robots, production lines, or safety
  systems without explicit site authorization and an approved procedure.
- Preserve manual control, safe-state behaviour, network segmentation,
  deterministic operation, and auditable change control.
- Prefer allow-listed protocols, identities, commands, and network paths.
- Use simulation, digital twins, hardware-in-the-loop test rigs, or isolated
  cells before production.

## Design

Identify operators, maintainers, safety engineers, quality teams, and affected
production processes. Define alarms, manual override, safe state, recovery,
and loss-of-connectivity behaviour.

## Architecture

Document zones and conduits, protocol direction, identity, time sensitivity,
offline operation, vendor access, patch constraints, redundancy, recovery
objectives, and safety-system independence.

## Code

Use bounded inputs and outputs, explicit timeouts, deterministic state
transitions, safe defaults, authenticated commands, and tests for replay,
stale telemetry, disconnect, restart, and partial failure.

## Implementation

Require simulation evidence, maintenance-window planning, site authorization,
backup and restore, staged cell deployment, operator communication, rollback,
and immediate containment criteria.

## Review

Block on safety bypass, uncontrolled actuation, unauthenticated commands,
flat-network assumptions, unbounded retry, unsafe default state, missing
rollback, or direct unapproved production access.

## Boundary

This pack is an engineering test policy, not a functional-safety,
cybersecurity, or regulatory certification.
