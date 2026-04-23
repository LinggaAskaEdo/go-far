# Variables
BINARY_NAME		:= app
BIN_DIR        	:= ./bin
SRC_DIR        	:= ./cmd/api
CMD_PATH       	:= $(SRC_DIR)/main.go
DOCS_DIR       	:= ./api/openapi
COVERAGE_OUT   	:= coverage.out
COVERAGE_HTML  	:= coverage.html

# Go flags
GO             	:= go
GOBIN 			:= $(shell go env GOPATH)/bin
GOFLAGS        	:= -v
LDFLAGS        	:= -s -w -X main.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

# Tools
SWAG           	:= swag
GOLANGCI_LINT  	:= golangci-lint
GOOSE          	:= goose

# Colors
BLUE           	:= \033[36m
YELLOW         	:= \033[33m
GREEN          	:= \033[32m
WHITE          	:= \033[37m
RESET          	:= \033[0m
COMMA          	:= ,

.PHONY: all help build run clean swagger migrate deps lint test install-tools update sql-postgres-create sql-postgres-up sql-mysql-create sql-mysql-up monitoring-start monitoring-stop kill benchmark

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
	@printf "  $(GREEN)make kill$(RESET)                  Kill app running on port 8181\n"
	@echo ""
	@printf "$(YELLOW)Code Quality:$(RESET)\n"
	@printf "  $(GREEN)make lint$(RESET)                  Run golangci-lint\n"
	@printf "  $(GREEN)make test$(RESET)                  Run tests with coverage report\n"
	@echo ""
	@printf "$(YELLOW)Documentation:$(RESET)\n"
	@printf "  $(GREEN)make swagger$(RESET)               Generate Swagger API docs\n"
	@echo ""
	@printf "$(YELLOW)Database:$(RESET)\n"
	@printf "  $(GREEN)make sql-postgres-create$(RESET)   Create new postgres migration\n"
	@printf "  $(GREEN)make sql-postgres-up$(RESET)       Apply postgres migrations\n"
	@echo ""
	@printf "$(YELLOW)Monitoring:$(RESET)\n"
	@printf "  $(GREEN)make monitoring-start$(RESET)             Start monitoring stack\n"
	@printf "  $(GREEN)make monitoring-stop$(RESET)              Stop monitoring stack\n"
	@echo ""
	@printf "$(YELLOW)Benchmark:$(RESET)\n"
	@printf "  $(GREEN)make benchmark$(RESET)                Run API benchmark with Apache Bench\n"
	@echo ""

## Execute build and run
all: clean deps swagger build run

## Clean build artifacts and coverage files
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)/
	@rm -rf logs/
	@rm -f $(COVERAGE_OUT) $(COVERAGE_HTML)
	@printf "$(BLUE)Clean complete$(RESET)\n"

## Run linter
lint:
	@echo "Running linter..."
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) run; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Install with: make install-tools$(RESET)"; \
		exit 1; \
	fi
	@echo "Lint complete"

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
build: clean swagger
	@echo "Building application..."
	@$(GO) mod tidy
	@$(GO) generate $(SRC_DIR)
	@mkdir -p $(BIN_DIR)
	@$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) $(SRC_DIR)
	@printf "$(BLUE)Build complete: $(BIN_DIR)/$(BINARY_NAME)$(RESET)\n"

## Run the application
run:
	@echo "Starting application..."
	@$(BIN_DIR)/$(BINARY_NAME)

## Kill app running on port 8181
kill:
	@echo "Killing app on port 8181..."
	@pkill -f "./bin/app" || true
	@echo "Done"

## Download and tidy dependencies
deps:
	@echo "Installing dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "Dependencies installed"

## Install development tools
install-tools:
	@echo "Installing tools..."
	@$(GO) install github.com/swaggo/swag/cmd/swag@latest
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@$(GO) install github.com/pressly/goose/v3/cmd/goose@latest
	@$(GO) install mvdan.cc/gofumpt@latest
	@$(GO) install github.com/daixiang0/gci@latest
	@$(GO) install golang.org/x/tools/cmd/goimports@latest
	@$(GO) install honnef.co/go/tools/cmd/staticcheck@latest
	@$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest
	@$(GO) install -v github.com/go-critic/go-critic/cmd/go-critic@latest
	@$(GO) install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
	@printf "$(BLUE)Tools installed$(RESET)\n"

## Start monitoring stack (Grafana, Prometheus, Loki, Tempo)
monitoring-start:
	@echo "Starting monitoring stack..."
	@./scripts/start-monitoring.sh
	@echo "Monitoring stack started"

## Stop monitoring stack
monitoring-stop:
	@echo "Stopping monitoring stack..."
	@./scripts/stop-monitoring.sh
	@echo "Monitoring stack stopped"

## Create SQL migration files for postgres
sql-postgres-create:
	@echo "Creating postgres SQL migration files..."
	@read -p "Enter migration name (use underscores): " name; \
		$(GOOSE) -dir ./db/migrations create postgres_$${name} sql

## Apply up migrations for postgres
sql-postgres-up:
	@echo "Applying up migrations for postgres..."; \
		{ \
			stty -echo ; \
			trap 'stty echo' EXIT ; \
			read -p "Enter postgres password: " pass ; \
			stty echo ; \
			echo ; \
			$(GOOSE) -dir ./db/migrations postgres "host=localhost user=postgres password=$$pass dbname=go-far sslmode=disable" up ; \
		}

## Run API benchmark with Apache Bench
benchmark:
	@echo "Running API benchmark..."
	@./scripts/benchmark.sh