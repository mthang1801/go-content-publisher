# Test Generator Workflow

Use this workflow when generating or expanding tests for a Go service.

## Steps
1. Identify the behavior, risk, and owning layer.
2. Pick the narrowest useful test level.
3. Model success path, business rejection path, and dependency failure path.
4. Prefer table-driven tests when they reduce duplication without hiding intent.
5. Add fixtures/builders that stay local to the bounded context.
6. Run the target tests first, then wider verification if the seam is shared.

## Heuristics
- Domain logic: unit tests with no network or DB.
- Application orchestration: unit or integration depending on transaction risk.
- Repository adapters: integration tests against real dependencies where feasible.
- HTTP contracts: handler tests with clear request/response assertions.
