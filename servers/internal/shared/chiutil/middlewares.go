package chiutil

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared/ratelimit"
)

func UseBasicMiddlewares(
	ctx context.Context,
	r *chi.Mux,
	logger *slog.Logger,
	requestTimeout time.Duration,
) {
	rl := ratelimit.NewRateLimiter()
	r.Use(rl.LimitByRequest)

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Timeout(requestTimeout))
	r.Use(middleware.Heartbeat("/heartbeat"))
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.Compress(5, "gzip", "deflate"))
	r.Use(httplog.RequestLogger(logger, &httplog.Options{
		// Debug - log all responses (incl. OPTIONS)
		// Info - log responses (excl. OPTIONS)
		// Warn - log 4xx and 5xx responses only (except for 429)
		// Error - log 5xx responses only
		Level:  slog.LevelInfo,
		Schema: httplog.SchemaOTEL,
		// returns 500 on panic, unless the response status was set.
		// panics are automatically logged, regardless of this setting.
		RecoverPanics: true,
		LogRequestBody: func(req *http.Request) bool {
			return true
		},
		LogResponseBody: func(req *http.Request) bool {
			return true
		},
	}))
	//r.Use(custom_middleware.MeasureExecTime(logger))

	r.Use(middleware.WithValue("logger", logger))
}
