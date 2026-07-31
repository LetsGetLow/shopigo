package httptransport

import (
	"net/http"
)

// RouteRegistrar registers routes on a ServeMux.
type RouteRegistrar interface {
	RegisterRoutes(*http.ServeMux)
}

// NewRouter builds the API v1 router.
func NewRouter(registrars ...RouteRegistrar) http.Handler {
	apiV1Mux := http.NewServeMux()
	for _, registrar := range registrars {
		registrar.RegisterRoutes(apiV1Mux)
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiV1Mux))
	mux.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/v1/", http.StatusMovedPermanently)
	})

	return mux
}
