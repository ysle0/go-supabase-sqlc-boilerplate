package user_profile

import (
	"github.com/go-chi/chi/v5"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/feature/user_profile/get_profile"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/feature/user_profile/update_profile"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared/middleware"
)

func MapRoutes(r chi.Router, apiVersion string) {
	r.Route("/"+apiVersion+"/user-profile", func(r chi.Router) {
		r.Use(middleware.ApiVersionWith(apiVersion))

		r.Post("/get", get_profile.Map)
		r.Post("/update", update_profile.Map)
	})
}
