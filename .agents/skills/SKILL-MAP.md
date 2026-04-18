# Skill Map

This file maps agent bundle skills to their underlying copied base, bridge, and precise local skills.

## Agent Bundle Map

| Agent | Primary Bundle Skill | Role Card |
| --- | --- | --- |
| `go-orchestrator` | `go-orchestrator-skill` | `agents/go-orchestrator.md` |
| `go-architect` | `go-architect-skill` | `agents/go-architect.md` |
| `go-backend-developer` | `go-backend-developer-skill` | `agents/go-backend-developer.md` |
| `go-test-writer` | `go-test-writer-skill` | `agents/go-test-writer.md` |
| `go-code-reviewer` | `go-code-reviewer-skill` | `agents/go-code-reviewer.md` |
| `go-debugger` | `go-debugger-skill` | `agents/go-debugger.md` |
| `go-db-optimizer` | `go-db-optimizer-skill` | `agents/go-db-optimizer.md` |
| `go-devops-engineer` | `go-devops-engineer-skill` | `agents/go-devops-engineer.md` |
| `go-technical-writer` | `go-technical-writer-skill` | `agents/go-technical-writer.md` |

## Underlying Stack Map

| Bundle Skill | Copied Base Skills | Bridge Skills | Precise Local Skills |
| --- | --- | --- | --- |
| `go-orchestrator-skill` | `strategic-compact`, `repo-scan`, `search-first`, `senior-sysdesign` | `strategic-compact`, `repo-scan`, `search-first`, `senior-sysdesign` | `jira`, `git-workflow` |
| `go-architect-skill` | `go-senior-expert`, `backend-patterns-skill`, `senior-sysdesign` | `go-senior-expert`, `backend-patterns-skill`, `senior-sysdesign` | `go-clean-architecture`, `go-ddd`, `go-backend-patterns`, `go-microservices`, `go-saga`, `security-review` |
| `go-backend-developer-skill` | `go-senior-expert`, `backend-patterns-skill`, `senior-dev` | `go-senior-expert`, `backend-patterns-skill`, `senior-dev` | `go-backend-patterns`, `go-clean-architecture`, `go-ddd`, `go-gorm-postgres`, `go-error-handling`, `go-testing-tdd`, `go-redis`, `go-microservices`, `go-saga`, `go-stream-pipeline`, `git-workflow`, `security-review` |
| `go-test-writer-skill` | `go-senior-expert`, `senior-dev` | `go-senior-expert`, `senior-dev` | `go-testing-tdd`, `go-backend-patterns`, `go-clean-architecture`, `go-gorm-postgres`, `go-error-handling`, `go-debugging` |
| `go-code-reviewer-skill` | `go-senior-expert`, `backend-patterns-skill`, `senior-dev`, `senior-sysdesign`, `senior-dba` | `go-senior-expert`, `backend-patterns-skill`, `senior-dev`, `senior-sysdesign`, `senior-dba` | `go-backend-patterns`, `go-clean-architecture`, `go-ddd`, `go-gorm-postgres`, `go-testing-tdd`, `go-error-handling`, `go-redis`, `go-microservices`, `go-saga`, `go-stream-pipeline`, `security-review`, `git-workflow` |
| `go-debugger-skill` | `go-senior-expert`, `backend-patterns-skill`, `senior-dev`, `senior-dba`, `devops-senior-expert` | `go-senior-expert`, `backend-patterns-skill`, `senior-dev`, `senior-dba`, `devops-senior-expert` | `go-debugging`, `go-error-handling`, `go-backend-patterns`, `go-gorm-postgres`, `go-redis`, `go-stream-pipeline`, `go-microservices`, `go-testing-tdd` |
| `go-db-optimizer-skill` | `senior-dba`, `supabase-postgres-best-practices`, `go-senior-expert` | `senior-dba`, `supabase-postgres-best-practices`, `go-senior-expert` | `go-gorm-postgres`, `go-clean-architecture`, `go-ddd`, `go-testing-tdd`, `security-review` |
| `go-devops-engineer-skill` | `devops-senior-expert`, `go-senior-expert` | `devops-senior-expert`, `go-senior-expert` | `git-workflow`, `security-review`, `go-microservices`, `go-redis`, `go-stream-pipeline` |
| `go-technical-writer-skill` | `doc-writer` | `doc-writer` | `jira`, `git-workflow`, `security-review` |

## Reading Order Example

For a DB-heavy review:

1. `.agents/skills/go-code-reviewer-skill/SKILL.md`
2. `skills/senior-dba/SKILL.md`
3. `.agents/skills/senior-dba/SKILL.md`
4. `.agents/skills/go-gorm-postgres/SKILL.md`
5. `.agents/skills/security-review/SKILL.md`
