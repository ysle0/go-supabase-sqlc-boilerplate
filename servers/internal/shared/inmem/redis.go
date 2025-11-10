// package inmem -
package inmem

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/MatusOllah/slogcolor"
	"github.com/redis/go-redis/v9"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared"
)

type MemDBVal struct {
	RankingRedis  *redis.Client
	LoggingRedis  *redis.Client
	CacheRedis    *redis.Client
	QuestionRedis *redis.Client
}

var (
	RankingKey       = 0
	LoggingKey       = 1
	CacheKey         = 2
	QuestionKey      = 3
	cacheInitOnce    sync.Once
	loggingInitOnce  sync.Once
	rankingInitOnce  sync.Once
	questionInitOnce sync.Once
	rankingClient    *redis.Client
	loggingClient    *redis.Client
	cacheClient      *redis.Client
	questionClient   *redis.Client
	logger           = slog.New(slogcolor.NewHandler(os.Stdout, &slogcolor.Options{
		Level:       slog.LevelDebug,
		SrcFileMode: slogcolor.ShortFile,
	}))
)

func PullClient(r *http.Request, kind int) *redis.Client {
	var k any
	switch kind {
	case CacheKey:
		k = CacheKey
	case RankingKey:
		k = RankingKey
	case LoggingKey:
		k = LoggingKey
	case QuestionKey:
		k = QuestionKey
	}
	if val, ok := r.Context().Value(k).(MemDBVal); !ok {
		return nil
	} else {
		return val.RankingRedis
	}
}

// GetClient - returns redis client
func GetClient(ctx context.Context, kind int) *redis.Client {
	switch kind {
	case CacheKey:
		cacheInitOnce.Do(func() {
			cacheRedisURL := shared.EnvString("CACHE_REDIS_URL", "")
			cacheClient = connect(ctx, cacheRedisURL, true)
		})
		return cacheClient
	case RankingKey:
		rankingInitOnce.Do(func() {
			rankingRedisURL := shared.EnvString("RANKING_REDIS_URL", "")
			rankingClient = connect(ctx, rankingRedisURL, true)
		})
		return rankingClient
	case LoggingKey:
		loggingInitOnce.Do(func() {
			loggingRedisURL := shared.EnvString("LOGGING_REDIS_URL", "")
			loggingClient = connect(ctx, loggingRedisURL, true)
		})
		return loggingClient
	case QuestionKey:
		questionInitOnce.Do(func() {
			questionRedisURL := shared.EnvString("QUESTION_REDIS_URL", "")
			questionClient = connect(ctx, questionRedisURL, true)
		})
		return questionClient
	default:
		logger.Error(fmt.Sprintf("failed to GetClient(): %d", kind))
		return nil
	}
}

func connect(ctx context.Context, url string, bPerformPing bool) *redis.Client {
	logger.Debug("Start Connecting to Redis", "url", url)
	opt, err := redis.ParseURL(url)
	if err != nil {
		logger.Info("failed to parse Redis URL", "error", err)
		return nil
	}
	logger.Debug("Connecting to Redis", "addr", opt.Addr, "db", opt.DB)

	client := redis.NewClient(opt)

	if bPerformPing {
		if pong, err := client.Ping(ctx).Result(); err != nil {
			logger.Error("failed to connect Redis", "error", err)
			return nil
		} else {
			logger.Info("Redis connected", "pong", pong)
		}
	} else {
		logger.Info("Redis connected")
	}

	return client
}
