# Code Review Prompt

Use this prompt as a scaffold:

```text
Review the change as a Go service reviewer.

Focus on:
- correctness and regressions
- architecture boundary violations
- domain/application/infrastructure leakage
- transaction, idempotency, and failure handling
- security and observability risks
- missing or weak tests

Report findings first with file references, then open questions, then a short summary.
```
