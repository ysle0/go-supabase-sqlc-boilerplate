package supabase_postgres

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/MatusOllah/slogcolor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared"
)

var (
	maxConn     int32 = 10
	minConn     int32 = 1
	maxIdleTime       = 10 * time.Second
	maxLifetime       = 30 * time.Second
	logger            = slog.New(slogcolor.NewHandler(os.Stdout, &slogcolor.Options{
		Level:       slog.LevelDebug,
		SrcFileMode: slogcolor.Nop,
		TimeFormat:  time.DateTime,
	}))
	pool *DBPooler
	once sync.Once
)

// DBKey -
type DBKey struct{}

// DBVal -
type DBVal struct {
	Pooler *DBPooler
}

type DBPooler struct {
	Pool *pgxpool.Pool
}

func PullDbPooler(ctx context.Context) *DBPooler {
	return ctx.Value(DBKey{}).(DBVal).Pooler
}

func GetDBPooler() *DBPooler {
	once.Do(func() {
		postgresqlURL := shared.EnvString("POSTGRESQL_URL", "postgresql://postgres:postgres@127.0.0.1:54322/postgres")

		cfg, err := pgxpool.ParseConfig(postgresqlURL)
		if err != nil {
			logger.Error("failed to parse Supabase PostgreSQL URL to config", "error", err)
			return
		}

		logger.Debug("Postgresql connection:", "database_url", postgresqlURL)

		cfg.MaxConns = maxConn
		cfg.MinConns = minConn
		cfg.MaxConnIdleTime = maxIdleTime
		cfg.MaxConnLifetime = maxLifetime
		cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

		// Debug Tracing Logs

		if os.Getenv("RUN_INTEGRATION_TESTS") == "false" {
			cfg.ConnConfig.Tracer = &QueryTracer{logger: logger}
		}

		//cfg.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		//	// log.Printf("[pgx] acquiring connection")
		//	return true
		//}
		//
		//cfg.AfterRelease = func(conn *pgx.Conn) bool {
		//	// log.Printf("[pgx] releasing connection")
		//	return true
		//}
		//
		//cfg.BeforeConnect = func(ctx context.Context, cfg *pgx.ConnConfig) error {
		//	// log.Printf("[pgx] connecting to database")
		//	return nil
		//}

		p, err := pgxpool.NewWithConfig(context.Background(), cfg)
		if err != nil {
			logger.Error("failed to create connection pool", "error", err)
			return
		}

		pool = &DBPooler{Pool: p}
	})
	return pool
}
