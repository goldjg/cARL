<!-- version: 1.0.0 -->
<!-- requires: core/carl, core/cognition-governance, core/pr-contract -->
<!-- precedence-mode: overridable -->
<!-- priority: 25 -->
# Brindleforge Manufacturing Process Pilot Pack

Fictional company: Brindleforge Manufacturing, an industrial manufacturer.
Use this intentionally lightweight scenario for synthetic analytics,
simulation, offline process studies, and disposable non-production pilots.

## Adherence signal

At the first substantive response, include
`Enterprise scenario: Brindleforge Manufacturing / process pilot / lightweight`.
In the final response, include `Brindleforge pilot evidence`.

## Lightweight workflow

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
