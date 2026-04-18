# Go Service Agent System

This directory defines the reusable operating system for specialist agents in Go service repositories.

## Start Here

Read in this order:

1. `GO-TEAM.md`
2. `interaction-protocol.md`
3. `architecture.md`
4. the specialist card you are about to use

## What Lives Here

```text
agents/
├── GO-TEAM.md               # team inventory, routing order, trigger matrix, SDLC flow
├── README.md                # entrypoint for the agent system
├── architecture.md          # architectural constraints all agents must honor
├── interaction-protocol.md  # handoff, escalation, and review rules
└── go-*.md                  # detailed operating contracts for each specialist role
```

## Rules Of Engagement

- Agent cards are operating contracts, not short role labels.
- Each agent has one primary entry bundle skill under `.agents/skills/`.
- `go-orchestrator` routes work, but it does not replace specialists.
- Specialists own narrow scopes and must not silently widen them.
- Business/domain language stays generic unless a task explicitly names a domain.
- Architecture and review gates remain mandatory for non-trivial work.

## Current Team

- `go-orchestrator`
- `go-architect`
- `go-backend-developer`
- `go-test-writer`
- `go-code-reviewer`
- `go-debugger`
- `go-db-optimizer`
- `go-devops-engineer`
- `go-technical-writer`

## Relationship To The Rest Of `.agents`

- `skills/` teaches techniques and reference patterns
- `workflows/` defines required execution sequences
- `rules/` encodes hard constraints
- `agents/` decides who should do the work and how handoffs happen
