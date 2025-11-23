MIGRATIONS_DIR := internal/infrastructure/db/migrations
# default dsn (if there are no .env)
PG_DSN := postgres://user:password@localhost:5432/pr_reviewer?sslmode=disable

ifneq (,$(wildcard .env))
    include .env
    export
endif

.PHONY: help up down migrate-create migrate-up migrate-status 
# commands list
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' 
# start app
up:
	docker-compose up -d --build
# stop app
down:
	docker-compose down
# check logs
logs: 
	docker-compose logs -f

# go install github.com/pressly/goose/v3/cmd/goose@latest
# create new migration: make migrate-create name=init_schema
migrate-create: 
	goose -dir $(MIGRATIONS_DIR) create $(name) sql
migrate-up: 
	goose -dir $(MIGRATIONS_DIR) postgres "$(PG_DSN)" up
migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(PG_DSN)" down
migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(PG_DSN)" status


.PHONY: lint clean
lint:
	golangci-lint run ./...
clean:
	rm -f pr-reviewer-assigner