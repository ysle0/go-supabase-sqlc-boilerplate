package middleware

import (
	"context"
	"net/http"
)

type ApiVersionKey struct{}
type ApiVersionVal struct {
	version string
}

func SetApiVersionToHeader(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(
		"api-version",
		r.Context().Value(ApiVersionKey{}).(ApiVersionVal).version,
	)
}

func ApiVersionWith(version string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), ApiVersionKey{}, ApiVersionVal{version}))
			next.ServeHTTP(w, r)
		})
	}
}
