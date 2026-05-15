.PHONY: build test vet lint run run-pg ui ui-build sqlc proto up down logs clean help

build:
	go build -o bin/cargo .

test:
	go test ./...

vet:
	go vet ./...

lint: vet
	@echo "lint OK"

run:
	go run . run

run-pg:
	go run . run --store postgres

ui:
	cd ui && bun run dev

ui-build:
	cd ui && bun run build

ui-install:
	cd ui && bun install

ui-check:
	cd ui && bun run check

sqlc:
	sqlc generate

proto:
	cd proto && buf generate

dev: up

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	docker compose down -v

help:
	@echo "Go:"
	@echo "  make build       Build API binary to bin/cargo"
	@echo "  make test        Run all Go tests"
	@echo "  make vet         Run go vet"
	@echo "  make run         Start API server (in-memory store)"
	@echo "  make run-pg      Start API server (PostgreSQL store)"
	@echo ""
	@echo "UI:"
	@echo "  make ui          Start Svelte dev server"
	@echo "  make ui-build    Production build of the SPA"
	@echo "  make ui-install  Install UI dependencies"
	@echo "  make ui-check    Typecheck Svelte/TS code"
	@echo ""
	@echo "Codegen:"
	@echo "  make sqlc        Regenerate sqlc code"
	@echo "  make proto       Regenerate gRPC/protobuf code"
	@echo ""
	@echo "Docker:"
	@echo "  make up          Build and start all services"
	@echo "  make down        Stop all services"
	@echo "  make logs        Tail logs from all services"
	@echo "  make clean       Stop services and remove volumes"
