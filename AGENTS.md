# Agent Instructions for notification-service

## Project overview
- Go service module: `github.com/BahramRousta/notification-service`
- Core application entrypoint: `cmd/notifier/main.go`
- Business logic lives under `internal/notification`
- PostgreSQL persistence is implemented in `internal/notification/store`
- Domain model definitions are in `internal/notification/domain`
- Integration tests live in `test/` and use `testcontainers-go` + PostgreSQL
- Database migrations are expected under `migrations/`

## Key conventions
- `cmd/` holds executable entrypoints only.
- `internal/` contains non-public application code and packages.
- `test/` contains package-level integration tests and setup utilities.
- `Store.Create` translates PostgreSQL unique constraint violations into `notification.ErrDuplicate`.
- `Message` uses explicit fields for status, channel, priority, retry state, and provider metadata.

## What to prioritize
- Preserve package boundaries: avoid exporting internal packages unintentionally.
- Keep database access isolated in `internal/notification/store`.
- Use `context.Context` for store operations and passing cancellation.
- Follow the current domain model rather than inventing alternate message fields.

## Useful commands
- Run unit tests: `go test ./...`
- Run integration tests: `go test ./test`
- Build the module: `go test ./...` (no separate build target present)

## Notes for the agent
- There is no existing `README.md` or `.github/copilot-instructions.md` in this repo.
- `cmd/notifier/main.go` is currently a placeholder; new features should likely start in `internal/notification` and then expose them through `cmd/`.
- If you need database schema details, check `migrations/` and `test/setup.go`.
