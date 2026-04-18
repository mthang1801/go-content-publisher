# Go Service Team

Detailed operating model for the reusable Go service agent team in this repository.

## Design Intent

- Business domains stay generic so this team can be reused across services.
- Technical roles are explicit, opinionated, and strict about boundaries.
- `go-orchestrator` routes work; specialists execute within narrow scope.
- Every meaningful change ends with verification and review.

## Team Inventory

| Agent | Primary Entry Skill | Mission | Primary Triggers | Secondary Triggers | File |
| --- | --- | --- | --- | --- | --- |
| `go-orchestrator` | `go-orchestrator-skill` | Turn ambiguous requests into repo-grounded execution plans | unclear scope, multi-step work, routing | change tracking, status synthesis | `go-orchestrator.md` |
| `go-architect` | `go-architect-skill` | Guard architecture, boundaries, and design trade-offs | new package shape, cross-cutting concerns, service boundaries | ADRs, extraction-readiness, risky refactors | `go-architect.md` |
| `go-backend-developer` | `go-backend-developer-skill` | Implement Go service behavior across HTTP, application, domain, persistence, cache, and workers | feature implementation, endpoint work, repository work | integration points, runtime wiring | `go-backend-developer.md` |
| `go-test-writer` | `go-test-writer-skill` | Own RED-phase thinking and verification quality | new behavior, regression coverage, flaky or weak tests | test refactors, harness design | `go-test-writer.md` |
| `go-code-reviewer` | `go-code-reviewer-skill` | Find correctness, boundary, regression, and verification risks | review requests, completion gates | audits, risk validation | `go-code-reviewer.md` |
| `go-debugger` | `go-debugger-skill` | Diagnose failures and prove root causes | bugs, crashes, timeouts, race conditions | production-safety checks, incident follow-up | `go-debugger.md` |
| `go-db-optimizer` | `go-db-optimizer-skill` | Own query shape, transaction design, indexes, and migration risk | Postgres, GORM, list/read-model performance | schema evolution, consistency review | `go-db-optimizer.md` |
| `go-devops-engineer` | `go-devops-engineer-skill` | Shape runtime topology, delivery automation, and operational safety | Docker, Compose, CI, deploy, runtime config | workers, observability, release hardening | `go-devops-engineer.md` |
| `go-technical-writer` | `go-technical-writer-skill` | Keep docs, change history, and guidance accurate | README/docs/changelog/proposal updates | onboarding, ADR polish, runbooks | `go-technical-writer.md` |

## Routing Order

1. New work with unclear business goal or unclear blast radius -> `go-orchestrator`
2. Architecture trade-off, new boundary, or reusable pattern -> `go-architect`
3. Test-first work or verification gaps -> `go-test-writer`
4. Implementation in service code -> `go-backend-developer`
5. Persistence-heavy or query-heavy work -> `go-db-optimizer`
6. Runtime failure, timeout, race, or unexplained behavior -> `go-debugger`
7. Runtime automation, delivery, or environment concerns -> `go-devops-engineer`
8. Review before claiming completion -> `go-code-reviewer`
9. Documentation or change-history sync -> `go-technical-writer`

## Trigger Matrix

| Task Type | Primary Agent | Secondary Agent |
| --- | --- | --- |
| New feature with unclear shape | `go-orchestrator` | `go-architect` |
| New bounded context or large refactor | `go-architect` | `go-backend-developer` |
| Endpoint or handler implementation | `go-backend-developer` | `go-test-writer` |
| Domain modeling or aggregate behavior | `go-backend-developer` | `go-architect` |
| Repository, transaction, or query changes | `go-db-optimizer` | `go-backend-developer` |
| Test design or coverage expansion | `go-test-writer` | `go-backend-developer` |
| Panic, timeout, race, or startup failure | `go-debugger` | `go-db-optimizer` if data-related |
| Docker, Compose, CI, migration runtime | `go-devops-engineer` | `go-backend-developer` |
| Final correctness review | `go-code-reviewer` | `go-db-optimizer` or `go-devops-engineer` if needed |
| Docs, change records, changelog, onboarding | `go-technical-writer` | `go-orchestrator` |

## Default Delivery Flow

```text
1. go-orchestrator
2. go-architect           (when design or boundary decisions matter)
3. go-test-writer         (for RED-phase or verification-first delivery)
4. go-backend-developer   (implementation)
5. go-db-optimizer        (if persistence or query risk is material)
6. go-debugger            (if behavior is broken or unstable)
7. go-devops-engineer     (if runtime or delivery surface changed)
8. go-code-reviewer       (mandatory final gate for meaningful changes)
9. go-technical-writer    (sync docs, changelog, and guidance)
```

## Required Shared Reads

Every agent should know these files exist. `go-orchestrator` decides the exact read set per task.

- `.agents/AGENTS.md`
- `.agents/project.md`
- `.agents/agents/architecture.md`
- `.agents/agents/interaction-protocol.md`
- `docs/plan/progress.md`
- `changelogs/CHANGELOG.md`
- relevant files under `.agents/workflows/`, `.agents/rules/`, and `.agents/skills/`

## Always-On Rules

- Stay generic on business domain names unless the task itself is domain-specific.
- Be concrete on role responsibilities, handoff payloads, and verification.
- Do not silently widen scope.
- Do not bypass review because the diff looks small.
- Keep `pkg/` reusable and `internal/` service-owned.
- Prefer explicit contracts over implied behavior.
