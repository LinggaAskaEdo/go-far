# Variables
BINARY_NAME    := app
BIN_DIR        := ./bin
SRC_DIR        := ./src
CMD_PATH       := $(SRC_DIR)/cmd/app.go
DOCS_DIR       := ./docs
COVERAGE_OUT   := coverage.out
COVERAGE_HTML  := coverage.html

# Go flags
GO             := go
GOFLAGS        := -v
LDFLAGS        := -s -w -X main.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

# Tools
SWAG           := swag
GOLANGCI_LINT  := golangci-lint
GOOSE          := goose

# Colors
BLUE           := \033[36m
YELLOW         := \033[33m
GREEN          := \033[32m
WHITE          := \033[37m
RESET          := \033[0m
COMMA          := ,

.PHONY: all help build run clean swagger migrate deps cert-install cert-create fmt vet lint test check install-tools update sql-postgres-create sql-postgres-up sql-mysql-create sql-mysql-up mon-start mon-stop

## Show this help message
help:
	@echo ""
	@printf "$(BLUE)Go-FAR Project - Makefile Commands$(RESET)\n"
	@echo ""
	@printf "$(YELLOW)General:$(RESET)\n"
	@printf "  $(GREEN)make help$(RESET)                  Show this help message\n"
	@printf "  $(GREEN)make all$(RESET)                   Clean, build, and run app\n"
	@printf "  $(GREEN)make deps$(RESET)                  Download and install dependencies\n"
	@printf "  $(GREEN)make update$(RESET)                Update all dependencies to latest\n"
	@printf "  $(GREEN)make install-tools$(RESET)         Install dev tools (swag, lint, etc)\n"
	@echo ""
	@printf "$(YELLOW)Build & Run:$(RESET)\n"
	@printf "  $(GREEN)make build$(RESET)                 Build application with optimizations\n"
	@printf "  $(GREEN)make run$(RESET)                   Run the built application\n"
	@printf "  $(GREEN)make clean$(RESET)                 Remove build artifacts\n"
	@echo ""
	@printf "$(YELLOW)Code Quality:$(RESET)\n"
	@printf "  $(GREEN)make check$(RESET)                 Run all checks (fmt, vet, lint)\n"
	@printf "  $(GREEN)make fmt$(RESET)                   Format code with go fmt\n"
	@printf "  $(GREEN)make vet$(RESET)                   Run go vet for common issues\n"
	@printf "  $(GREEN)make lint$(RESET)                  Run golangci-lint\n"
	@printf "  $(GREEN)make test$(RESET)                  Run tests with coverage report\n"
	@echo ""
	@printf "$(YELLOW)Documentation:$(RESET)\n"
	@printf "  $(GREEN)make swagger$(RESET)               Generate Swagger API docs\n"
	@echo ""
	@printf "$(YELLOW)Database:$(RESET)\n"
	@printf "  $(GREEN)make migrate$(RESET)               Run database migrations (postgres)\n"
	@printf "  $(GREEN)make sql-postgres-create$(RESET)   Create new postgres migration\n"
	@printf "  $(GREEN)make sql-postgres-up$(RESET)       Apply postgres migrations\n"
	@printf "  $(GREEN)make sql-mysql-create$(RESET)      Create new mysql migration\n"
	@printf "  $(GREEN)make sql-mysql-up$(RESET)          Apply mysql migrations\n"
	@echo ""
	@printf "$(YELLOW)Certificates:$(RESET)\n"
	@printf "  $(GREEN)make cert-install$(RESET)          Install OpenSSL\n"
	@printf "  $(GREEN)make cert-create$(RESET)           Generate RSA key pair (4096-bit)\n"
	@echo ""
	@printf "$(YELLOW)Monitoring:$(RESET)\n"
	@printf "  $(GREEN)make mon-start$(RESET)             Start monitoring stack\n"
	@printf "  $(GREEN)make mon-stop$(RESET)              Stop monitoring stack\n"
	@echo ""

## Execute build and run
all: clean deps check swagger build run

## Clean build artifacts and coverage files
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)/
	@rm -rf logs/
	@rm -f $(COVERAGE_OUT) $(COVERAGE_HTML)
	@printf "$(BLUE)Clean complete$(RESET)\n"

## Format code
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@echo "Format complete"

## Run go vet
vet:
	@echo "Running go vet..."
	@$(GO) vet ./...
	@echo "Vet complete"

## Run linter
lint:
	@echo "Running linter..."
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) run && printf "Linting complete\n"; \
	else \
		printf "$(YELLOW)golangci-lint not installed. Install with: make install-tools$(RESET)\n"; \
		echo "Skipping lint..."; \
	fi

## Run all checks (fmt, vet, lint)
check: fmt vet lint
	@printf "$(BLUE)All checks passed$(RESET)\n"

## Update dependencies to latest versions
update:
	@echo "Updating dependencies..."
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@echo "Dependencies updated"

## Generate swagger documentation
swagger:
	@echo "Generating Swagger docs..."
	@($(SWAG) fmt -d $(SRC_DIR) 2>&1 | grep -v "warning: failed to get package name in dir") || true
	@($(SWAG) init -g $(CMD_PATH) -o $(DOCS_DIR) 2>&1 | grep -v "warning: failed to get package name in dir") || true
	@echo "Fixing generated docs (removing LeftDelim/RightDelim)..."
	@sed -i.bak '/LeftDelim/d' $(DOCS_DIR)/docs.go 2>/dev/null || sed -i '/LeftDelim/d' $(DOCS_DIR)/docs.go 2>/dev/null
	@sed -i.bak '/RightDelim/d' $(DOCS_DIR)/docs.go 2>/dev/null || sed -i '/RightDelim/d' $(DOCS_DIR)/docs.go 2>/dev/null
	@rm -f $(DOCS_DIR)/docs.go.bak 2>/dev/null || true
	@printf "$(BLUE)Swagger docs generated and fixed successfully$(RESET)\n"

## Run tests with coverage
test:
	@echo "Running tests..."
	@$(GO) test -v -race -coverprofile=$(COVERAGE_OUT) ./...
	@$(GO) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@printf "$(BLUE)Tests complete. Coverage report: $(COVERAGE_HTML)$(RESET)\n"

## Build the application with optimizations
build: clean check swagger
	@echo "Building application..."
	@$(GO) mod tidy
	@$(GO) generate $(SRC_DIR)/cmd
	@mkdir -p $(BIN_DIR)
	@$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) $(SRC_DIR)/cmd
	@printf "$(BLUE)Build complete: $(BIN_DIR)/$(BINARY_NAME)$(RESET)\n"

## Run the application
run:
	@echo "Starting application..."
	@$(BIN_DIR)/$(BINARY_NAME)

## Run database migrations (postgres default)
migrate:
	@echo "Running migrations..."
	@psql -U postgres -d gofar -f etc/migrations/000001_create_users_table.sql
	@echo "Migrations complete"

## Download and tidy dependencies
deps:
	@echo "Installing dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "Dependencies installed"

## Install OpenSSL for certificates
cert-install:
	@echo "Installing OpenSSL..."
	@sudo apt install -y openssl

## Generate RSA key pair if not exists
cert-create:
	@echo "Generating RSA key pair if not exists..."
	@if [ ! -f ./etc/cert/id_rsa ]; then \
		mkdir -p ./etc/cert && \
		openssl genrsa -out ./etc/cert/id_rsa 4096 && \
		openssl rsa -in ./etc/cert/id_rsa -pubout -out ./etc/cert/id_rsa.pub; \
		echo "$(BLUE)Key pair generated successfully$(RESET)\n"; \
	else \
		echo "Key pair already exists"; \
	fi

## Install development tools
install-tools:
	@echo "Installing tools..."
	@$(GO) install github.com/swaggo/swag/cmd/swag@latest
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install github.com/pressly/goose/v3/cmd/goose@latest
	@printf "$(BLUE)Tools installed$(RESET)\n"

## Start monitoring stack (Grafana, Prometheus, Loki, Tempo)
mon-start:
	@echo "Starting monitoring stack..."
	@./start-monitoring.sh
	@echo "Monitoring stack started"

## Stop monitoring stack
mon-stop:
	@echo "Stopping monitoring stack..."
	@./stop-monitoring.sh
	@echo "Monitoring stack stopped"

## Create SQL migration files for postgres
sql-postgres-create:
	@echo "Creating postgres SQL migration files..."
	@read -p "Enter migration name (use underscores): " name; \
		$(GOOSE) -dir ./etc/migrations/postgres create postgres_$${name} sql

## Apply up migrations for postgres
sql-postgres-up:
	@echo "Applying up migrations for postgres..."; \
		{ \
			stty -echo ; \
			trap 'stty echo' EXIT ; \
			read -p "Enter postgres password: " pass ; \
			stty echo ; \
			echo ; \
			$(GOOSE) -dir ./etc/migrations/postgres postgres "host=localhost user=postgres password=$$pass dbname=go_far sslmode=disable" up ; \
		}

## Create SQL migration files for mysql
sql-mysql-create:
	@echo "Creating mysql SQL migration files..."
	@read -p "Enter migration name (use underscores): " name; \
		$(GOOSE) -dir ./etc/migrations/mysql create mysql_$${name} sql

## Apply up migrations for mysql
sql-mysql-up:
	@echo "Applying up migrations for mysql..."; \
		{ \
			stty -echo ; \
			trap 'stty echo' EXIT ; \
			read -p "Enter mysql password: " pass ; \
			stty echo ; \
			echo ; \
			$(GOOSE) -dir ./etc/migrations/mysql mysql "host=localhost user=root password=$$pass dbname=go_far sslmode=disable" up ; \
		}
