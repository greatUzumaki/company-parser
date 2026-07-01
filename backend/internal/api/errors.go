package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/parse-companies/backend/internal/export"
	"github.com/parse-companies/backend/internal/provider"
)

// writeJSON writes v as JSON with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError maps a domain/infra error to an HTTP status + JSON body. Typed
// sentinels drive the mapping; everything else is a 500. No error is swallowed.
func writeError(w http.ResponseWriter, logger *slog.Logger, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, provider.ErrUpstreamBusy):
		status = http.StatusServiceUnavailable
	case errors.Is(err, export.ErrBadFormat):
		status = http.StatusBadRequest
	case errors.Is(err, pgx.ErrNoRows):
		status = http.StatusNotFound
	}
	if status == http.StatusInternalServerError {
		logger.Error("request failed", "err", err)
	}
	writeJSON(w, status, map[string]string{"error": http.StatusText(status)})
}
