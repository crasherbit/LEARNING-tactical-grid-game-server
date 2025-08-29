.PHONY: run test test-integration lint migrate-up migrate-down clean

# Variables
DB_URL ?= postgres://app:app@localhost:5432/app?sslmode=disable

# Build and run
build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

# Testing
test:
	go test -v ./internal/...

test-integration:
	go test -v -tags=integration ./tests/integration/...

# Linting
lint:
	golangci-lint run

# Database migrations
migrate-up:
	goose -dir migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DB_URL)" down

# Clean
clean:
	rm -rf bin/
	go clean

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f app
