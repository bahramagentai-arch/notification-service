-include .env
export


BINARY=./bin/notifier

MODULE=$(shell go list -m)

PKGS=./...

.DEFAULT_GOAL := help

help:
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "  dev            start postgres in docker, run the service"
	@echo "  run            run the service (postgres must already be up)"
	@echo "  build          compile to $(BINARY)"
	@echo "  test           run all tests"
	@echo "  test-store     run only store integration tests (needs Docker)"
	@echo "  lint           run go vet"
	@echo ""
	@echo "  db-up          start postgres container"
	@echo "  db-down        stop postgres container"
	@echo "  db-logs        tail postgres logs"
	@echo "  migrate-up     apply all pending migrations"
	@echo "  migrate-down   roll back the last migration"
	@echo "  migrate-status show current migration version"
	@echo ""
	@echo "  clean          remove build artifacts"
	@echo ""

# -------Docker---------

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

db-logs:
	docker compose logs -f postgres


# -------Migrations---------
migrate-up:
	@echo "→ applying migrations"
	@migrate -path ./migrations -database "$(DATABASE_URL)" up
	@echo "✓ migrations applied"

migrate-down:
	@echo "→ rolling back one migration"
	@migrate -path ./migrations -database "$(DATABASE_URL)" down 1
	@echo "✓ rolled back"

migrate-status:
	@migrate -path ./migrations -database "$(DATABASE_URL)" version

.PHONY: help migrate-up migrate-down migrate-status db-up db-down db-logs