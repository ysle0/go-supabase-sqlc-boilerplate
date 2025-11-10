package helpers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	PostgresImage = "postgres:16-alpine"
	RedisImage    = "redis:8-alpine"
	PostgresDB    = "testdb"
	PostgresUser  = "testuser"
	PostgresPass  = "testpass"
)

type TestContainers struct {
	PostgresContainer *postgres.PostgresContainer
	RedisContainer    testcontainers.Container
	PostgresConnStr   string
	RedisAddr         string
	DBPool            *pgxpool.Pool
	RedisClient       *redis.Client
}

// StartPostgresContainer starts a PostgreSQL container and runs migrations
func StartPostgresContainer(ctx context.Context) (*postgres.PostgresContainer, string, error) {
	pgContainer, err := postgres.Run(
		ctx,
		PostgresImage,
		postgres.WithDatabase(PostgresDB),
		postgres.WithUsername(PostgresUser),
		postgres.WithPassword(PostgresPass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to start postgres container: %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, "", fmt.Errorf("failed to get connection string: %w", err)
	}

	return pgContainer, connStr, nil
}

// StartRedisContainer starts a Redis container
func StartRedisContainer(ctx context.Context) (testcontainers.Container, string, error) {
	redisContainer, err := rediscontainer.Run(
		ctx,
		RedisImage,
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to start redis container: %w", err)
	}

	host, err := redisContainer.Host(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get redis host: %w", err)
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		return nil, "", fmt.Errorf("failed to get redis port: %w", err)
	}

	redisAddr := fmt.Sprintf("%s:%s", host, port.Port())
	return redisContainer, redisAddr, nil
}

// CreateSupabaseRoles creates Supabase-specific roles needed for RLS policies
func CreateSupabaseRoles(ctx context.Context, pool *pgxpool.Pool) error {
	roles := []string{
		"CREATE ROLE anon NOLOGIN NOINHERIT",
		"CREATE ROLE authenticated NOLOGIN NOINHERIT",
		"CREATE ROLE service_role NOLOGIN NOINHERIT BYPASSRLS",
		"CREATE ROLE supabase_auth_admin NOLOGIN NOINHERIT",
		"CREATE ROLE supabase_storage_admin NOLOGIN NOINHERIT",
		"CREATE ROLE dashboard_user NOLOGIN",
	}

	for _, roleSQL := range roles {
		_, err := pool.Exec(ctx, roleSQL)
		if err != nil {
			// Ignore "role already exists" errors
			if !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("failed to create role: %w", err)
			}
		}
	}

	return nil
}

// RunSchemaFile executes a single schema SQL file
func RunSchemaFile(ctx context.Context, pool *pgxpool.Pool, schemaPath string) error {
	// Create Supabase roles first
	if err := CreateSupabaseRoles(ctx, pool); err != nil {
		return fmt.Errorf("failed to create Supabase roles: %w", err)
	}

	// Read schema file
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file %s: %w", schemaPath, err)
	}

	// Execute schema
	_, err = pool.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed to execute schema %s: %w", schemaPath, err)
	}

	return nil
}

// RunMigrations executes all SQL migration files in order
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsPath string) error {
	// Create Supabase roles first
	if err := CreateSupabaseRoles(ctx, pool); err != nil {
		return fmt.Errorf("failed to create Supabase roles: %w", err)
	}

	// Get all migration files
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to find migration files: %w", err)
	}

	// Sort files by name (timestamp-based naming ensures correct order)
	sort.Strings(files)

	// Execute each migration file
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Execute migration
		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}

	return nil
}

// SetupTestContainers initializes both PostgreSQL and Redis containers with migrations
func SetupTestContainers(ctx context.Context) (*TestContainers, error) {
	tc := &TestContainers{}

	// Start PostgreSQL container
	pgContainer, connStr, err := StartPostgresContainer(ctx)
	if err != nil {
		return nil, err
	}
	tc.PostgresContainer = pgContainer
	tc.PostgresConnStr = connStr

	// Start Redis container
	redisContainer, redisAddr, err := StartRedisContainer(ctx)
	if err != nil {
		// Cleanup PostgreSQL on Redis failure
		_ = pgContainer.Terminate(ctx)
		return nil, err
	}
	tc.RedisContainer = redisContainer
	tc.RedisAddr = redisAddr

	// Create database connection pool
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		_ = tc.Cleanup(ctx)
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = tc.Cleanup(ctx)
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	tc.DBPool = pool

	// Run schema.sql instead of migrations (migrations have game_stages, but queries expect stages)
	schemaPath := "/Users/ysl/bedrijf/spreadit/server/supabase/schemas/schema.sql"
	if err := RunSchemaFile(ctx, pool, schemaPath); err != nil {
		_ = tc.Cleanup(ctx)
		return nil, fmt.Errorf("failed to run schema: %w", err)
	}

	// Create Redis client
	tc.RedisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Test Redis connection
	if err := tc.RedisClient.Ping(ctx).Err(); err != nil {
		_ = tc.Cleanup(ctx)
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return tc, nil
}

// Cleanup terminates all containers and closes connections
func (tc *TestContainers) Cleanup(ctx context.Context) error {
	var errs []string

	if tc.RedisClient != nil {
		if err := tc.RedisClient.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("redis client close: %v", err))
		}
	}

	if tc.DBPool != nil {
		tc.DBPool.Close()
	}

	if tc.RedisContainer != nil {
		if err := tc.RedisContainer.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("redis container: %v", err))
		}
	}

	if tc.PostgresContainer != nil {
		if err := tc.PostgresContainer.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("postgres container: %v", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// TruncateAllTables removes all data from all tables (for test cleanup)
func (tc *TestContainers) TruncateAllTables(ctx context.Context) error {
	query := `
		DO $$
		DECLARE
			r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
				EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' RESTART IDENTITY CASCADE';
			END LOOP;
		END $$;
	`
	_, err := tc.DBPool.Exec(ctx, query)
	return err
}

// SetEnvironmentVariables sets the environment variables for the application to use test containers
func (tc *TestContainers) SetEnvironmentVariables() error {
	// Set PostgreSQL URL
	if err := os.Setenv("POSTGRESQL_URL", tc.PostgresConnStr); err != nil {
		return fmt.Errorf("failed to set POSTGRESQL_URL: %w", err)
	}

	// Set Redis URLs (all pointing to the same test Redis for simplicity)
	redisURL := fmt.Sprintf("redis://%s/0", tc.RedisAddr)
	envVars := []string{
		"RANKING_REDIS_URL",
		"LOGGING_REDIS_URL",
		"CACHE_REDIS_URL",
		"QUESTION_REDIS_URL",
	}

	for _, envVar := range envVars {
		if err := os.Setenv(envVar, redisURL); err != nil {
			return fmt.Errorf("failed to set %s: %w", envVar, err)
		}
	}

	return nil
}
