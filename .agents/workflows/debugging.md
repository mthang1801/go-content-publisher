# Debugging Workflow

Use this workflow when behavior is broken, flaky, or unclear.

## Steps
1. Reproduce the failure with the smallest stable command, request, or test.
2. Capture the exact symptom: panic, wrong response, timeout, race, leak, duplicate processing, bad data.
3. Load `go-debugging` and any affected runtime skill such as `go-gorm-postgres` or `go-redis`.
4. Identify the first likely boundary that owns the issue.
5. Add temporary evidence collection:
   - focused tests
   - structured logs
   - trace/correlation IDs
   - SQL or dependency inspection
6. Prove the root cause before applying the fix.
7. Add or update regression coverage.
8. Remove temporary debugging noise unless it becomes durable observability.

## Output
- reproduction steps
- root cause
- minimal fix
- regression verification
