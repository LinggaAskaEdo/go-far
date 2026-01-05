.PHONY: all help build run clean swagger migrate deps cert-install cert-create fmt vet lint test check install-tools sql-postgres-create sql-postgres-up sql-mysql-create sql-mysql-up

help: ## Show this help message
	@printf "\033[36m%-30s\033[0m %s\n" "Target" "Description"
	@printf "\033[36m%-30s\033[0m %s\n" "------" "-----------"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[33m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: build run ## Execute all steps `clean check swagger build run`

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf ./bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

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

check: fmt vet lint ## Run all checks
	@echo "All checks passed"

update: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "Dependencies updated"

swagger: ## Generate swagger documentation
	@echo "Generating Swagger docs..."
	@(swag fmt -d ./src 2>&1 | grep -v "warning: failed to get package name in dir") || true
	@(swag init -g ./src/cmd/app.go -o ./docs 2>&1 | grep -v "warning: failed to get package name in dir") || true
	@echo "Fixing generated docs (removing LeftDelim/RightDelim)..."
	@sed -i.bak '/LeftDelim/d' ./docs/docs.go 2>/dev/null || sed -i '/LeftDelim/d' ./docs/docs.go 2>/dev/null
	@sed -i.bak '/RightDelim/d' ./docs/docs.go 2>/dev/null || sed -i '/RightDelim/d' ./docs/docs.go 2>/dev/null
	@rm -f ./docs/docs.go.bak 2>/dev/null || true
	@echo "Swagger docs generated and fixed successfully"

build: clean update check swagger ## Build the application
	@echo "Building application..."
	@go mod tidy
	@go generate ./src/cmd
	@go build -o ./bin/app ./src/cmd
	@echo "Build complete: bin/app"

run: ## Run the application
	@echo "Starting application..."
	@./bin/app

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

install-tools: ## Install development tools
	@echo "Installing tools..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "Tools installed"

mon-start: ## Start monitoring stack (Grafana, Prometheus, Loki, Tempo)
	@echo "Starting monitoring stack..."
	@./start-monitoring.sh
	@echo "Monitoring stack started"

mon-stop: ## Stop monitoring stack
	@echo "Stopping monitoring stack..."
	@./stop-monitoring.sh
	@echo "Monitoring stack stopped"

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
