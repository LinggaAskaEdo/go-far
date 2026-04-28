# Go-Far - Golang Standard Project Layout CRUD API

A production-ready RESTful API built with Go following [golang-standards/project-layout](https://github.com/golang-standards/project-layout), featuring PostgreSQL, Redis, JWT authentication, role-based rate limiting, and OpenTelemetry tracing.

## Version: 1.18.0

## 🚀 Features

- **Go Standard Layout** - Separation of concerns with handlers, services, and repositories
- **REST API** - Built with Go's native `net/http` (Go 1.25+ pattern matching)
- **Database** - PostgreSQL with pgx (MySQL supported)
- **Caching** - Redis
- **Authentication** - JWT with HS-256 signing, refresh tokens, and role-based access
- **Role-Based Rate Limiting** - Per-role sliding window rate limiting via Redis Lua script (atomic, race-condition free)
- **Observability** - OpenTelemetry tracing with OTLP exporter, Prometheus metrics on separate port
- **Scheduled Jobs** - Cron-based job scheduler
- **Circuit Breaker** - Per-job circuit breaker using failsafe-go for external API protection
- **API Documentation** - Swagger/OpenAPI 2.0
- **Graceful Shutdown** - Proper cleanup of resources
  - **Many-to-Many Relationships** - Users and cars via junction table
  - **Generic Query Decoder** - Reflection-based HTTP query decoder for reusable filter handling
- **SQL Query Cleaner** - Utility to clean SQL queries for logging (masks sensitive values like passwords)
- **Go Templating** - Native Go templates for dynamic SQL query loading
- **Custom Validators** - Separate validator configuration package for reusable validation logic
- **API Benchmarking** - Built-in Apache Bench integration for performance testing

## 📁 Project Structure

```text
go-far/
├── cmd/api/                    # Application entry point
│   └── main.go                 # Main application bootstrap
├── internal/                   # Application packages
│   ├── app/                    # App bootstrap & initialization
│   ├── config/                 # Configuration management
│   ├── handler/                # HTTP & scheduler handlers
│   │   ├── http/               # REST API handlers
│   │   │   ├── auth_handler.go
│   │   │   ├── car_handler.go
│   │   │   ├── helper.go
│   │   │   ├── router.go
│   │   │   └── user_handler.go
│   │   └── scheduler/          # Cron job handlers
│   ├── infra/                  # Infrastructure & configuration modules
│   │   ├── database/           # PostgreSQL connection (pgx)
│   │   ├── grace/              # Graceful shutdown
│   │   ├── http/               # HTTP client & mux
│   │   ├── logger/             # Zerolog logger
│   │   ├── middleware/         # Request middleware (CORS, rate limiting)
│   │   ├── query/              # SQL query loader (Go templates)
│   │   ├── redis/              # Redis client
│   │   ├── scheduler/          # Cron scheduler
│   │   ├── token/              # JWT token management (HS-256)
│   │   ├── tracer/             # OpenTelemetry tracer
│   │   └── validator/          # Custom validators
│   ├── model/                  # Data contracts
│   │   ├── dto/                # Data Transfer Objects
│   │   ├── entity/             # Domain entities
│   │   └── errors/             # Error handling
│   ├── preference/             # Constants & shared values
│   ├── repository/             # Data access layer
│   │   ├── car/                # Car repository
│   │   └── user/               # User repository
│   ├── service/                # Business logic layer
│   │   ├── car/                # Car service
│   │   └── user/               # User service
│   └── util/                   # Utility functions
├── api/                        # Generated API docs
│   └── openapi/                # Swagger documentation
├── configs/                    # Application configuration
│   └── queries/             # SQL queries with Go templates
├── db/                         # Database migrations
├── deployments/                # Deployment configs
├── scripts/                    # Utility scripts
├── Makefile                    # Build & run commands
└── go.mod                      # Go module definition
```

## 🛠️ Tech Stack

| Component       | Technology                        | Dependencies URL                                     |
| --------------- | --------------------------------- | ---------------------------------------------------- |
| Framework       | Go 1.25+ `net/http`               | <https://pkg.go.dev/net/http>                        |
| Database        | PostgreSQL (MySQL supported)      | <https://www.postgresql.org>                         |
| DB Driver       | pgx (jackc/pgx)                   | <https://github.com/jackc/pgx>                       |
| Cache           | Redis                             | <https://redis.io>                                   |
| Redis Driver    | go-redis (redis/go-redis)         | <https://github.com/redis/go-redis>                  |
| Auth            | JWT (HS-256) + bcrypt             | <https://github.com/golang-jwt/jwt>                  |
| Logging         | Zerolog                           | <https://github.com/rs/zerolog>                      |
| Logging File    | lumberjack (natefinch/lumberjack) | <https://github.com/natefinch/lumberjack>            |
| Tracing         | OpenTelemetry                     | <https://opentelemetry.io>                           |
| Metric          | Prometheus (client_golang)        | <https://github.com/prometheus/client_golang>        |
| Scheduler       | robfig/cron/v3                    | <https://github.com/robfig/cron>                     |
| Circuit Breaker | failsafe-go                       | <https://github.com/failsafe-go/failsafe>            |
| Validation      | go-playground/validator           | <https://github.com/go-playground/validator>         |
| Docs            | Swagger (swaggo)                  | <https://github.com/swaggo/swag>                     |
| SQL Templating  | Go text/template                  | <https://pkg.go.dev/text/template>                   |
| Query Decoder   | Custom reflection-based           | -                                                    |
| Benchmark       | Apache Bench (ab)                 | <https://httpd.apache.org/docs/2.4/programs/ab.html> |

## 📋 API Endpoints

### Auth

| Method    | Endpoint           | Description                          |
| --------  | ------------------ | ------------------------------------ |
| POST      | `/auth/register`   | Register a new user                  |
| POST      | `/auth/login`      | Login and get tokens                 |
| POST      | `/auth/refresh`    | Refresh access token                 |

### Health Check

| Method   | Endpoint      | Description                                    |
| -------- | ------------- | ---------------------------------------------- |
| GET      | `/health`     | Health check (public)                          |
| GET      | `/ready`      | Readiness check with DB & Redis ping (public)  |

### Users

| Method | Endpoint          | Description            |
|--------|-------------------|------------------------|
| POST   | `/users`          | Create user            |
| GET    | `/users/{id}`     | Get user by ID         |
| GET    | `/users`          | List users (paginated) |
| PUT    | `/users/{id}`     | Update user            |
| DELETE | `/users/{id}`     | Delete user            |

### Cars

| Method | Endpoint                        | Description                              |
|--------|---------------------------------|------------------------------------------|
| POST   | `/cars`                         | Create car                               |
| POST   | `/cars/bulk`                    | Create multiple cars                     |
| GET    | `/cars/{id}`                    | Get car by ID                            |
| GET    | `/cars/{id}/owner`              | Get car with owner details               |
| PUT    | `/cars/{id}`                    | Update car                               |
| DELETE | `/cars/{id}`                    | Delete car                               |
| POST   | `/cars/{id}/transfer`           | Transfer ownership                       |
| PUT    | `/cars/availability`            | Bulk update availability                 |
| GET    | `/users/{user_id}/cars`         | List cars by user (IDOR protected)       |
| GET    | `/users/{user_id}/cars/count`   | Count cars by user (IDOR protected)      |

### Swagger Documentation

Access Swagger UI at: `http://localhost:8181/swagger/index.html`

## 🔐 User Roles

| Role    | Description                 |
|---------|---------------------------- |
| `admin` | Full access to all features |
| `user`  | Standard user access        |
| `guest` | Limited read-only access    |

Roles are assigned during registration and stored as a PostgreSQL enum type.

## 🔒 Security

- **IDOR Protection** - User-scoped car endpoints (`/users/{user_id}/cars`) enforce ownership checks. Non-admin users can only access their own resources.
- **Readiness Probe** - The `/ready` endpoint actively checks database and Redis connectivity, returning `503` if dependencies are unavailable.

## ⚙️ Configuration

Edit `configs/config.yaml` or use environment variables:

```yaml
app:
  name: go-far
  version: 1.18.0
  environment: development  # development/staging/production)
  shutdown_timeout: 5s

http:
  server:
    app_name: go-far
    mode: release # debug, release
    port: 8181
    write_timeout: 10s
    read_timeout: 10s
    idle_timeout: 60s
    max_body_bytes: 1048576 # 1MB (1 << 20), typical: 1-10MB for REST APIs
  metrics_server:
    mode: release # debug, release
    port: 9191
    write_timeout: 10s
    read_timeout: 10s
    idle_timeout: 60s
    max_body_bytes: 1048576 # 1MB
  client:
    enabled: true
    max_idle_conns: 100
    max_idle_conns_per_host: 10
    max_conns_per_host: 20
    idle_conn_timeout: 90s
    dial_timeout: 30s
    keep_alive: 30s
    tls_handshake_timeout: 10s
    response_header_timeout: 10s
    expect_continue_timeout: 1s
    disable_compression: false
    timeout: 10s
    circuit_breaker:
      max_retries: 2
      backoff_min: 100ms
      backoff_max: 2s

database: # postgres, mysql, oracle
  postgres:
    enabled: true
    driver: postgres
    host: localhost
    port: 5432
    user: postgres
    password: ${POSTGRES_DOCKER_PASSWORD}
    dbname: go-far
    sslmode: false
    max_open_conns: 25
    max_idle_conns: 5
    conn_max_lifetime: 1h
    conn_max_idle_time: 30m

redis:
  enabled: true
  network: tcp
  address: "localhost:6379"
  password: ""
  cache_ttl: 60s
  max_retries: 3
  min_retry_backoff: 8ms
  max_retry_backoff: 512ms
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s
  pool_size: 10
  min_idle_conns: 5
  max_idle_conns: 5
  max_active_conns: 10
  pool_timeout: 30s

logger:
  enabled: true
  level: debug # debug, info, warn, error
  format: json # json, console
  output: stdout # stdout, file
  path: ./logs/app.log
  max_size: 100 # megabytes
  max_backups: 7
  max_age: 30 # days
  compress: true # disabled by default

middleware:
  public_paths:
    - /health
    - /ready
    - /swagger
    - /swagger/*
    - /auth/*
    - /metrics
    - /debug/*
    - /favicon.ico
  auth_rate_limit:
    command: "3-M"  # 3 requests per minute per IP
    limit: 3
  rate_limiter:
    command: "1-S" # 1 request per second
    limit: 500 # max 500 clients
  role_rate_limit:
    admin:
      command: "1-M"
      limit: 10000000
    user:
      command: "1-M"
      limit: 5
    guest:
      command: "1-M"
      limit: 1

scheduler:
  enabled: true
  jobs:
    user_generator:
      enabled: false
      random_user_url: "https://randomuser.me/api/"
      cron: "0 0 */1 * * *" # Every 1 hour
      batch_size: 5
      min_age: 18
      max_age: 80
    car_generator:
      enabled: false
      nhtsa_api_url: "https://vpic.nhtsa.dot.gov/api/vehicles"
      cron: "0 */30 * * * *" # Every 30 minutes
      batch_size: 3
      min_year: 2015
      max_year: 2025

token:
  expired_token: 5m
  expired_refresh_token: 15m

queries:
  path: ./configs/queries/

tracer:
  enabled: true
  endpoint: "localhost:4317"
  protocol: "grpc"

metric:
  enabled: true

pyroscope:
  enabled: false
```

### Scheduler Jobs

| Job              | Schedule          | Description                                      | Enabled |
| ---------------- | ----------------- | ------------------------------------------------ | ------- |
| `user_generator` | Every 1 hour      | Generates random users from randomuser.me API    | false   |
| `car_generator`  | Every 30 minutes  | Generates random cars from NHTSA API             | false   |

### Environment Variables

```bash
# Server
export SERVER_PORT=8181

# Database
export POSTGRES_DOCKER_PASSWORD=your_password
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_NAME=go_far

# MySQL (if used)
export MYSQL_DOCKER_PASSWORD=your_mysql_password
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_USER=root
export MYSQL_DB_NAME=go_far

# JWT Secret (set in config or environment)
export JWT_SECRET_GO_FAR=your-64-byte-hex-secret

# Redis
export REDIS_ADDRESS=localhost:6379
export REDIS_PASSWORD=

# CORS
export ALLOWED_ORIGINS=https://example.com,https://app.example.com

# Tracing
export TRACER_ENDPOINT=localhost:4317

# Logging
export LOG_LEVEL=info
```

## 🚦 Getting Started

### Prerequisites

- Go 1.25
- PostgreSQL 14+
- Redis 7+
- Make

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/yourusername/go-far.git
   cd go-far
   ```

2. **Install dependencies**

   ```bash
   make deps
   ```

3. **Install development tools**

   ```bash
   make install-tools
   ```

4. **Setup JWT secret**

   ```bash
   # Set JWT_SECRET_GO_FAR environment variable or in configs/config.yaml
   export JWT_SECRET_GO_FAR=$(openssl rand -hex 64)
   ```

5. **Setup database**

   ```bash
   createdb go_far
   make sql-postgres-up
   ```

6. **Update configuration**

   ```bash
   cp configs/config.yaml configs/config.yaml.local
   # Edit configs/config.yaml.local with your settings
   ```

7. **Run the application**

   ```bash
   make all          # Clean, build, and run
   # Or individually
   make build        # Build the binary
   make run          # Run the application
   ```

### Make Commands

Run `make help` to see all available commands:

```bash
make help              # Show this help message
make all               # Clean, build, and run app
make deps              # Download and install dependencies
make update            # Update all dependencies to latest
make install-tools      # Install dev tools (swag, lint, etc)

make build             # Build application with optimizations
make run               # Run the built application
make clean             # Remove build artifacts
make kill              # Kill app running on port 8181

make lint              # Run golangci-lint
make test             # Run tests with coverage report

make swagger          # Generate Swagger API docs

make sql-postgres-create   # Create new postgres migration
make sql-postgres-up     # Apply postgres migrations

make monitoring-start  # Start monitoring stack
make monitoring-stop   # Stop monitoring stack

make benchmark       # Run API benchmark with Apache Bench
```

## 📝 Example Requests

### Register User

```bash
curl -X POST http://localhost:8181/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securePass123!",
    "age": 30,
    "role": "user"
  }'
```

### Login

```bash
curl -X POST http://localhost:8181/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securePass123!"
  }'
```

### Create User

```bash
curl -X POST http://localhost:8181/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password": "securePass456!",
    "age": 25,
    "role": "user"
  }'
```

### Create Car

```bash
curl -X POST http://localhost:8181/cars \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-uuid",
    "brand": "Toyota",
    "model": "Camry",
    "year": 2024,
    "color": "Blue",
    "license_plate": "ABC123"
  }'
```

### List Users with Pagination

```bash
curl -X GET "http://localhost:8181/users?page=1&page_size=10&sort_by=name&sort_dir=asc"
```

### Generic Query Decoder Usage

The `util.DecodeURL` function automatically decodes HTTP query parameters to DTOs using struct tags:

```go
// DTO with param or form tags
type UserFilter struct {
    Name     string `param:"name"`
    Email    string `param:"email"`
    MinAge   int    `param:"min_age"`
    MaxAge   int    `param:"max_age"`
    Page     int64  `param:"page"`
    PageSize int64  `param:"page_size"`
    SortBy   string `param:"sort_by"`
    SortDir  string `param:"sort_dir"`
}

// Handler usage - single line instead of 30+ lines
filter := util.DecodeURL[dto.UserFilter](r.URL.Query())
```

Supports: `string`, `int`, `int64`, `float64`, `bool`, `time.Time`, `[]string`

## 🔒 Authentication

The API uses JWT with HS-256 signing. Include the token in the Authorization header:

```text
Authorization: Bearer <your_jwt_token>
```

Tokens are generated on `/auth/login` and can be refreshed using `/auth/refresh`.

## 📊 Observability

### Logging

Logs are written to stdout with rotation. Format is configurable (JSON/Console).

### Tracing

OpenTelemetry traces are exported to `localhost:4317` (OTLP/gRPC). Configure your collector accordingly.

### Metrics

Metrics collection is available via OpenTelemetry but disabled by default (set `tracer.enabled: true` in config).

## 🧪 Testing

```bash
go test ./...
go test ./... -cover
```

## 📄 License

Apache 2.0 - See [LICENSE](LICENSE) for details.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📞 Support

- Issues: GitHub Issues
- Email: <lemp.otis@gmail.com>

### v1.18.0

- Added metrics server on separate port (9191) for Prometheus scraping
- Added metrics config: `http.metrics_server` in config.yaml
- Added Go runtime and process collectors (CPU, memory, goroutines)
- Added HTTP request counter (`http_server_requests_total`)
- Added dedicated `/metrics` endpoint exposed on port 9191
- Separated metrics server from main API server
- Fixed trace not exported to Tempo (missing app_name in http.server config)
- Added app_name validation to ensure OTel handler is created

### v1.17.0

- Added PostgreSQL pool metrics collector (`pgx_pool_*`) for connection monitoring
- Added Redis pool metrics collector (`redis_pool_*`) for connection monitoring
- Added HTTP request duration metrics (`http_request_duration_seconds`)
- Implemented custom Prometheus registry to avoid conflicts with pyroscope
- Added `client` label to Redis metrics to distinguish between apps/auth/limiter instances
- Refactored metrics initialization to reduce cognitive complexity

### v1.16.0

- Added Prometheus metrics via `/metrics` endpoint for database pool and Redis monitoring
- Added Grafana dashboard for go-far in `deployments/monitoring/grafana/dashboards/`
- Added HTTP/HTTPS OTLP protocol support for tracer (configurable via `tracer.protocol`)
- Added configurable service name and version for OpenTelemetry (`tracer.name`, `tracer.version`)
- Added `go.opentelemetry.io/otel/sdk/trace` composite text propagator (TraceContext + Baggage)
- Improved tracer error handling (graceful fallback instead of panic)
- Refactored middleware: consolidated swagger/metrics/debug path handling with `shouldSkipAuthAndLog()`
- Updated public paths: added `/metrics`, `/debug/*`; removed `/debug/statsviz`
- Enhanced `.golangci.yml` with new linters (wastedassign, copyloopvar, prealloc, mirror, nilnil, protogetter, decorder)
- Improved linter documentation and settings with gocritic checks
- Organized code structure per decorder linter: types/constants before functions in error_code.go, user.go, response.go, request_filter.go

### v1.15.0

- Added trace/span/request_id logging to scheduler jobs
- Conditional logging: trace_id/span_id printed when tracing enabled, req_id always printed
- Jobs now propagate context with trace IDs from scheduler to job handlers

### v1.14.0

- Added trace/span/request_id logging to scheduler jobs
- Added GenerateTraceID/GenerateSpanID functions to middleware
- Conditional logging: trace_id/span_id printed when tracing enabled
- Jobs propagate context with trace IDs from scheduler to handlers
- Updated golangci.yml with additional linters (predeclared, misspell, usetesting)

### v1.13.0

- Added Pyroscope profiler integration for continuous profiling
- Restructured config with dedicated app section
- Fixed query loader arg syntax
- Updated dependency: statsviz to pyroscope
- Disabled car_generator by default

### v1.12.0

- Added comprehensive `.golangci.yml` with 20+ linters including reliability, concurrency, style, and modernization categories
- Added new linters: noctx, exhaustive, durationcheck, errorlint, contextcheck, revive, funlen, gocognit, nestif, wsl_v5, perfsprint, intrange, modernize
- Updated Makefile with simplified targets (removed fmt, vet, check; integrated into build)
- Added gofumpt and gci formatters in install-tools
- Updated app version to v1.12.0

### v1.11.0

- Added circuit breaker configuration in http client package (`internal/infra/http`)
- Integrated failsafe-go for external API protection with retry policy
- Circuit breaker with 503 status code handling and exponential backoff
- Per-job circuit breaker configuration in scheduler (user_generator, car_generator)
- State change logging for circuit breaker monitoring

### v1.10.0

- Fixed go-critic issues in util package
- Updated Makefile fmt target with gofumpt
- Added circuit breaker with failsafe-go for external API protection

### v1.9.0

- Migration: Replace sqlx with pgx (jackc/pgx/v5) for PostgreSQL driver
- Updated database config to use pgxpool instead of sqlx.DB
- Updated repositories to use pgx Tx, Query, QueryRow APIs
- Fixed scan field mismatches to match SQL column count
- Fixed RowsAffected() return type (no error in pgx)

### v1.8.1

- Refactor: Move queryLoader.Compile calls from car_.go to car_sql.go
- Code quality: Define constant for duplicate cache key literal in car repository

### v1.8.0

- Added dedicated validator package (`internal/infra/validator`) with configurable validation rules
- Implemented SQL query cleaning utility for safe logging
- Added dynamic SQL query loader with native Go templates
- Added API benchmarking with Apache Bench (`make benchmark`)
- Added UserV2 handler, repository, and service for enhanced user operations
- Improved security by masking sensitive values in debug logs
- Added error codes and error messages for structured error handling

### v1.7.0

- Added generic query decoder (`util.DecodeQuery`) for reusable filter handling
- Added sql_builder utility for dynamic SQL query building with filtering/pagination
- Added many-to-many relationships (users and cars via junction table)
- Improved IDOR protection on user-scoped endpoints
