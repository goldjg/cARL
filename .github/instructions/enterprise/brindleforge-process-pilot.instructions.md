<!-- version: 1.0.0 -->
<!-- requires: core/carl, core/cognition-governance, core/pr-contract -->
<!-- precedence-mode: additive -->
<!-- priority: 25 -->
# Brindleforge Manufacturing Process Pilot Pack

Fictional company: Brindleforge Manufacturing, an industrial manufacturer.
Use this bounded-lightweight scenario for synthetic analytics, simulation,
offline process studies, and disposable non-production pilots.

## Observable marker fixture

When this pack is active, include
`Enterprise scenario: Brindleforge Manufacturing / process pilot / bounded-lightweight`.
In the final response, include `Brindleforge pilot evidence`.

These markers are observable hooks for an actual agent-execution test. Pack
presence or effective-set validation does not execute a model, and seeing the
markers does not prove hidden reasoning or complete instruction compliance.

## Bounded-lightweight workflow

Bounded-lightweight means a narrower safe operating boundary, not relaxed
governance. Repository invariants and the cARL lifecycle remain mandatory.

- A concise improvement hypothesis, metric, simulation boundary, and time box
  may replace a full production change package.
- A short data-flow sketch may replace detailed plant architecture when no
  live OT or production integration exists.
- Targeted tests against generated or replay-safe synthetic telemetry may
  replace plant acceptance testing.
- A local pilot may use self-review; shared or promoted work requires an
  independent reviewer.
- Prefer disposable tooling and measurable learning over production-grade
  platform expansion.

## Discipline expectations

- Design: define the operator or process outcome and guard against misleading
  metrics.
- Architecture: isolate the pilot and identify every hypothetical production
  attachment point.
- Code: keep transformations deterministic and fixtures synthetic.
- Implementation: use simulators or offline sandboxes with bounded cost and an
  expiry date.
- Review: verify data provenance, metric validity, isolation, cleanup, and
  absence of live control paths.

## Non-negotiable boundary

No live controller, robot, actuator, safety function, production network,
vendor access path, production schedule, or real-time operator decision may
depend on this pilot. If any boundary is crossed, stop and activate
`brindleforge-connected-factory`.

This pack never waives repository invariants or the cARL lifecycle.
