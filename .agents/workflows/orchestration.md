# Orchestration Workflow

Use this workflow when work is large enough to need planning, specialist routing, or tracked change artifacts.

## Steps
1. Read `docs/plan/progress.md`, `changelogs/CHANGELOG.md`, `.agents/project.md`, and `.agents/agents/architecture.md`.
2. Clarify the goal, constraints, affected bounded contexts, and affected runtime pieces.
3. Decide whether the work is trivial or needs `changelogs/changes/<slug>/`.
4. Emit a concise execution plan with owners, verification, and doc updates.
5. Route work to specialists:
   - `go-architect` for package shape and boundaries
   - `go-backend-developer` for implementation
   - `go-test-writer` for RED-phase or regression coverage
   - `go-db-optimizer` for Postgres/GORM concerns
   - `go-debugger` for failure analysis
   - `go-code-reviewer` before completion
6. Require verification evidence before marking work complete.
7. Update `progress.md`, `CHANGELOG.md`, and any change records.

## Output Shape
Prefer a short YAML plan:

```yaml
goal: ...
scope:
  - ...
owners:
  architect: ...
  implementation: ...
verification:
  - ...
docs:
  - ...
risks:
  - ...
```

## Stop Conditions
- Missing requirements or ambiguous ownership.
- Proposed change would violate clean architecture rules.
- Verification cannot be run or reproduced.
