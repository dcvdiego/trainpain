.PHONY: help run build test migrate-up migrate-down docker-up docker-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the API server
	go run cmd/api/main.go

build: ## Build the API server
	go build -o bin/api cmd/api/main.go

test: ## Run tests
	go test -v ./...

docker-up: ## Start Docker Compose services
	docker-compose up -d

docker-down: ## Stop Docker Compose services
	docker-compose down

migrate-up: ## Run database migrations up
	migrate -path migrations -database "postgresql://trainpain_user:trainpain_dev_password@localhost:5432/trainpain?sslmode=disable" up

migrate-down: ## Run database migrations down
	migrate -path migrations -database "postgresql://trainpain_user:trainpain_dev_password@localhost:5432/trainpain?sslmode=disable" down

migrate-create: ## Create a new migration (usage: make migrate-create name=migration_name)
	migrate create -ext sql -dir migrations -seq $(name)

deps: ## Download dependencies
	go mod download
	go mod tidy

clean: ## Clean build artifacts
	rm -rf bin/

dev: docker-up migrate-up run ## Start development environment
