package users

import (
	"net/http"

	"example.com/m/middleware"
	"github.com/go-chi/chi/v5"
)

func Routes() http.Handler {
	r := chi.NewRouter()

	r.Group(func(protected chi.Router) {
		protected.Use(middleware.Auth)

		protected.Get("/me", GetMe)
		protected.Put("/me", UpdateMe)
		protected.Get("/badges", GetBadges)
		protected.Get("/challenges", GetChallenges)
	})

	return r
}
