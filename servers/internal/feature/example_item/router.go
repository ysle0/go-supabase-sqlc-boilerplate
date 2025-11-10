package example_item

import "github.com/go-chi/chi/v5"

// RegisterRoutes registers all item routes
func RegisterRoutes(r chi.Router, handler *Handler) {
	r.Route("/items", func(r chi.Router) {
		r.Get("/", handler.ListItems)
		r.Post("/", handler.CreateItem)
		r.Get("/{id}", handler.GetItem)
		r.Put("/{id}", handler.UpdateItem)
		r.Delete("/{id}", handler.DeleteItem)
	})
}
