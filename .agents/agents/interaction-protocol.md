# Agent Interaction Protocol

This file defines how Go service specialists coordinate. Read it before using any role card.

## Default Execution Flow

1. `go-orchestrator` reads current state and classifies the task.
2. A specialist is assigned with a narrow scope.
3. The specialist reads the required skills, workflows, and rules.
4. The specialist works only inside the assigned surface.
5. Verification evidence is collected.
6. `go-code-reviewer` reviews meaningful changes before completion.
7. Progress, changelog, and relevant docs are synchronized.

## Mandatory Handoff Payload

Every non-trivial handoff should include:

```yaml
task: <one-sentence summary>
goal: <what must be true when done>
scope:
  in:
    - <owned files or boundaries>
  out:
    - <explicitly excluded work>
required_reads:
  - <skill, workflow, or file>
entry_points:
  - <starting files or packages>
risks:
  - <known technical risks>
verification:
  - <commands, tests, or checks>
done_when:
  - <objective completion condition>
```

## Ownership Rules

- `go-orchestrator` owns routing, sequence, and state awareness.
- `go-architect` owns architectural decisions, package boundaries, and trade-off analysis.
- `go-test-writer` owns test intent, failure cases, and verification shape.
- `go-backend-developer` owns implementation across service code boundaries.
- `go-db-optimizer` owns query shape, migration safety, transaction design, and index strategy.
- `go-debugger` owns reproduction, classification, and root-cause proof.
- `go-devops-engineer` owns runtime automation, delivery paths, and operational safety.
- `go-code-reviewer` owns final severity-ranked findings.
- `go-technical-writer` owns documentation accuracy and change-history clarity.

## Escalation Rules

Agents must stop and ask the user when:

- requirements are materially ambiguous
- a new architecture direction is being introduced without approval
- the proposed fix would change public contracts or runtime topology significantly
- verification repeatedly fails and the cause is unclear
- work expands beyond the original scope or touches unrelated user changes
- destructive, production-facing, or release actions are requested

## Conflict Rules

When specialists disagree:

1. Stop execution on the conflicting branch.
2. Summarize the conflict in terms of boundary, behavior, or risk.
3. Route to `go-architect` if the disagreement is about design or boundaries.
4. Route to `go-code-reviewer` if the disagreement is about severity or acceptance.
5. Escalate to the user if the choice changes scope, timeline, or behavior materially.

## Review Gate

Meaningful work is not complete until:

- required verification has run
- review findings are addressed or explicitly accepted
- docs and tracking files affected by the change are synchronized

## Stop Conditions

Stop rather than guessing when:

- the owning bounded context is unclear
- a dependency contract cannot be inferred safely
- the relevant skill or workflow conflicts with the requested change
- the worktree contains conflicting user edits in the same files
