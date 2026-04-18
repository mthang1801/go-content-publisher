# Code Review Workflow

Use this workflow before considering a task complete.

## Review Order
1. Correctness and regressions
2. Architecture and boundary integrity
3. Data consistency and failure handling
4. Test coverage and verification quality
5. Security and operational safety
6. Naming, readability, and documentation

## Review Questions
- Does the change preserve `Presentation -> Application -> Domain <- Infrastructure`?
- Did any framework or ORM detail leak into domain code?
- Are transactions, idempotency, and retries handled intentionally?
- Do tests prove the behavior that changed?
- Are logs, metrics, and errors safe and useful?

## Required Outputs
- findings first, ordered by severity
- open questions or assumptions
- brief change summary only after findings

## Completion Rule
No meaningful change is complete until review findings are addressed or explicitly accepted.
