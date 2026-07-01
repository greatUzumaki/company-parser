package api

import "net/http"

// Routes builds the HTTP handler with all routes and middleware.
func (s *Server) Routes(allowedOrigin string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /api/v1/search", s.handleSearch)
	mux.HandleFunc("POST /api/v1/search/stream", s.handleSearchStream)
	mux.HandleFunc("GET /api/v1/searches", s.handleListSearches)
	mux.HandleFunc("GET /api/v1/searches/{id}", s.handleGetSearch)
	mux.HandleFunc("GET /api/v1/searches/{id}/export", s.handleExport)
	mux.HandleFunc("GET /api/v1/regions", s.handleRegions)
	mux.HandleFunc("GET /api/v1/categories", s.handleCategories)
	mux.HandleFunc("GET /api/v1/campaign/status", s.handleCampaignStatus)
	mux.HandleFunc("POST /api/v1/campaign/send", s.handleCampaignSend)

	return cors(allowedOrigin)(mux)
}

// cors allows the frontend origin to call the API from the browser.
func cors(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
