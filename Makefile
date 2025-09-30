.PHONY: help proxy dev build test clean

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

proxy: ## Start Cloud SQL Proxy (run in separate terminal)
	@echo "Starting Cloud SQL Proxy..."
	@echo "Note: Make sure you have Cloud SQL Proxy installed and authenticated with gcloud"
	cloud-sql-proxy twigger:us-central1:dev-twigger-db1 --port 5432

dev: ## Run the application in development mode with Cloud SQL Proxy
	@echo "Starting development server..."
	CLOUD_SQL_PROXY=true go run cmd/main.go

build: ## Build the application
	go build -o bin/twigger-backend cmd/main.go

test: ## Run tests
	go test ./...

migrate-up: ## Run database migrations
	@echo "Running database migrations..."
	go run cmd/migrate/main.go up

migrate-down: ## Rollback one migration
	@echo "Rolling back one migration..."
	go run cmd/migrate/main.go down 1

migrate-reset: ## Reset database (rollback all migrations)
	@echo "Resetting database..."
	go run cmd/migrate/main.go down all

verify-backups: ## Verify Cloud SQL backup configuration and status
	@echo "Verifying Cloud SQL backups..."
	./scripts/verify-backups.sh

test-schema: ## Test comprehensive database schema functionality
	@echo "Testing database schema..."
	go run cmd/test-schema/main.go

reset-db: ## Reset database completely (removes all data)
	@echo "Resetting database..."
	go run cmd/reset-db/main.go

apply-schema: ## Apply comprehensive schema directly
	@echo "Applying comprehensive schema..."
	go run cmd/apply-schema/main.go

clean: ## Clean build artifacts
	rm -rf bin/

install-proxy: ## Install Cloud SQL Proxy
	@echo "Installing Cloud SQL Proxy..."
	curl -o cloud-sql-proxy https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.13.0/cloud-sql-proxy.windows.amd64.exe
	chmod +x cloud-sql-proxy
	mv cloud-sql-proxy $(GOPATH)/bin/ || mv cloud-sql-proxy ~/go/bin/