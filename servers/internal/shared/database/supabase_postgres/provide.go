package supabase_postgres

import (
	"context"
	"net/http"
)

// Provide - Provides a PostgreSQL database connection pool in request context.
func Provide(ctx context.Context) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(
				r.Context(),
				DBKey{},
				DBVal{Pooler: GetDBPooler()},
			)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
