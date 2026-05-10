---
tags: [decision, sample, tooling, node]
date: 2026-04-25
status: active
repo: tolvi-examples
ticket: none
---

## Why

Build failures across the team were tracing back to subtle Node version differences. A developer on Node 20.10 produced a lockfile that another developer on Node 20.18 could not reproduce; a CI runner on Node 18 silently accepted code that used a 20-only API. We wanted the Node version (and the package manager version) to be declared per-repo and enforced automatically when anyone — human or CI — entered the repo.

## How

We adopted Volta for Node and package-manager version management. Each repo's `package.json` carries a `volta` block pinning the exact Node and package manager versions. Volta hooks shell into PATH management so the right `node` binary is selected the moment a developer `cd`s into the repo, with no manual `nvm use`. CI installs Volta as a first-step bootstrap and lets the same `package.json` block drive the runner's tool versions. Documentation in the repo README points new developers at a single `curl | bash` install command.

## Outcome

Version-related build failures have stopped. Onboarding documentation collapsed from a half page about installing Node and pinning a version to a single command. The Volta footprint is small enough that it has not interfered with developers who already use other version managers for non-Tolvi work.
