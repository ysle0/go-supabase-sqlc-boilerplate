# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go microservices monorepo with a clean architecture pattern. The project is organized as a monorepo with the main Go code in `servers/` directory, PostgreSQL schemas in `supabase/`, and code generation scripts in `script/`.

## Build & Development Commands

### Building
```bash
# From servers/ directory
cd servers
go build ./...                    # Build all packages
go build ./cmd/api                # Build specific service
go build ./cmd/ws                 # Build WebSocket service
go build ./cmd/stats              # Build stats service
go build ./cmd/logging            # Build logging service
```

### Running Services
```bash
# From servers/ directory
go run ./cmd/api                  # API service (port 8080)
go run ./cmd/ws                   # WebSocket service (port 8081)
go run ./cmd/stats                # Stats service (port 8084)
go run ./cmd/logging              # Logging service (port 8082)
```

### Testing
```bash
# From servers/ directory
go test ./...                     # Run all tests
go test -cover ./...              # Run with coverage
go test -v ./internal/feature/... # Run specific package tests
go test -run TestName             # Run single test
```

### Code Generation
```bash
# From repository root
./script/gen-sqlc.bash           # Generate SQLC type-safe queries
./script/gen-proto.bash          # Generate Protocol Buffer code
./script/gen-typing-sb.bash      # Generate TypeScript types for Supabase
```

### Database Management
```bash
# From repository root
./script/reset-local-sb.bash     # Reset local Supabase database
./script/reset-remote-sb.bash    # Reset remote database (use with caution!)
```

## Architecture

### Monorepo Structure
- `servers/` - All Go microservices and shared code
  - `cmd/` - Service entry points (main packages)
  - `internal/` - Internal packages (not importable outside servers/)
  - `test/` - Integration tests and test helpers
- `supabase/` - Database schemas, queries, and migrations
- `script/` - Build and code generation scripts

### Service Organization

Each microservice in `cmd/` follows a similar pattern:
- `main.go` - Entry point, dependency setup, graceful shutdown
- `server.go` - Server struct, Start/Shutdown methods, implements `shared.Closer` interface

All servers must implement the `shared.Closer` interface for graceful shutdown:
```go
type Closer interface {
    Close(ctx context.Context) error
}
```

### Architecture Patterns

**Vertical Slice Architecture** (in `internal/feature/`):
Features are organized by use case, not by layer. Each feature is self-contained:
```
internal/feature/user_profile/
  ├── router.go              # Route mapping
  ├── get_profile/
  │   ├── endpoint.go        # HTTP handler
  │   └── dto.go            # Request/response DTOs
  └── update_profile/
      ├── endpoint.go
      └── dto.go
```

Each endpoint file (`endpoint.go`) contains a `Map` function that handles the HTTP request directly. The pattern is:
1. Extract logger and database connection from context
2. Parse request body using `httputil.GetReqBodyWithLog`
3. Execute business logic (queries, validation)
4. Return response using `httputil.OkWithMsg` or `httputil.ErrWithMsg`

**Traditional Three-Layer Architecture** (in other `internal/` packages):
```
Handler (HTTP) → Service (Business Logic) → Repository (Data Access)
```

Example in `internal/ws_example/`:
- `handler.go` - HTTP/WebSocket handlers
- `service.go` - Business logic interface and implementation
- Repository layer uses SQLC-generated code

### Key Shared Components

**Redis Streams Consumer** (`internal/shared/redisstream/consumer.go`):
- Generic consumer implementation using Go generics: `Consumer[T any]`
- Requires:
  - `Config` struct with stream settings
  - `ParseFunc[T]` to convert Redis messages to your type
  - Transfer channel to send parsed messages
- Usage pattern (see `internal/stats/consumer/event_consumer.go`):
  ```go
  consumer := redisstream.NewConsumer(logger, redisClient, eventCh, parseFunc, config)
  consumer.ConsumeLoop(ctx, wg)  // Note: uses sync.WaitGroup, not errgroup
  ```

**Database Access**:
- Uses SQLC for type-safe SQL queries (generated code in `internal/shared/database/sqlc/postgres/`)
- PostgreSQL connection pooling via `pgxpool`
- Two database access patterns:
  1. SQLC generated queries (type-safe, preferred)
  2. Direct pgx queries (for complex cases not covered by SQLC)

**HTTP Utilities** (`internal/shared/httputil/`):
- `GetReqBodyWithLog` - Parse and validate request bodies
- `OkWithMsg` / `ErrWithMsg` - Standardized response formatting
- `NewHttpUtilContext` - Create context wrapper for responses

**Middleware** (`internal/shared/middleware/`):
- Chi v5 router based
- Key middlewares: logging, recovery, CORS, API versioning

**Graceful Shutdown** (`internal/shared/gracefulExit.go`):
- `WaitForGracefulExit(ctx, timeout, closer)` - Handles SIGINT/SIGTERM
- All servers must implement `shared.Closer` interface

### WebSocket Implementation

WebSocket service (`cmd/ws/`) uses gorilla/websocket with:
- Session management via `sync.Map`
- Packet-based message handling
- Ping/Pong for connection health
- Graceful session cleanup on shutdown

Pattern:
```go
session.New(conn, logger) -> session.ReadLoop() / session.WriteLoop()
```

### Redis Integration

**Multiple Redis databases** (see `internal/shared/inmem/redis.go`):
- `RankingKey` (DB 0) - Ranking data
- `LoggingKey` (DB 1) - Logging streams
- `CacheKey` (DB 2) - General caching
- `QuestionKey` (DB 3) - Question data

Access pattern:
```go
redisClient := inmem.GetClient(ctx, inmem.CacheKey)
```

**Redis Streams** for event-driven architecture:
- Stats service consumes events from `stats:events` stream
- Logging service produces to `logging:messages` stream (currently placeholder)

### Testing Infrastructure

**Test Fixtures** (`test/helpers/testdb.go`):
- `TestFixtures` struct provides helper methods for setting up test data
- Uses raw SQL and SQLC queries
- Testcontainers for PostgreSQL and Redis in integration tests

Key test helper patterns:
```go
fixtures := helpers.NewTestFixtures(pool)
user, _ := fixtures.CreateTestUser(ctx, publicID)
fixtures.CreateTestStage(ctx, userID, categoryID, stageNum, coinRewards, crownRewards, isClaimed)
```

## Important Implementation Details

### Error Handling
- Use `shared.WrapError` for error context
- Log errors with structured logging (slog)
- Return appropriate HTTP status codes via httputil

### Environment Variables
All services use `shared.EnvString`, `shared.EnvInt`, `shared.EnvDuration` for configuration with defaults.

### Code Generation Dependencies
- SQLC requires `supabase/queries/*.sql` and `supabase/sqlc.yaml`
- Changes to database schema require running `./script/gen-sqlc.bash`

### Module Path
All imports use: `github.com/your-org/go-monorepo-boilerplate/servers/internal/...`

## Common Issues

1. **Import paths**: Always use full module path starting with `github.com/your-org/go-monorepo-boilerplate/servers`
2. **sync.WaitGroup vs errgroup**: Redis stream consumer uses `sync.WaitGroup`, not `errgroup.Group`
3. **Closer interface**: All server Start methods must return `shared.Closer`, not `func() error`
4. **Database connections**: Always release pooled connections with `defer dbconn.Release()`
5. **Context propagation**: Use `r.Context()` for HTTP handlers, pass context through all layers

## Dependencies

Key external dependencies:
- Chi v5 - HTTP router
- pgx/v5 - PostgreSQL driver
- go-redis/v9 - Redis client
- gorilla/websocket - WebSocket implementation
- testcontainers - Integration testing
- slogcolor - Structured logging with color output
