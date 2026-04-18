# Agent Bundle Skills

## Context

The repository already had strong role cards and a growing skill stack, but the activation path was still fragmented. Each agent card listed copied base skills, repo-local bridge skills, and precise local skills directly. That made the system accurate, but it did not give each role a single skill entrypoint.

## Decision

Introduce one bundle skill per agent role.

Each bundle skill becomes the primary entry surface for its corresponding agent and owns:

- trigger conditions
- the underlying skill stack
- loading order
- workflow emphasis
- output expectations
- escalation hints

The role card remains the operating contract for responsibilities, workflow, boundaries, and completion standards, but it no longer acts as the place where the entire skill stack is manually expanded.

## Bundle Skills

- `go-orchestrator-skill`
- `go-architect-skill`
- `go-backend-developer-skill`
- `go-test-writer-skill`
- `go-code-reviewer-skill`
- `go-debugger-skill`
- `go-db-optimizer-skill`
- `go-devops-engineer-skill`
- `go-technical-writer-skill`

## Consequences

- every role now has a single entry skill
- agent activation is easier to port to another repository
- copied base skills, bridge skills, and precise local skills remain reusable building blocks rather than entrypoints
- role cards and bundle skills must stay synchronized when triggers or scope change

## Scope Notes

- this pass does not remove any existing bridge or precise local skills
- this pass changes entry surfaces and mapping, not technical architecture
- this pass keeps business-domain wording generic
