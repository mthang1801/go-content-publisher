# New Feature Workflow

Use this workflow for non-trivial features, refactors, or cross-cutting enhancements.

## Steps
1. Confirm the business goal and acceptance criteria.
2. Create or update `changelogs/changes/<slug>/proposal.md` and `tasks.md`.
3. Load the relevant skills before coding:
   - `go-clean-architecture`
   - `go-ddd`
   - `go-backend-patterns`
   - plus infra/runtime skills as needed
4. Write or update failing tests at the narrowest useful layer.
5. Implement the smallest passing change.
6. Refactor for boundary correctness, naming, and observability.
7. Run verification and capture evidence.
8. Request review from `go-code-reviewer`.
9. Update progress and changelog.

## Verification Minimum
- format or lint checks if present
- tests for the changed behavior
- dependency-specific checks if Postgres, Redis, or worker flow changed

## Do Not
- start with transport code when the domain is still unclear
- skip change records for meaningful work
- close work before review and verification
