# Copied Base Skills

This directory contains a copied local base skill set for this repository.

## Purpose

- keep this repository self-contained
- let Go sample agents reference a stable local base skill set
- separate broad, portable expertise from repo-local enforcement under `.agents/skills/`

## Usage Rule

When an agent card mentions copied base skills, load them from this directory first. Then load the corresponding repo-local bridge or precise skill under `.agents/skills/`.

This is a physical copied skill set for the repository. Do not resolve broad base skills from any external root skill tree when operating inside this repo.

## Do Not Assume

- copied base skills know the exact structure of this repository
- copied base skills replace local architecture, workflow, or governance skills

Use them for broad expertise first, then narrow with `.agents/skills/`.
