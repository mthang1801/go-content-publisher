# Agent Card Hardening

## Context

The initial `.agents/agents` scaffold established the correct team shape, but the role cards were still too light. They described who the specialists were without fully describing how they should operate.

## Decision

Promote the Go agent cards from short profiles into operating contracts.

This means:

- add a team-level source of truth in `GO-TEAM.md`
- tighten `interaction-protocol.md` around handoff payloads and escalation
- rewrite each specialist card with trigger conditions, mandatory reads, bundled skills, workflow, boundaries, non-goals, and success metrics
- expand `architecture.md` into a structural specification for `cmd/`, `internal/`, `pkg/`, testing, docs, and future runtime topology
- expand `go-orchestrator.md` into a fuller routing and governance document closer in depth to the reference systems
- align the remaining specialist cards to the same depth so the whole roster is internally consistent
- split skill loading into two explicit layers: copied local base skills under `skills/` and repo-local bridge plus precision skills under `.agents/skills/`

## Consequences

- the agent system becomes more reusable because technical role behavior is explicit
- orchestrator routing becomes less dependent on guesswork
- contributors can read a single card and understand ownership, handoff, and completion criteria
- the documentation surface grows, so future edits must keep role cards in sync with actual repo behavior

## Scope Notes

- business-domain language remains generic
- the roster remains backend-first for now
- this hardening pass does not add service code under `cmd/`, `internal/`, or `pkg/`
