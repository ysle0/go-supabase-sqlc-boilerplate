package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// MeasureExecTime
func MeasureExecTime(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		start := time.Now()
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				execTime := time.Since(start)
				logger.Info("execTime measure", "execTime", execTime)
			}()
			next.ServeHTTP(w, r)
		})
	}
}
