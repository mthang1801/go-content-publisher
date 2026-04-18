# Security Go Rules

- Validate input at the boundary and enforce authorization explicitly.
- Never expose raw internal errors, stack traces, SQL messages, or secrets to clients.
- Keep secrets in environment or secret providers, not in source or logs.
- Use safe defaults for timeouts, TLS, and retry behavior on outbound clients.
- Treat internal services and background workers as untrusted by default for validation and auth decisions.
- Record security-relevant decisions in tests or change docs when behavior changes.
