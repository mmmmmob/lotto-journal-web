# Load env vars from repo root .env.local (optional — falls back to system environment)
-include ../../.env.local
export

DOCKER_BIN := $(shell command -v docker 2>/dev/null)

ifeq ($(strip $(DOCKER_BIN)),)
DOCKER_BIN := $(HOME)/.docker/bin/docker
endif

MIGRATION_PATH := ./migrations
DB_URL         := postgres://$(DB_USERNAME):$(DB_PASSWORD)@localhost:5432/$(DB_NAME)?sslmode=disable

# ── App ──────────────────────────────────────────────────────────────────────

.PHONY: run build clean

run:
	@go tool air

build:
	@mkdir -p dist
	@go build -o dist/main app/main.go

clean:
	@rm -rf tmp dist

# ── Database ─────────────────────────────────────────────────────────────────

.PHONY: db-start db-stop

db-start:
	@$(DOCKER_BIN) compose --env-file ../../.env.local -f ../../docker-compose.yml up -d
	@echo "Waiting for database to be healthy..."
	@until [ "$$( $(DOCKER_BIN) inspect --format='{{.State.Health.Status}}' lotto-db 2>/dev/null )" = "healthy" ]; do \
		sleep 0.5; \
	done
	@echo "Database is ready!"

db-stop:
	@$(DOCKER_BIN) compose -f ../../docker-compose.yml stop -t 2

# ── Migrations ───────────────────────────────────────────────────────────────

.PHONY: migrate-up migrate-up-one migrate-down migrate-down-all migrate-version migrate-force

migrate-up:
	@migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" up

migrate-up-one:
	@migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" up 1

migrate-down:
	@migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" down 1

migrate-down-all:
	@migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" down

migrate-version:
	@migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" version

migrate-force:
	@migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" force $(N)

# ── Swagger & Mocks ──────────────────────────────────────────────────────────

.PHONY: swagger mock

swagger:
	@go run github.com/swaggo/swag/cmd/swag@latest init -g app/main.go --ot go,json,yaml

mock:
	@go run github.com/vektra/mockery/v3@v3.7.1

