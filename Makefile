.PHONY: help build run test clean swagger migrate deps cert-install cert-create lint fmt vet check install-tools

help: ## Show this help message
	@printf "\033[36m%-30s\033[0m %s\n" "Target" "Description"
	@printf "\033[36m%-30s\033[0m %s\n" "------" "-----------"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[33m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building application..."
	@go mod tidy
	@go generate ./src/cmd
	@go build -o ./bin/app ./src/cmd
	@echo "Build complete: bin/app"

run: ## Run the application
	@echo "Starting application..."
	@./bin/app

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf ./bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

swagger: ## Generate swagger documentation
	@echo "Generating Swagger docs..."
	@rm -rf ./docs/
	@swag fmt
	@swag init -g ./src/cmd/app.go -o ./docs
	@echo "Fixing generated docs (removing LeftDelim/RightDelim)..."
	@sed -i.bak '/LeftDelim/d' ./docs/docs.go || sed -i '/LeftDelim/d' ./docs/docs.go
	@sed -i.bak '/RightDelim/d' ./docs/docs.go || sed -i '/RightDelim/d' ./docs/docs.go
	@rm -f ./docs/docs.go.bak
	@echo "Swagger docs generated and fixed successfully"

migrate: ## Run database migrations
	@echo "Running migrations..."
	@psql -U postgres -d gofar -f migrations/000001_create_users_table.sql
	@echo "Migrations complete"

deps: ## Install dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

cert-install: ## Install certificates
	@echo "Installing OpenSSL..."
	@sudo apt install openssl

cert-create: ## Generate RSA key pair if not exists
	@echo "Generating RSA key pair if not exists..."
	@if ! ls -AU "./etc/cert/" | read _; then \
		openssl genrsa -out ./etc/cert/id_rsa 4096 && openssl rsa -in ./etc/cert/id_rsa -pubout -out ./etc/cert/id_rsa.pub; \
	else \
		echo "Directory is not empty !!!"; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet complete"

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run
	@echo "Linting complete"

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

check: fmt vet lint ## Run all checks

install-tools: ## Install development tools
	@echo "Installing tools..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "Tools installed"

sql-postgres-create: ## Create SQL migration files for postgres
	@echo "Creating postgres SQL migration files..."
	@read -p "Enter migration name (use underscores): " name; \
		goose -dir ./etc/migrations/postgres create postgres_$${name} sql

sql-postgres-up: ## Apply up migrations for postgres
	@echo "Applying up migrations for postgres..."; \
		{ \
			stty -echo ; \
			trap 'stty echo' EXIT ; \
			read -p "Enter postgres password: " pass ; \
			stty echo ; \
			echo ; \
			goose -dir ./etc/migrations/postgres postgres "host=localhost user=postgres password=$$pass dbname=gofar sslmode=disable" up ; \
		}

sql-mysql-create: ## Create SQL migration files for mysql
	@echo "Creating mysql SQL migration files..."
	@read -p "Enter migration name (use underscores): " name; \
		goose -dir ./etc/migrations/mysql create mysql_$${name} sql

sql-mysql-up: ## Apply up migrations for mysql
	@echo "Applying up migrations for mysql..."; \
		{ \
			stty -echo ; \
			trap 'stty echo' EXIT ; \
			read -p "Enter mysql password: " pass ; \
			stty echo ; \
			echo ; \
			goose -dir ./etc/migrations/mysql mysql "host=localhost user=root password=$$pass dbname=gofar sslmode=disable" up ; \
		}