# Boilerplate Go, Supabase avec SQLC

Boilerplate de Go prêt pour la production avec Vertical Slice Architecture et intégration Supabase

[![English](https://img.shields.io/badge/lang-English-blue.svg)](README.md)
[![한국어](https://img.shields.io/badge/lang-한국어-red.svg)](README.ko.md)
[![Français](https://img.shields.io/badge/lang-Français-yellow.svg)](README.fr.md)
[![Nederlands](https://img.shields.io/badge/lang-Nederlands-orange.svg)](README.nl.md)

## Fonctionnalités Principales

- **Architecture Microservices** : Services indépendants avec séparation claire des préoccupations
- **Vertical Slice Architecture** : Structure complète par fonctionnalité avec haute cohésion et faible couplage
- **Intégration Supabase** : Gestion simplifiée de la base de données PostgreSQL et migrations via Supabase
- **Stack Moderne** : Go 1.25, Chi v5, PostgreSQL (Supabase), Redis
- **Communication Temps Réel** : Support WebSocket
- **Event-Driven** : Traitement d'événements basé sur Redis Streams
- **Sécurité des Types** : Requêtes SQL type-safe via SQLC
- **Arrêt Gracieux** : Nettoyage approprié des ressources et gestion des connexions

## Structure du Projet

```
.
├── servers/                    # Microservices Go
│   ├── cmd/                    # Points d'entrée des services
│   │   ├── api/                # Service API REST (port 8080)
│   │   ├── ws/                 # Service WebSocket (port 8081)
│   │   ├── stats/              # Service statistiques (port 8084)
│   │   └── logging/            # Service logging (port 8082)
│   ├── internal/
│   │   ├── feature/            # Fonctionnalités métier (Vertical Slice)
│   │   ├── shared/             # Infrastructure partagée
│   │   ├── stats/              # Traitement des statistiques
│   │   ├── logging/            # Service de logging
│   │   └── ws_example/         # Gestionnaires WebSocket
│   └── test/                   # Tests d'intégration
├── supabase/                   # Gestion de la base de données Supabase
│   ├── schemas/                # Définitions de schéma de base de données
│   ├── queries/                # Fichiers de requêtes SQLC
│   ├── migrations/             # Migrations de base de données (Supabase CLI)
│   └── config.toml             # Configuration du projet Supabase
└── script/                     # Scripts de génération de code et gestion de base de données
    ├── gen-sqlc.bash           # Génération de code SQLC
    ├── gen-proto.bash          # Génération de code Protocol Buffer
    ├── gen-typing-sb.bash      # Génération de types TypeScript
    ├── reset-local-sb.bash     # Réinitialisation de la BD locale Supabase
    └── reset-remote-sb.bash    # Réinitialisation de la BD distante Supabase
```

## Stack Technique

### Core

- **Go 1.25** : Support des génériques
- **Chi v5** : Routeur HTTP léger
- **gorilla/websocket** : Implémentation WebSocket

### Couche de Données

- **Supabase** : Hébergement PostgreSQL et plateforme de gestion de base de données
- **PostgreSQL** : Base de données principale (hébergée sur Supabase)
- **SQLC** : Génération de code SQL type-safe pour Go et TypeScript
  - Génère du code Go type-safe à partir de requêtes SQL
  - Génère du code TypeScript pour les Supabase Edge Functions
  - **Note** : La génération TypeScript ne supporte pas les annotations `:exec`, `:execrows`, `:execresult`, `:batchexec` (utilisez `:one` ou `:many` à la place)
- **pgx/v5** : Driver PostgreSQL haute performance
- **Supabase CLI** : Environnement de développement local et gestion des migrations

### Cache & Messagerie

- **Redis** : Stockage de données en mémoire
- **Redis Streams** : Streaming d'événements

## Démarrage Rapide

### Prérequis

- Go 1.25+
- Supabase CLI ([Guide d'installation](https://supabase.com/docs/guides/cli))
- Redis 7+
- Docker (pour exécuter Supabase localement)

### Installation

```bash
# 1. Cloner le dépôt
git clone https://github.com/your-org/go-monorepo-boilerplate.git
cd go-monorepo-boilerplate

# 2. Démarrer l'environnement local Supabase
supabase start
# Les informations de connexion PostgreSQL seront affichées

# 3. Configurer les variables d'environnement
cd servers
cp .env.example .env
# Modifier .env avec les informations de connexion Supabase

# 4. Installer les dépendances
go mod download

# 5. Générer du code type-safe à partir de requêtes SQL
cd ..
./script/gen-sqlc.bash
# Cela génère :
# - Du code Go type-safe pour les services backend (servers/internal/sql/)
# - Des types TypeScript pour les Supabase Edge Functions (supabase/functions/_shared/queries/)

# 6. (Optionnel) Réinitialiser la base de données si nécessaire
./script/reset-local-sb.bash
```

### Exécution des Services

```bash
cd servers

# Service API
go run ./cmd/api

# Service WebSocket
go run ./cmd/ws

# Service statistiques
go run ./cmd/stats

# Service logging
go run ./cmd/logging
```

## Développement

### Build

```bash
cd servers
go build ./...                    # Compiler tous les packages
go build ./cmd/api                # Compiler un service spécifique
```

### Tests

```bash
cd servers
go test ./...                     # Exécuter tous les tests
go test -cover ./...              # Exécuter avec couverture
go test -v ./internal/feature/... # Exécuter les tests d'un package spécifique
```

### Génération de Code

```bash
# Exécuter depuis la racine du dépôt
./script/gen-sqlc.bash           # Générer du code Go et TypeScript type-safe à partir de SQL
                                 # - Go : servers/internal/sql/ (support complet de toutes les annotations SQLC)
                                 # - TypeScript : supabase/functions/_shared/queries/
                                 #   (limitations : :exec, :execrows, :execresult, :batchexec non supportés)
./script/gen-proto.bash          # Générer le code Protocol Buffer
./script/gen-typing-sb.bash      # Générer les types de schéma de base de données TypeScript
```

**IMPORTANT** : Lors de l'écriture de requêtes SQL pour la génération TypeScript, utilisez les annotations `:one` ou `:many` au lieu des annotations de la famille `:exec`. Pour les requêtes qui ne retournent pas de données, utilisez `:one` avec une clause `RETURNING` ou sélectionnez une valeur fictive.

### Gestion de la Base de Données (Supabase)

```bash
# Gestion de l'environnement local Supabase
supabase start                   # Démarrer Supabase local
supabase stop                    # Arrêter Supabase local
supabase status                  # Vérifier le statut de Supabase

# Migrations
supabase db reset                # Réinitialiser la BD locale (ré-exécuter toutes les migrations)
supabase migration new <name>    # Créer une nouvelle migration
supabase db push                 # Appliquer les migrations à la BD distante

# Réinitialisation de la BD via scripts
./script/reset-local-sb.bash     # Réinitialiser la BD Supabase locale et créer les données initiales
./script/reset-remote-sb.bash    # Réinitialiser la BD Supabase distante (utiliser avec précaution !)
```

### Workflow d'Intégration Supabase

Ce projet utilise Supabase comme plateforme de gestion de base de données :

1. **Développement Local** : Exécuter l'environnement PostgreSQL basé sur Docker avec `supabase start`
2. **Gestion du Schéma** : Définir les tables dans `supabase/schemas/`, stocker les migrations dans `supabase/migrations/`
3. **Requêtes Type-Safe** : Générer le code Go à partir du SQL dans `supabase/queries/` en utilisant SQLC
4. **Déploiement** : Appliquer les migrations aux projets distants en utilisant Supabase CLI

**Avantages Clés** :

- Configuration rapide de l'environnement de développement local (basé sur Docker)
- Contrôle de version automatisé des migrations
- Gestion visuelle de la base de données avec Supabase Studio
- Déploiement en production simplifié
- **Génération de code type-safe** : Écrivez SQL une seule fois, générez automatiquement du code Go et TypeScript type-safe via `./script/gen-sqlc.bash`

## Patterns d'Architecture

### Vertical Slice Architecture (Pattern Principal)

L'architecture centrale de ce projet est **Vertical Slice Architecture**. Chaque fonctionnalité est une tranche verticale complète contenant toutes les couches (HTTP → Logique Métier → Accès aux Données).

**Caractéristiques** :

- Haute cohésion par fonctionnalité (tout le code nécessaire pour une fonctionnalité au même endroit)
- Faible couplage (dépendances minimales entre les fonctionnalités)
- Développement et maintenance rapides (travail indépendant par fonctionnalité)

**Exemple de Structure** (`internal/feature/user_profile/`) :

```
internal/feature/user_profile/
  ├── router.go              # Mapping des routes (fonction MapRoutes)
  ├── get_profile/
  │   ├── endpoint.go        # Gestionnaire HTTP (fonction Map)
  │   └── dto.go            # DTOs requête/réponse
  └── update_profile/
      ├── endpoint.go        # Gestionnaire HTTP (fonction Map)
      └── dto.go            # DTOs requête/réponse
```

**Pattern d'Endpoint** :

La fonction `Map` de chaque endpoint gère directement :

1. Extraire le logger et la connexion DB du contexte
2. Parser le corps de la requête en utilisant `httputil.GetReqBodyWithLog`
3. Exécuter la logique métier (requêtes, validation, etc.)
4. Retourner la réponse en utilisant `httputil.OkWithMsg` ou `httputil.ErrWithMsg`

### Patterns de Support

**Structure Basée sur les Composants** (Services WebSocket, Stats, Logging) :

- Structuré par préoccupations techniques (sessions, gestion des paquets, consommation d'événements, etc.)
- Implémentation directe sans séparation en couches

**Architecture Event-Driven** :

- Traitement asynchrone basé sur Redis Streams
- Pattern Consumer-Processor

**Pattern Repository** (`internal/repository/`) :

- Modèle pour l'abstraction de l'accès aux données
- Exemples d'interface CRUD

### Composants Partagés Clés

- **Redis Streams Consumer** : Consommateur d'événements basé sur les génériques
- **Accès à la Base de Données** : Requêtes générées par SQLC ou requêtes pgx directes
- **Utilitaires HTTP** : Gestion standardisée des requêtes/réponses
- **Arrêt Gracieux** : Basé sur l'interface `shared.Closer`

## Points de Terminaison API

### Service API (Port 8080)

- `GET /health` - Vérification de santé
- `GET /ready` - Vérification de disponibilité
- `GET /api/v1/ping` - Ping
- `POST /api/v1/user-profile/get` - Obtenir le profil utilisateur
- `POST /api/v1/user-profile/update` - Mettre à jour le profil utilisateur

### Service WebSocket (Port 8081)

- `GET /health` - Vérification de santé
- `GET /ws` - Connexion WebSocket

### Service Statistiques (Port 8084)

- `GET /health` - Vérification de santé
- `GET /metrics` - Obtenir les métriques

## Licence

Apache License 2.0 - Voir le fichier [LICENSE](LICENSE) pour plus de détails

## Contribution

Les pull requests sont les bienvenues !

## Support

Si vous rencontrez des problèmes, veuillez créer une issue GitHub.
