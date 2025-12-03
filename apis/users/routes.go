package users

import (
	"net/http"

	"example.com/m/apis/useractions"
	"example.com/m/middleware"
	"github.com/go-chi/chi/v5"
)

func Routes() http.Handler {
	r := chi.NewRouter()

	r.Group(func(protected chi.Router) {
		protected.Use(middleware.Auth)

		protected.Get("/profile", GetProfileHandler)
		protected.Put("/profile", UpdateProfileHandler)
		protected.Post("/avatar", UploadAvatarHandler)
		protected.Get("/stats", GetStatsHandler)
		protected.Get("/saved-places", useractions.GetSavedPlacesHandler)
		protected.Post("/save-place", useractions.SavePlaceHandler)
		protected.Delete("/save-place/{place_id}", useractions.RemoveSavedPlaceHandler)
		protected.Post("/visit", useractions.LogVisitHandler)
		protected.Get("/visit-history", useractions.GetVisitHistoryHandler)
	})

	return r
}
