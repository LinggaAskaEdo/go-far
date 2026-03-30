# Go-Far - Clean Architecture CRUD API

A production-ready RESTful API built with Go following Clean Architecture principles, featuring PostgreSQL, Redis, JWT authentication, rate limiting, and OpenTelemetry tracing.

## 🚀 Features

- **Clean Architecture** - Separation of concerns with handlers, services, and repositories
- **REST API** - Built with Gin framework
- **Database** - PostgreSQL with sqlx (MySQL supported)
- **Caching** - Redis with snappy compression
- **Authentication** - JWT with RSA-256 signing and refresh tokens
- **Rate Limiting** - Dual-layer (route + global) using Lua scripts
- **Observability** - OpenTelemetry tracing with OTLP exporter
- **Scheduled Jobs** - Cron-based job scheduler
- **API Documentation** - Swagger/OpenAPI 2.0
- **Graceful Shutdown** - Proper cleanup of resources

## 📁 Project Structure

```text
go-far/
├── src/
│   ├── cmd/                    # Application entry point
│   │   ├── main.go             # Main application bootstrap
│   │   ├── app.go              # Dependency injection & initialization
│   │   └── conf.go             # Configuration loading
│   ├── config/                 # Configuration modules
│   │   ├── auth/               # JWT authentication
│   │   ├── database/           # Database connection
│   │   ├── grace/              # Graceful shutdown
│   │   ├── logger/             # Zerolog logger
│   │   ├── middleware/         # Request middleware (CORS, rate limiting)
│   │   ├── query/              # SQL query loader
│   │   ├── redis/              # Redis client
│   │   ├── scheduler/          # Cron scheduler
│   │   ├── server/             # HTTP server & Gin engine
│   │   └── tracer/             # OpenTelemetry tracer
│   ├── domain/                 # Business entities
│   │   ├── user.go             # User entity
│   │   └── car.go              # Car entity
│   ├── dto/                    # Data Transfer Objects
│   │   ├── request.go          # Request DTOs
│   │   ├── response.go         # Response DTOs
│   │   └── pagination.go       # Pagination support
│   ├── errors/                 # Error handling
│   ├── handler/                # HTTP & scheduler handlers
│   │   ├── rest/               # REST API handlers
│   │   │   ├── user.go         # User handlers
│   │   │   └── car.go          # Car handlers
│   │   └── scheduler/          # Cron job handlers
│   ├── preference/             # Constants
│   ├── repository/             # Data access layer
│   │   ├── user/               # User repository
│   │   └── car/                # Car repository
│   ├── service/                # Business logic layer
│   │   ├── user/               # User service
│   │   └── car/                # Car service
│   └── util/                   # Utility functions
├── etc/
│   ├── cert/                   # RSA keys for JWT
│   ├── migrations/             # Database migrations
│   └── queries/                # SQL queries with templates
├── docs/                       # Swagger documentation
├── logs/                       # Application logs
├── config.yaml                 # Application configuration
├── Makefile                    # Build & run commands
└── go.mod                      # Go module definition
```

## 🛠️ Tech Stack

| Component  | Technology                   |
| ---------- | ---------------------------- |
| Framework  | Gin                          |
| Database   | PostgreSQL (MySQL supported) |
| ORM        | sqlx                         |
| Cache      | Redis                        |
| Auth       | JWT (RSA-256)                |
| Logging    | Zerolog                      |
| Tracing    | OpenTelemetry                |
| Scheduler  | robfig/cron/v3               |
| Validation | go-playground/validator      |
| Docs       | Swagger (swaggo)             |

## 📋 API Endpoints

### Health Check

| Method | Endpoint    | Description       |
|--------|-------------|-------------------|
| GET    | `/health`   | Health check      |
| GET    | `/ready`    | Readiness check   |

### Users

| Method | Endpoint     | Description            |
|--------|--------------|------------------------|
| POST   | `/users`     | Create user            |
| GET    | `/users/:id` | Get user by ID         |
| GET    | `/users`     | List users (paginated) |
| PUT    | `/users/:id` | Update user            |
| DELETE | `/users/:id` | Delete user            |

### Cars

| Method | Endpoint                       | Description                   |
|--------|--------------------------------|-------------------------------|
| POST   | `/cars`                        | Create car                    |
| POST   | `/cars/bulk`                   | Create multiple cars          |
| GET    | `/cars/:id`                    | Get car by ID                 |
| GET    | `/cars/:id/owner`              | Get car with owner details    |
| PUT    | `/cars/:id`                    | Update car                    |
| DELETE | `/cars/:id`                    | Delete car                    |
| POST   | `/cars/:id/transfer`           | Transfer ownership            |
| PUT    | `/cars/availability`           | Bulk update availability      |
| GET    | `/cars/by-user/:user_id`       | List cars by user             |
| GET    | `/cars/by-user/:user_id/count` | Count cars by user            |

### Swagger Documentation

Access Swagger UI at: `http://localhost:8181/swagger/index.html`

## ⚙️ Configuration

Edit `config.yaml` or use environment variables:

```yaml
server:
  port: 8181
  write_timeout: 10s
  read_timeout: 10s

postgres:
  host: localhost
  port: 5432
  user: postgres
  password: your_password
  dbname: go_far

redis:
  address: localhost:6379
  password: ""

auth:
  private_key: ./etc/cert/id_rsa
  public_key: ./etc/cert/id_rsa.pub
  expired_token: 5m
  expired_refresh_token: 15m
```

### Environment Variables

```bash
# Server
export SERVER_PORT=8181

# Database
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_NAME=go_far

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

- Go 1.24+
- PostgreSQL 14+
- Redis 7+
- Make (optional)

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/yourusername/go-far.git
   cd go-far
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Generate RSA keys for JWT**

   ```bash
   mkdir -p etc/cert
   openssl genrsa -out etc/cert/id_rsa 2048
   openssl rsa -in etc/cert/id_rsa -pubout -out etc/cert/id_rsa.pub
   ```

4. **Setup database**

   ```bash
   createdb go_far
   # Run migrations if any
   ```

5. **Update configuration**

   ```bash
   cp config.yaml config.yaml.local
   # Edit config.yaml.local with your settings
   ```

6. **Run the application**

   ```bash
   make run
   # Or manually
   go run src/cmd/*.go
   ```

### Using Make

```bash
make run          # Run the application
make build        # Build the binary
make test         # Run tests
make lint         # Run linters
make clean        # Clean build artifacts
```

## 📝 Example Requests

### Create User

```bash
curl -X POST http://localhost:8181/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
  }'
```

### Create Car

```bash
curl -X POST http://localhost:8181/cars \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your_token>" \
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

## 🔒 Authentication

The API uses JWT with RSA-256 signing. Include the token in the Authorization header:

```text
Authorization: Bearer <your_jwt_token>
```

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
