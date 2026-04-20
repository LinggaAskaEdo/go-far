# Go-Far - DDD CRUD API

A production-ready RESTful API built with Go following Domain-Driven Design principles, featuring PostgreSQL, Redis, JWT authentication, role-based rate limiting, and OpenTelemetry tracing.

## 🚀 Features

- **Domain-Driven Design** - Separation of concerns with handlers, services, and repositories
- **REST API** - Built with Go's native `net/http` (Go 1.22+ pattern matching)
- **Database** - PostgreSQL with sqlx (MySQL supported)
- **Caching** - Redis with snappy compression
- **Authentication** - JWT with HS-256 signing, refresh tokens, and role-based access
- **Role-Based Rate Limiting** - Per-role sliding window rate limiting via Redis Lua script (atomic, race-condition free)
- **Observability** - OpenTelemetry tracing with OTLP exporter
- **Scheduled Jobs** - Cron-based job scheduler
- **API Documentation** - Swagger/OpenAPI 2.0
- **Graceful Shutdown** - Proper cleanup of resources
  - **Many-to-Many Relationships** - Users and cars via junction table
  - **Generic Query Decoder** - Reflection-based HTTP query decoder for reusable filter handling
- **SQL Query Cleaner** - Utility to clean SQL queries for logging (masks sensitive values like passwords)
- **Query Loader** - Dynamic SQL query loading from files with tqla template support
- **Custom Validators** - Separate validator configuration package for reusable validation logic
- **API Benchmarking** - Built-in Apache Bench integration for performance testing

## 📁 Project Structure

```text
go-far/
├── src/
│   ├── cmd/                    # Application entry point
│   │   ├── main.go             # Main application bootstrap
│   │   ├── app.go              # Dependency injection & initialization
│   │   └── conf.go             # Configuration loading
│   ├── config/                 # Infrastructure & configuration modules
│   │   ├── database/           # Database connection
│   │   ├── grace/              # Graceful shutdown
│   │   ├── logger/             # Zerolog logger
│   │   ├── middleware/         # Request middleware (CORS, rate limiting)
│   │   ├── query/              # SQL query loader (tqla templates)
│   │   ├── redis/              # Redis client
│   │   ├── scheduler/          # Cron scheduler
│   │   ├── server/             # HTTP server & native net/http router
│   │   ├── token/              # JWT token management (HS-256)
│   │   └── tracer/             # OpenTelemetry tracer
│   ├── handler/                # HTTP & scheduler handlers (interface layer)
│   │   ├── rest/               # REST API handlers
│   │   │   ├── auth.go         # Auth endpoints (register, login, refresh)
│   │   │   ├── user.go         # User handlers
│   │   │   └── car.go          # Car handlers
│   │   └── scheduler/          # Cron job handlers
│   ├── model/                  # Shared data contracts
│   │   ├── entity/             # Domain entities
│   │   │   ├── user.go         # User entity with roles
│   │   │   └── car.go          # Car entity
│   │   ├── dto/                # Data Transfer Objects
│   │   │   ├── request.go      # Request DTOs
│   │   │   ├── response.go     # Response DTOs
│   │   │   └── pagination.go   # Pagination support
│   │   └── errors/             # Error handling
│   ├── preference/             # Constants & shared values
│   ├── repository/             # Data access layer (infrastructure)
│   │   ├── user/               # User repository
│   │   └── car/                # Car repository
│   ├── service/                # Business logic layer (domain services)
│   │   ├── user/               # User service
│   │   └── car/                # Car service
│   └── util/                   # Utility functions
├── etc/
│   ├── docs/                   # Swagger documentation (generated)
│   ├── migrations/             # Database migrations
│   ├── queries/                # SQL queries with tqla templates
│   └── postman/                # Postman collection
├── logs/                       # Application logs
├── config.yaml                 # Application configuration
├── Makefile                    # Build & run commands
└── go.mod                      # Go module definition
```

## 🛠️ Tech Stack

| Component       | Technology                   |
| --------------- | ---------------------------- |
| Framework       | Go 1.22+ `net/http`          |
| Database        | PostgreSQL (MySQL supported) |
| DB Driver       | pgx (jackc/pgx)              |
| Cache           | Redis                        |
| Auth            | JWT (HS-256) + bcrypt        |
| Logging         | Zerolog                      |
| Tracing         | OpenTelemetry                |
| Scheduler       | robfig/cron/v3               |
| Validation      | go-playground/validator      |
| Docs            | Swagger (swaggo)             |
| SQL Templating  | tqla                         |
| Query Decoder   | Custom reflection-based      |
| Benchmark       | Apache Bench (ab)            |

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

Edit `config.yaml` or use environment variables:

```yaml
server:
  port: 8181
  write_timeout: 10s
  read_timeout: 10s
  max_body_bytes: 1048576  # 1MB, typical: 1-10MB for REST APIs

postgres:
  host: localhost
  port: 5432
  user: postgres
  password: ${POSTGRES_DOCKER_PASSWORD}
  dbname: go_far

redis:
  address: localhost:6379
  password: ""

token:
  expired_token: 5m
  expired_refresh_token: 15m

scheduler:
  enabled: true
  jobs:
    user_generator:
      enabled: true
      cron: "0 0 */1 * * *"  # Every 1 hour
      batch_size: 5
      min_age: 18
      max_age: 80
    car_generator:
      enabled: true
      cron: "0 */30 * * * *"  # Every 30 minutes
      batch_size: 3
      min_year: 2015
      max_year: 2025

middleware:
  public_paths:
    - /health
    - /ready
    - /swagger/
    - /auth/login
    - /auth/register
    - /auth/refresh
  rate_limiter:
    command: "1-S"
    limit: 500
  role_rate_limit:
    admin:
      command: "1-M"
      limit: 10
    user:
      command: "1-M"
      limit: 5
    guest:
      command: "1-M"
      limit: 1
```

### Scheduler Jobs

The application includes automatic data seeding via cron jobs:

| Job              | Schedule        | Description                                      |
| ---------------- | --------------- | ------------------------------------------------ |
| `user_generator` | Every 1 hour    | Generates random users with realistic names      |
| `car_generator`  | Every 30 mins   | Generates random cars with real models/colors    |

**Car Generator** includes 70+ real car models from major brands (Toyota, Honda, Ford, BMW, Mercedes-Benz, Audi, Tesla, etc.) with authentic colors and randomly generated US-format license plates.

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

# JWT Secret (loaded from /etc/environment)
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

### /etc/environment

The app automatically loads `/etc/environment` at startup. Add your secrets there:

```bash
JWT_SECRET_GO_FAR=cc085ae29c8ada6eedba8e7f04fb669e...
POSTGRES_DOCKER_PASSWORD=a5k4CooL
```

## 🚦 Getting Started

### Prerequisites

- Go 1.25+
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
   # Add to /etc/environment (app loads automatically at startup)
   echo "JWT_SECRET_GO_FAR=$(openssl rand -hex 64)" | sudo tee -a /etc/environment
   ```

5. **Setup database**

   ```bash
   createdb go_far
   make sql-postgres-up
   ```

6. **Update configuration**

   ```bash
   cp config.yaml config.yaml.local
   # Edit config.yaml.local with your settings
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
make help            # Show all available commands
make all             # Clean, build, and run app
make deps            # Download and install dependencies
make update          # Update all dependencies to latest
make install-tools   # Install dev tools (swag, lint, etc)

make build           # Build application with optimizations
make run             # Run the built application
make clean           # Remove build artifacts and logs

make check           # Run all checks (fmt, vet, lint)
make fmt             # Format code with go fmt
make vet             # Run go vet for common issues
make lint            # Run golangci-lint
make test            # Run tests with coverage report

make swagger         # Generate Swagger API docs

make migrate         # Run database migrations (postgres)
make sql-postgres-create   # Create new postgres migration
make sql-postgres-up       # Apply postgres migrations
make sql-mysql-create      # Create new mysql migration
make sql-mysql-up          # Apply mysql migrations

make cert-install    # Install OpenSSL (auto-detects package manager)
make cert-create     # Generate RSA key pair (4096-bit)

make mon-start       # Start monitoring stack
make mon-stop        # Stop monitoring stack

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

The `util.DecodeQuery` function automatically decodes HTTP query parameters to DTOs using struct tags:

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
filter := util.DecodeQuery[dto.UserFilter](r.URL.Query())
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

Logs are written to `logs/app.log` with rotation. Format is configurable (JSON/Console).

### Tracing

OpenTelemetry traces are exported to `localhost:4317` (OTLP/gRPC). Configure your collector accordingly.

### Metrics

Metrics collection is available via OpenTelemetry (configuration required).

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

## 📋 Changelog

### v1.8.0

- Added dedicated validator package (`src/config/validator`) with configurable validation rules
- Implemented SQL query cleaning utility for safe logging
- Added dynamic SQL query loader with tqla template support
- Added API benchmarking with Apache Bench (`make benchmark`)
- Added UserV2 handler, repository, and service for enhanced user operations
- Improved security by masking sensitive values in debug logs

### v1.8.1

- Refactor: Move queryLoader.Compile calls from car_.go to car_sql.go
- Code quality: Define constant for duplicate cache key literal in car repository
- Migration: Replace sqlx with pgx (jackc/pgx/v5) for PostgreSQL driver

### v1.7.0

- Added generic query decoder for reusable filter handling
- Added many-to-many relationships (users and cars)
- Improved IDOR protection on user-scoped endpoints
