# Fundamentals, Boundaries & Testing in Go

Derived from:
- `documents/assets/go/fundamental/README.md`
- `documents/assets/go/fundamental/errors/README.md`
- `documents/assets/go/fundamental/interfaces/README.md`
- `documents/assets/go/fundamental/packages/README.md`
- `documents/assets/go/fundamental/testing/README.md`
- `documents/assets/go/fundamental/oop/README.md`
- `documents/assets/go/fundamental/typescript-to-go/README.md`

## When To Read

Load this reference when the task is really about language semantics, package boundaries, interface ownership, tests, or a TS/Java mental-model mismatch rather than framework specifics.

## Symptom Router

| Symptom | Open lane | Why |
| --- | --- | --- |
| `errors.Is`, `errors.As`, wrapping, `errors.Join`, custom error shape | `fundamental/errors/` | Error meaning and call-site branching are the real question |
| "where should this interface live?", nil-interface confusion, mocking seams | `fundamental/interfaces/` | Contract ownership and boundary design dominate |
| import cycles, `internal/`, `pkg/`, `go.mod`, `go work`, monorepo shape | `fundamental/packages/` | Repo topology and module boundaries are the problem |
| unit test shape, fuzzing, benchmarks, integration test confidence | `fundamental/testing/` | Verification strategy is the pressure point |
| Java/TypeScript OOP expectations collide with Go | `fundamental/oop/` and `fundamental/typescript-to-go/` | The mental model needs translation first |

## Non-Negotiables

- Errors are values. Wrap at boundaries, preserve cause chains, and expose sentinel errors only when callers truly need stable branching.
- Interfaces belong to the consumer side. Prefer tiny seams, frequently one method, and avoid speculative "god interfaces".
- Packages are a readability boundary. Prefer capability-based names and use `internal/` to enforce encapsulation.
- Test the cheapest layer that can prove correctness: value object and aggregate tests first, then adapter/integration tests.
- Use race tests for concurrency, benchmarks for performance claims, and fuzzing for hostile inputs.

## Mental-Model Translation

| Coming from | Go equivalent |
| --- | --- |
| class | `struct` + methods |
| inheritance | embedding + composition |
| exception throwing | returned `error` |
| `async/await` everywhere | synchronous by default; add goroutines only for a real coordination need |
| framework DI | constructor injection + consumer-defined interfaces |
| nullable/optional everything | zero value, pointer, or explicit value object depending on semantics |

## Practical Defaults

- Use table-driven tests with subtests for behavior matrices.
- Reach for `httptest` before spinning up a full server.
- Reach for container-backed tests when DB/broker/cache semantics matter more than pure function behavior.
- If a review comment says "this feels un-Go-like", inspect package naming, interface placement, error shape, and test style before touching architecture.
