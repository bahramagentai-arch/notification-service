# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

SMS notification service in Go. Module `github.com/BahramRousta/notification-service`. Early stage: persistence layer (`Store.Create`) and integration tests exist; `cmd/notifier/main.go` is still a placeholder. New features start in `internal/notification`, then get exposed through `cmd/`.

## Commands

- Run all tests: `go test ./...`
- Run integration tests only: `go test ./test`
- Run a single test: `go test ./test -run TestStore_Create`
- Run a single subtest: `go test ./test -run 'TestStore_Create/returns_ErrDuplicate'`
- Build: `go build ./...`

Integration tests require **Docker** — `test/setup_test.go` (`TestMain`) spins up a `postgres:latest` container via `testcontainers-go`, applies migrations with `golang-migrate`, then runs the suite against the live DB. No external DB setup needed; there are no pure unit tests yet.

## Architecture

Layered by package, strict boundaries:

- `internal/notification/domain` — pure model. `Message` struct plus typed enums `Status` (PENDING/SENDING/SENT/FAILED), `Channel` (SMS only), `Priority` (Low/Normal/High). Sentinel errors in `error.go` (e.g. `ErrDuplicate`). No DB or external imports here.
- `internal/notification/store` — PostgreSQL via `pgxpool` (pgx v5). `Store` wraps a connection pool (config in `db.go`). `Repository` interface + `Store.Create` in `respository.go` (note misspelled filename). `test_helper.go` holds `TruncateForTest` for integration tests.
- `cmd/notifier` — executable entrypoint (placeholder).
- `test/` — package-level integration tests.
- `migrations/` — `golang-migrate` SQL files, applied automatically by `TestMain`.

### Key conventions

- All store operations take `context.Context` first. Keep DB access isolated in `store`; don't leak pgx types into `domain` or callers.
- `Store.Create` maps the PostgreSQL unique-violation code `23505` (via `isDuplicateKey`) to `domain.ErrDuplicate`, wrapped with `%w`. Callers detect it with `errors.Is`.
- SQL uses pgx `NamedArgs` (`@name` placeholders), not positional params.
- DB schema is the source of truth for field defaults (`max_attempts` 3, `priority` 0, etc.); `Message` field semantics are documented inline in `message.go` — follow the existing model rather than inventing fields.
- Messages dedup on `idempotency_key` (UNIQUE). Dispatch ordering index is `(status, priority DESC, created_at ASC)`.

## Gotcha

The migration files are currently **swapped**: `migrations/000001_create_messages_table.up.sql` holds the DROP statements and `...down.sql` holds the CREATE statements. `migrate up` should create the schema — fix by swapping their contents before relying on migrations.
