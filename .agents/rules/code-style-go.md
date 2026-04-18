# Code Style Go Rules

- Favor explicit constructors and small structs over hidden setup.
- Thread `context.Context` through I/O and request-scoped operations.
- Keep methods short enough that invariants and side effects are obvious.
- Use wrapped errors with `%w` and preserve actionable context.
- Prefer table-driven tests when they improve coverage without obscuring intent.
- Use comments sparingly and only for non-obvious intent or invariants.
- Avoid package-level globals except for true constants or deliberately shared immutable configuration.
