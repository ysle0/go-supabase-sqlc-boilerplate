# Go Supabase with SQLC Boilerplate

Production-ready Go boilerplate with Vertical Slice Architecture and Supabase integration

[![English](https://img.shields.io/badge/lang-English-blue.svg)](README.md)
[![한국어](https://img.shields.io/badge/lang-한국어-red.svg)](README.ko.md)
[![Français](https://img.shields.io/badge/lang-Français-yellow.svg)](README.fr.md)
[![Nederlands](https://img.shields.io/badge/lang-Nederlands-orange.svg)](README.nl.md)

## Key Features

- **Microservices Architecture**: Independent services with clear separation of concerns
- **Vertical Slice Architecture**: Feature-complete structure with high cohesion and low coupling
- **Supabase Integration**: Simplified PostgreSQL database management and migrations via Supabase
- **Modern Stack**: Go 1.25, Chi v5, PostgreSQL (Supabase), Redis
- **Real-time Communication**: WebSocket support
- **Event-Driven**: Redis Streams-based event processing
- **Type Safety**: Type-safe SQL queries through SQLC
- **Graceful Shutdown**: Proper resource cleanup and connection handling

## Project Structure

```
.
├── servers/                    # Go microservices
│   ├── cmd/                    # Service entry points
│   │   ├── api/                # REST API service (port 8080)
│   │   ├── ws/                 # WebSocket service (port 8081)
│   │   ├── stats/              # Stats service (port 8084)
│   │   └── logging/            # Logging service (port 8082)
│   ├── internal/
│   │   ├── feature/            # Business features (Vertical Slice)
│   │   ├── shared/             # Shared infrastructure
│   │   ├── stats/              # Stats processing
│   │   ├── logging/            # Logging service
│   │   └── ws_example/         # WebSocket handlers
│   └── test/                   # Integration tests
├── supabase/                   # Supabase database management
│   ├── schemas/                # Database schema definitions
│   ├── queries/                # SQLC query files
│   ├── migrations/             # Database migrations (Supabase CLI)
│   └── config.toml             # Supabase project configuration
└── script/                     # Code generation and database management scripts
    ├── gen-sqlc.bash           # SQLC code generation
    ├── gen-proto.bash          # Protocol Buffer code generation
    ├── gen-typing-sb.bash      # TypeScript type generation
    ├── reset-local-sb.bash     # Supabase local DB reset
    └── reset-remote-sb.bash    # Supabase remote DB reset
```

## Tech Stack

### Core
- **Go 1.25**: Generics support
- **Chi v5**: Lightweight HTTP router
- **gorilla/websocket**: WebSocket implementation

### Data Layer
- **Supabase**: PostgreSQL hosting and database management platform
- **PostgreSQL**: Main database (hosted on Supabase)
- **SQLC**: Type-safe SQL code generation for Go and TypeScript
  - Generates type-safe Go code from SQL queries
  - Generates TypeScript code for Supabase Edge Functions
  - **Note**: TypeScript generation doesn't support `:exec`, `:execrows`, `:execresult`, `:batchexec` annotations (use `:one` or `:many` instead)
- **pgx/v5**: High-performance PostgreSQL driver
- **Supabase CLI**: Local development environment and migration management

### Caching & Messaging
- **Redis**: In-memory data store
- **Redis Streams**: Event streaming

## Quick Start

### Prerequisites

- Go 1.25+
- Supabase CLI ([Installation Guide](https://supabase.com/docs/guides/cli))
- Redis 7+
- Docker (for running local Supabase)

### Installation

```bash
# 1. Clone repository
git clone https://github.com/your-org/go-monorepo-boilerplate.git
cd go-monorepo-boilerplate

# 2. Start Supabase local environment
supabase start
# PostgreSQL connection info will be displayed

# 3. Configure environment variables
cd servers
cp .env.example .env
# Edit .env with Supabase connection info

# 4. Install dependencies
go mod download

# 5. Generate type-safe code from SQL queries
cd ..
./script/gen-sqlc.bash
# This generates:
# - Type-safe Go code for backend services (servers/internal/sql/)
# - TypeScript types for Supabase Edge Functions (supabase/functions/_shared/queries/)

# 6. (Optional) Reset database if needed
./script/reset-local-sb.bash
```

### Running Services

```bash
cd servers

# API service
go run ./cmd/api

# WebSocket service
go run ./cmd/ws

# Stats service
go run ./cmd/stats

# Logging service
go run ./cmd/logging
```

## Development

### Build

```bash
cd servers
go build ./...                    # Build all packages
go build ./cmd/api                # Build specific service
```

### Testing

```bash
cd servers
go test ./...                     # Run all tests
go test -cover ./...              # Run with coverage
go test -v ./internal/feature/... # Run specific package tests
```

### Code Generation

```bash
# Run from repository root
./script/gen-sqlc.bash           # Generate type-safe Go and TypeScript code from SQL
                                 # - Go: servers/internal/sql/ (fully supports all SQLC annotations)
                                 # - TypeScript: supabase/functions/_shared/queries/
                                 #   (limitations: :exec, :execrows, :execresult, :batchexec not supported)
./script/gen-proto.bash          # Generate Protocol Buffer code
./script/gen-typing-sb.bash      # Generate TypeScript database schema types
```

**IMPORTANT**: When writing SQL queries for TypeScript generation, use `:one` or `:many` annotations instead of `:exec` family annotations. For queries that don't return data, use `:one` with a `RETURNING` clause or select a dummy value.

### Database Management (Supabase)

```bash
# Supabase local environment management
supabase start                   # Start local Supabase
supabase stop                    # Stop local Supabase
supabase status                  # Check Supabase status

# Migrations
supabase db reset                # Reset local DB (re-run all migrations)
supabase migration new <name>    # Create new migration
supabase db push                 # Apply migrations to remote DB

# DB reset via scripts
./script/reset-local-sb.bash     # Reset local Supabase DB and create initial data
./script/reset-remote-sb.bash    # Reset remote Supabase DB (use with caution!)
```

### Supabase Integration Workflow

This project leverages Supabase as the database management platform:

1. **Local Development**: Run Docker-based PostgreSQL environment with `supabase start`
2. **Schema Management**: Define tables in `supabase/schemas/`, store migrations in `supabase/migrations/`
3. **Type-Safe Queries**: Generate Go code from SQL in `supabase/queries/` using SQLC
4. **Deployment**: Apply migrations to remote projects using Supabase CLI

**Key Benefits**:
- Quickly set up local development environment (Docker-based)
- Automated migration version control
- Visual database management with Supabase Studio
- Simplified production deployment
- **Type-safe code generation**: Write SQL once, generate type-safe Go and TypeScript code automatically via `./script/gen-sqlc.bash`

## Architecture Patterns

### Vertical Slice Architecture (Primary Pattern)

The core architecture of this project is **Vertical Slice Architecture**. Each feature is a complete vertical slice containing all layers (HTTP → Business Logic → Data Access).

**Characteristics**:

- High cohesion by feature (all code needed for a feature in one place)
- Low coupling (minimal dependencies between features)
- Fast development and maintenance (independent work by feature)

**Structure Example** (`internal/feature/user_profile/`):

```
internal/feature/user_profile/
  ├── router.go              # Route mapping (MapRoutes function)
  ├── get_profile/
  │   ├── endpoint.go        # HTTP handler (Map function)
  │   └── dto.go            # Request/response DTOs
  └── update_profile/
      ├── endpoint.go        # HTTP handler (Map function)
      └── dto.go            # Request/response DTOs
```

**Endpoint Pattern**:

Each endpoint's `Map` function directly handles:

1. Extract logger and DB connection from context
2. Parse request body using `httputil.GetReqBodyWithLog`
3. Execute business logic (queries, validation, etc.)
4. Return response using `httputil.OkWithMsg` or `httputil.ErrWithMsg`

### Supporting Patterns

**Component-Based Structure** (WebSocket, Stats, Logging services):

- Structured by technical concerns (sessions, packet handling, event consumption, etc.)
- Direct implementation without layer separation

**Event-Driven Architecture**:

- Asynchronous processing based on Redis Streams
- Consumer-Processor pattern

**Repository Pattern** (`internal/repository/`):

- Template for data access abstraction
- CRUD interface examples

### Key Shared Components

- **Redis Streams Consumer**: Generic-based event consumer
- **Database Access**: SQLC-generated queries or direct pgx queries
- **HTTP Utilities**: Standardized request/response handling
- **Graceful Shutdown**: Based on `shared.Closer` interface

## API Endpoints

### API Service (Port 8080)
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /api/v1/ping` - Ping
- `POST /api/v1/user-profile/get` - Get user profile
- `POST /api/v1/user-profile/update` - Update user profile

### WebSocket Service (Port 8081)
- `GET /health` - Health check
- `GET /ws` - WebSocket connection

### Stats Service (Port 8084)
- `GET /health` - Health check
- `GET /metrics` - Get metrics

## License

Apache License 2.0 - See [LICENSE](LICENSE) file for details

## Contributing

Pull requests are welcome!

## Support

If you have any issues, please file a GitHub issue.
