# Go, Supabase met SQLC Boilerplate

Productie-klare Go boilerplate met Vertical Slice Architecture en Supabase integratie

[![English](https://img.shields.io/badge/lang-English-blue.svg)](README.md)
[![한국어](https://img.shields.io/badge/lang-한국어-red.svg)](README.ko.md)
[![Français](https://img.shields.io/badge/lang-Français-yellow.svg)](README.fr.md)
[![Nederlands](https://img.shields.io/badge/lang-Nederlands-orange.svg)](README.nl.md)

## Belangrijkste Kenmerken

- **Microservices Architectuur**: Onafhankelijke services met duidelijke scheiding van verantwoordelijkheden
- **Vertical Slice Architecture**: Feature-complete structuur met hoge cohesie en lage koppeling
- **Supabase Integratie**: Vereenvoudigd PostgreSQL database beheer en migraties via Supabase
- **Moderne Stack**: Go 1.25, Chi v5, PostgreSQL (Supabase), Redis
- **Real-time Communicatie**: WebSocket ondersteuning
- **Event-Driven**: Redis Streams-gebaseerde event verwerking
- **Type Veiligheid**: Type-safe SQL queries via SQLC
- **Graceful Shutdown**: Correcte resource opruiming en verbindingsbeheer

## Projectstructuur

```
.
├── servers/                    # Go microservices
│   ├── cmd/                    # Service entry points
│   │   ├── api/                # REST API service (poort 8080)
│   │   ├── ws/                 # WebSocket service (poort 8081)
│   │   ├── stats/              # Stats service (poort 8084)
│   │   └── logging/            # Logging service (poort 8082)
│   ├── internal/
│   │   ├── feature/            # Business features (Vertical Slice)
│   │   ├── shared/             # Gedeelde infrastructuur
│   │   ├── stats/              # Stats verwerking
│   │   ├── logging/            # Logging service
│   │   └── ws_example/         # WebSocket handlers
│   └── test/                   # Integratie tests
├── supabase/                   # Supabase database beheer
│   ├── schemas/                # Database schema definities
│   ├── queries/                # SQLC query bestanden
│   ├── migrations/             # Database migraties (Supabase CLI)
│   └── config.toml             # Supabase project configuratie
└── script/                     # Code generatie en database beheer scripts
    ├── gen-sqlc.bash           # SQLC code generatie
    ├── gen-proto.bash          # Protocol Buffer code generatie
    ├── gen-typing-sb.bash      # TypeScript type generatie
    ├── reset-local-sb.bash     # Supabase lokale DB reset
    └── reset-remote-sb.bash    # Supabase remote DB reset
```

## Tech Stack

### Core

- **Go 1.25**: Generics ondersteuning
- **Chi v5**: Lichtgewicht HTTP router
- **gorilla/websocket**: WebSocket implementatie

### Data Laag

- **Supabase**: PostgreSQL hosting en database beheer platform
- **PostgreSQL**: Hoofd database (gehost op Supabase)
- **SQLC**: Type-safe SQL code generatie voor Go en TypeScript
  - Genereert type-safe Go code uit SQL queries
  - Genereert TypeScript code voor Supabase Edge Functions
  - **Let op**: TypeScript generatie ondersteunt geen `:exec`, `:execrows`, `:execresult`, `:batchexec` annotaties (gebruik `:one` of `:many` in plaats daarvan)
- **pgx/v5**: High-performance PostgreSQL driver
- **Supabase CLI**: Lokale ontwikkelomgeving en migratie beheer

### Caching & Messaging

- **Redis**: In-memory data store
- **Redis Streams**: Event streaming

## Snel Starten

### Vereisten

- Go 1.25+
- Supabase CLI ([Installatiegids](https://supabase.com/docs/guides/cli))
- Redis 7+
- Docker (voor het draaien van lokale Supabase)

### Installatie

```bash
# 1. Repository klonen
git clone https://github.com/your-org/go-monorepo-boilerplate.git
cd go-monorepo-boilerplate

# 2. Supabase lokale omgeving starten
supabase start
# PostgreSQL verbindingsinformatie wordt weergegeven

# 3. Omgevingsvariabelen configureren
cd servers
cp .env.example .env
# Bewerk .env met Supabase verbindingsinformatie

# 4. Dependencies installeren
go mod download

# 5. Type-safe code genereren uit SQL queries
cd ..
./script/gen-sqlc.bash
# Dit genereert:
# - Type-safe Go code voor backend services (servers/internal/sql/)
# - TypeScript types voor Supabase Edge Functions (supabase/functions/_shared/queries/)

# 6. (Optioneel) Database resetten indien nodig
./script/reset-local-sb.bash
```

### Services Uitvoeren

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

## Ontwikkeling

### Build

```bash
cd servers
go build ./...                    # Alle packages bouwen
go build ./cmd/api                # Specifieke service bouwen
```

### Testen

```bash
cd servers
go test ./...                     # Alle tests uitvoeren
go test -cover ./...              # Uitvoeren met coverage
go test -v ./internal/feature/... # Specifieke package tests uitvoeren
```

### Code Generatie

```bash
# Uitvoeren vanaf repository root
./script/gen-sqlc.bash           # Type-safe Go en TypeScript code genereren uit SQL
                                 # - Go: servers/internal/sql/ (volledige ondersteuning van alle SQLC annotaties)
                                 # - TypeScript: supabase/functions/_shared/queries/
                                 #   (beperkingen: :exec, :execrows, :execresult, :batchexec niet ondersteund)
./script/gen-proto.bash          # Protocol Buffer code genereren
./script/gen-typing-sb.bash      # TypeScript database schema types genereren
```

**BELANGRIJK**: Bij het schrijven van SQL queries voor TypeScript generatie, gebruik `:one` of `:many` annotaties in plaats van `:exec` familie annotaties. Voor queries die geen data retourneren, gebruik `:one` met een `RETURNING` clausule of selecteer een dummy waarde.

### Database Beheer (Supabase)

```bash
# Supabase lokale omgeving beheer
supabase start                   # Lokale Supabase starten
supabase stop                    # Lokale Supabase stoppen
supabase status                  # Supabase status controleren

# Migraties
supabase db reset                # Lokale DB resetten (alle migraties opnieuw uitvoeren)
supabase migration new <name>    # Nieuwe migratie aanmaken
supabase db push                 # Migraties toepassen op remote DB

# DB reset via scripts
./script/reset-local-sb.bash     # Lokale Supabase DB resetten en initiële data aanmaken
./script/reset-remote-sb.bash    # Remote Supabase DB resetten (gebruik met voorzichtigheid!)
```

### Supabase Integratie Workflow

Dit project maakt gebruik van Supabase als database beheer platform:

1. **Lokale Ontwikkeling**: Docker-gebaseerde PostgreSQL omgeving draaien met `supabase start`
2. **Schema Beheer**: Tabellen definiëren in `supabase/schemas/`, migraties opslaan in `supabase/migrations/`
3. **Type-Safe Queries**: Go code genereren uit SQL in `supabase/queries/` met SQLC
4. **Deployment**: Migraties toepassen op remote projecten met Supabase CLI

**Belangrijkste Voordelen**:

- Snelle opzet van lokale ontwikkelomgeving (Docker-gebaseerd)
- Geautomatiseerd migratie versiebeheer
- Visueel database beheer met Supabase Studio
- Vereenvoudigde productie deployment
- **Type-safe code generatie**: Schrijf SQL één keer, genereer automatisch type-safe Go en TypeScript code via `./script/gen-sqlc.bash`

## Architectuur Patronen

### Vertical Slice Architecture (Primair Patroon)

De kern architectuur van dit project is **Vertical Slice Architecture**. Elke feature is een complete verticale slice met alle lagen (HTTP → Business Logic → Data Access).

**Kenmerken**:

- Hoge cohesie per feature (alle benodigde code voor een feature op één plek)
- Lage koppeling (minimale afhankelijkheden tussen features)
- Snelle ontwikkeling en onderhoud (onafhankelijk werk per feature)

**Structuur Voorbeeld** (`internal/feature/user_profile/`):

```
internal/feature/user_profile/
  ├── router.go              # Route mapping (MapRoutes functie)
  ├── get_profile/
  │   ├── endpoint.go        # HTTP handler (Map functie)
  │   └── dto.go            # Request/response DTOs
  └── update_profile/
      ├── endpoint.go        # HTTP handler (Map functie)
      └── dto.go            # Request/response DTOs
```

**Endpoint Patroon**:

De `Map` functie van elk endpoint behandelt direct:

1. Logger en DB verbinding extraheren uit context
2. Request body parsen met `httputil.GetReqBodyWithLog`
3. Business logica uitvoeren (queries, validatie, etc.)
4. Response retourneren met `httputil.OkWithMsg` of `httputil.ErrWithMsg`

### Ondersteunende Patronen

**Component-Gebaseerde Structuur** (WebSocket, Stats, Logging services):

- Gestructureerd op technische concerns (sessies, packet handling, event consumption, etc.)
- Directe implementatie zonder laag scheiding

**Event-Driven Architectuur**:

- Asynchrone verwerking gebaseerd op Redis Streams
- Consumer-Processor patroon

**Repository Patroon** (`internal/repository/`):

- Template voor data access abstractie
- CRUD interface voorbeelden

### Belangrijkste Gedeelde Componenten

- **Redis Streams Consumer**: Generic-gebaseerde event consumer
- **Database Access**: SQLC-gegenereerde queries of directe pgx queries
- **HTTP Utilities**: Gestandaardiseerde request/response afhandeling
- **Graceful Shutdown**: Gebaseerd op `shared.Closer` interface

## API Endpoints

### API Service (Poort 8080)

- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /api/v1/ping` - Ping
- `POST /api/v1/user-profile/get` - Gebruikersprofiel ophalen
- `POST /api/v1/user-profile/update` - Gebruikersprofiel bijwerken

### WebSocket Service (Poort 8081)

- `GET /health` - Health check
- `GET /ws` - WebSocket verbinding

### Stats Service (Poort 8084)

- `GET /health` - Health check
- `GET /metrics` - Metrics ophalen

## Licentie

Apache License 2.0 - Zie [LICENSE](LICENSE) bestand voor details

## Bijdragen

Pull requests zijn welkom!

## Ondersteuning

Als u problemen ondervindt, maak dan een GitHub issue aan.
