// Command server is the composition root: it loads config, wires concrete
// implementations together, and runs the HTTP server.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/parse-companies/backend/internal/api"
	"github.com/parse-companies/backend/internal/cache"
	"github.com/parse-companies/backend/internal/config"
	"github.com/parse-companies/backend/internal/provider/nominatim"
	"github.com/parse-companies/backend/internal/provider/overpass"
	"github.com/parse-companies/backend/internal/provider/wikidata"
	"github.com/parse-companies/backend/internal/service"
	"github.com/parse-companies/backend/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer st.Close()

	rc, err := cache.NewRedis(cfg.RedisURL)
	if err != nil {
		return err
	}

	httpClient := &http.Client{Timeout: 200 * time.Second}
	providers := []service.NamedProvider{
		{Name: "OpenStreetMap", P: overpass.New(httpClient, cfg.OverpassURL)},
		{Name: "Wikidata", P: wikidata.New(&http.Client{Timeout: 90 * time.Second}, cfg.WikidataURL)},
	}
	geocoder := nominatim.New(&http.Client{Timeout: 15 * time.Second}, cfg.NominatimURL)
	svc := service.New(providers, st, rc)

	srv := api.NewServer(svc, st, geocoder, logger)
	handler := srv.Routes(cfg.AllowedOrigin)

	httpSrv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		logger.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
	}()

	logger.Info("starting", "addr", cfg.HTTPAddr)
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
