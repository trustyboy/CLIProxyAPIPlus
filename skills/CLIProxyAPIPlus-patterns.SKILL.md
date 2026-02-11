---
name: cliproxyapiplus-patterns
description: Coding patterns extracted from CLIProxyAPIPlus (gf branch)
version: 1.0.0
source: local-git-analysis
analyzed_commits: 200
---

# CLIProxyAPIPlus Patterns

## Commit Conventions

This repository uses **conventional commit prefixes** frequently, alongside merge commits:
- `feat:` new features (e.g., model definitions, providers)
- `fix:` bug fixes (executors, translators, config behaviors)
- `refactor:` code cleanup / simplification
- `docs:` documentation updates (README, CHANGELOG)
- `chore:` maintenance tasks (submodule updates)
- `ci:` workflow or pipeline triggers
- `build:` build scripts / pipeline adjustments
- Merge commits are common when syncing upstream/main

## Code Architecture

High-level layout:

```
cmd/                 # server entrypoints
internal/            # core application logic
  api/               # HTTP server, handlers, modules
  auth/              # provider auth
  config/            # config parsing/migration
  logging/           # logging utilities
  registry/          # model definitions registry
  runtime/executor/  # provider executors
  translator/        # request/response conversion
sdk/                 # SDK and client helpers
web/                 # frontend (git submodule)
```

Patterns observed:
- Provider integrations are split between `internal/runtime/executor/*` and `internal/translator/*`.
- Model registry changes cluster in `internal/registry/model_definitions*.go`.
- Config changes often touch `internal/config/config.go` and `config.example.yaml` together.

## Workflows

### Update web submodule
1. Update the `web/` submodule commit.
2. Commit with `chore:` or `fix:` depending on context.

### Add or update model definitions
1. Modify `internal/registry/model_definitions*.go`.
2. If needed, update related executor or translator logic.
3. Add tests under `*_test.go` when behavior changes.

### Config changes
1. Update `internal/config/config.go` or migration helpers.
2. Reflect defaults/examples in `config.example.yaml`.
3. Adjust docs if config behavior is user-facing.

## Testing Patterns

- Go tests use `*_test.go` in the same package or under `test/`.
- Tests are common for translators, executors, and config migrations.
- Changes that affect request/response formats frequently add/adjust tests.
