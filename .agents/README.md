# Go Service Agent Template

This `.agents` directory is a reusable operating system for Go service repositories.

It keeps the strongest parts of the reference operating model:

- orchestrator-first planning
- spec-first changes
- TDD and verification gates
- specialist delegation
- progress and changelog discipline

The difference is that the team model is now explicit and detailed: each specialist card is an operating contract, not a short profile.

## Start With The Team Docs

Read these first:

1. `.agents/agents/GO-TEAM.md`
2. `.agents/agents/interaction-protocol.md`
3. `.agents/agents/architecture.md`

## Directory Map

```text
.agents/
├── agents/       # team inventory, role contracts, routing, and interaction rules
├── skills/       # repo-local bridge skills and precise implementation guidance
├── workflows/    # required execution flows
├── rules/        # hard constraints and codebase expectations
├── contexts/     # operating modes
├── prompts/      # reusable prompt scaffolds
├── hooks/        # optional automation hooks
├── scripts/      # helper scripts for local development
└── specs/        # agent-system-specific specs and evolution docs
```

## Design Goals

1. Be generic enough for future Go service repos.
2. Encode Go Clean Architecture and DDD correctly.
3. Keep technical roles detailed and strict even when business domains remain generic.
4. Support phased growth from single-service samples toward service-oriented systems.

## Relationship To This Sample

This repository is the first consumer of this template.

That means:

- some docs mention Gin, GORM, and Postgres because those are the first chosen runtime pieces
- the agent system itself must remain generic and reusable
- agent roles should be reusable in other Go repos without being rewritten around commerce-specific language

## Skill Architecture

This sample now uses two layers of skills:

1. `skills/`
   Copied local base skills bundled with the sample.
2. `.agents/skills/`
   Repo-local bundle skills, bridge skills, and precise local skills.

The intended read order is:

1. role bundle skill
2. copied base skill(s)
3. matching repo-local bridge skill(s)
4. precise repo-local skill(s) for the active surface

## Operating Model

The expected loop is:

1. orchestrator reads progress, changelog, and team protocol
2. routing chooses the smallest correct specialist chain
3. skills and workflows constrain implementation
4. verification and review gates run
5. docs and tracking artifacts are synchronized
