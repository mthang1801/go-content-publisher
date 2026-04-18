# Go Sample Project

## What This Project Is

This repository is a reusable Go service template plus a reference sample implementation.

The first sub-project is the `.agents` system. Application code comes later.

## Why It Exists

- create a Go-native replacement for the Nest sample operating model
- establish reusable agent workflows for future Go service repos
- prove a clean separation between platform guidance and sample business logic

## Current Runtime Direction

The first application phase will target:

- Gin for HTTP delivery
- Postgres with GORM
- Redis later
- Kafka and RabbitMQ after the platform and docs are mature enough

## Intended Business Demo

The later code sample will center on a `Place Order` flow involving:

- customer
- order
- product
- inventory
- promotion

That future domain is a sample consumer, not the identity of the agent template itself.
