# Testing Go Rules

- Every behavior change requires verification evidence.
- Prefer RED -> GREEN -> REFACTOR for non-trivial changes.
- Unit tests should dominate, but adapters touching Postgres, Redis, or brokers need integration coverage where risk is real.
- Race-prone code should be exercised with `go test -race` when feasible.
- Keep mocks at port boundaries. Do not deeply mock Gin, GORM, or broker clients unless there is no better seam.
- Failing tests must describe the business behavior or contract that broke.
