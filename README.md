# Go Microservices Boilerplate

A production-ready Go microservices boilerplate with clean architecture patterns, modern tooling, and comprehensive examples.

## 🚀 Features

- **Microservices Architecture**: Multiple independent services with clear separation of concerns
- **Clean Architecture**: Three-layer architecture (Handler → Service → Repository)
- **Modern Stack**: Go 1.25, Chi v5, PostgreSQL, Redis, Protocol Buffers
- **Real-time Communication**: WebSocket support with gorilla/websocket
- **Event-Driven**: Redis Streams consumer for event processing
- **Type Safety**: SQLC for type-safe SQL queries
- **Database**: PostgreSQL with Supabase support and Row Level Security
- **Containerization**: Multi-stage Docker builds with scratch base images
- **Graceful Shutdown**: Proper resource cleanup and connection handling
- **Observability**: Structured logging, pprof profiling support

## 📁 Project Structure

```
.
├── servers/
│   ├── cmd/                    # Service entry points
│   │   ├── api/                # REST API service
│   │   ├── ws/                 # WebSocket service
│   │   ├── stats/              # Stats service (Redis Streams)
│   │   └── logging/            # Logging service (gRPC)
│   ├── internal/
│   │   ├── feature/            # Business features
│   │   │   └── example_item/   # Example CRUD feature
│   │   ├── repository/         # Data access layer
│   │   ├── shared/             # Shared infrastructure
│   │   │   ├── database/       # DB connection pooling
│   │   │   ├── middleware/     # HTTP middlewares
│   │   │   ├── redisstream/    # Redis Streams consumer
│   │   │   └── ...
│   │   ├── stats/              # Stats processing
│   │   ├── logging/            # Logging service
│   │   └── ws_example/         # WebSocket handlers
│   ├── build/                  # Docker build configs
│   ├── test/                   # Integration tests
│   └── scripts/                # Build scripts
├── supabase/
│   ├── schemas/                # Database schemas
│   ├── queries/                # SQLC queries
│   └── migrations/             # Database migrations
├── script/                     # Code generation scripts
└── documentation/              # Additional documentation
```

## 🛠️ Tech Stack

### Core
- **Go 1.25**: Latest Go version with generics
- **Chi v5**: Lightweight HTTP router
- **gorilla/websocket**: WebSocket protocol implementation

### Data Layer
- **PostgreSQL**: Primary database
- **Supabase**: Managed PostgreSQL with built-in features
- **SQLC**: Type-safe SQL code generation
- **pgx/v5**: High-performance PostgreSQL driver

### Caching & Messaging
- **Redis**: In-memory data store
- **Redis Streams**: Event streaming and processing

### DevOps
- **Docker**: Multi-stage containerization
- **Task**: Task runner (Taskfile.yml)
- **Air**: Live reload for development

### Observability
- **slog**: Structured logging (Go standard library)
- **pprof**: Performance profiling

## 🚀 Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 14+
- Redis 7+
- Docker & Docker Compose (optional)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/your-org/go-monorepo-boilerplate.git
   cd go-monorepo-boilerplate
   ```

2. **Set up environment variables**
   ```bash
   cd servers
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Install dependencies**
   ```bash
   cd servers
   go mod download
   ```

4. **Set up database**
   ```bash
   # Apply migrations
   psql -U postgres -d your_db < supabase/migrations/20250101000000_initial_schema.sql
   ```

5. **Generate code (SQLC)**
   ```bash
   ./script/gen-sqlc.bash
   ```

### Running Services

#### Using Task CLI

```bash
# Run API service
task api:run

# Run WebSocket service
task ws:run

# Run Stats service
task stats:run

# Run Logging service
task logging:run
```

#### Manual Run

```bash
# API Service
cd servers
go run ./cmd/api

# WebSocket Service
go run ./cmd/ws

# Stats Service
go run ./cmd/stats

# Logging Service
go run ./cmd/logging
```

#### Using Docker

```bash
# Build all services
docker-compose build

# Run all services
docker-compose up

# Run specific service
docker-compose up api
```

## 📚 Services

### API Service (Port 8080)

REST API service with example CRUD operations.

**Endpoints:**
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /api/v1/ping` - Ping endpoint
- `GET /api/v1/items` - List items
- `POST /api/v1/items` - Create item
- `GET /api/v1/items/{id}` - Get item
- `PUT /api/v1/items/{id}` - Update item
- `DELETE /api/v1/items/{id}` - Delete item

### WebSocket Service (Port 8081)

Real-time WebSocket communication with packet handling.

**Features:**
- Session management
- Ping/Pong handlers
- Echo messages
- Graceful disconnection

### Stats Service (Port 8084)

Event-driven statistics processing using Redis Streams.

**Features:**
- Redis Streams consumer
- Event aggregation
- Metrics collection
- HTTP metrics endpoint

### Logging Service (Port 8082)

Centralized logging service using gRPC.

**Features:**
- Prioritized logging queues
- gRPC protocol
- Structured log storage

## 🧪 Testing

```bash
# Run all tests
cd servers
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests
task test:integration
```

## 📖 Documentation

- [Architecture](documentation/architecture.md) - System architecture and design patterns
- [Getting Started](documentation/getting-started.md) - Detailed setup guide
- [Database](documentation/database.md) - Database schema and queries
- [Deployment](documentation/deployment.md) - Deployment strategies
- [Patterns](documentation/patterns.md) - Code patterns and best practices

## 🔧 Development

### Code Generation

```bash
# Generate SQLC code
./script/gen-sqlc.bash

# Generate Protocol Buffers
./script/gen-proto.bash

# Generate TypeScript types for Supabase
./script/gen-typing-sb.bash
```

### Database Management

```bash
# Reset local database
./script/reset-local-sb.bash

# Reset remote database (use with caution!)
./script/reset-remote-sb.bash
```

## 🏗️ Architecture Patterns

### Three-Layer Architecture

```
Handler (HTTP) → Service (Business Logic) → Repository (Data Access)
```

### Dependency Injection

All components use constructor injection for better testability.

### Interface-Based Design

Core components define interfaces for easy mocking and testing.

## 📝 License

Apache License 2.0 - see [LICENSE](LICENSE) file for details

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📧 Support

For support, please open an issue in the GitHub repository.

---

**Generated with ❤️ using Go**
