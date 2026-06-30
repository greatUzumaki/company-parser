# Company Parser — developer tasks.
# DB and Redis are expected to be already running (user's Docker).
# Copy .env.example -> .env and adjust before running.

include .env
export

.PHONY: dev build test lint migrate-up migrate-down sqlc tidy frontend-dev

## Backend ---------------------------------------------------------------

dev: ## Run the API server with live env
	cd backend && go run ./cmd/server

build: ## Compile the server binary
	cd backend && go build -o ../bin/server ./cmd/server

test: ## Run all Go tests
	cd backend && go test ./...

lint: ## Vet + format check
	cd backend && go vet ./... && gofmt -l .

tidy: ## Sync go.mod
	cd backend && go mod tidy

## Database --------------------------------------------------------------

migrate-up: ## Apply migrations
	migrate -path backend/migrations -database "$(DATABASE_URL)" up

migrate-down: ## Roll back one migration
	migrate -path backend/migrations -database "$(DATABASE_URL)" down 1

sqlc: ## Regenerate type-safe DB code
	cd backend && sqlc generate

## Frontend --------------------------------------------------------------

frontend-dev: ## Run the Next.js dev server
	cd frontend && npm run dev
