include .env

# Database URL for migrations
DB_URL ?= postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?sslmode=$(DB_SSL_MODE)
MIGRATIONS_PATH ?= migrations

dep:
	go mod tidy

run:
	go env -w GOARCH=amd64
	go env -w GOOS=darwin
	go run main.go

build:
	go env -w GOARCH=amd64
	go env -w GOOS=linux
	go build -o event-tracking-service main.go

# Migration commands
migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1

migrate-down-all:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down -all

migrate-version:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

migrate-force:
	@read -p "Enter version to force: " version; \
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $$version

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $$name

migrate-goto:
	@read -p "Enter version to goto: " version; \
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" goto $$version
