# Repo-Local Skills

This directory is the repo-local skill layer for this repository.

## Skill Layers

There are now two skill roots in the repository:

1. `skills/`
   Broad copied local base skills bundled with the repository.
2. `.agents/skills/`
   Repo-local entry, bridge, and precise skills that know the Go service architecture, workflow, and governance model.

## Skill Categories

Inside `.agents/skills/` there are three categories:

1. agent bundle skills
   One entry skill per agent role such as `go-backend-developer-skill`.
2. bridge skills
   Repo-local reinterpretations of copied base skills such as `go-senior-expert` or `senior-dba`.
3. precise local skills
   Narrow repository rules such as `go-clean-architecture`, `go-gorm-postgres`, or `jira`.

## Loading Rule

When activating an agent role:

1. load the agent's primary bundle skill from `.agents/skills/`
2. let the bundle skill load the copied base skill(s) from `skills/`
3. let the bundle skill load the matching bridge skill(s) from `.agents/skills/`
4. let the bundle skill load only the precise local skill(s) needed by the active surface

When you are not activating a role but need a narrow repository rule directly:

1. load the orchestrator or role bundle skill first when possible
2. otherwise load the precise local skill directly, but only for a clearly bounded concern

Never substitute any external root skill tree in place of the copied local base skills.

## Important Distinction

- copied base skills answer: "what does a senior engineer usually know about this topic?"
- agent bundle skills answer: "what stack should this role load first?"
- repo-local bridge skills answer: "how should that broad skill be interpreted in this sample?"
- precise repo-local skills answer: "what exact rules apply here?"

See `SKILL-MAP.md` for the mapping.
