// Package config loads runtime configuration from the environment.
package config

import (
	"fmt"
	"os"
)

// Config holds all runtime settings. Every field maps to one env var so the
// service stays twelve-factor and the same binary runs in dev and prod.
type Config struct {
	HTTPAddr      string
	DatabaseURL   string
	RedisURL      string
	OverpassURL   string
	NominatimURL  string
	WikidataURL   string
	AllowedOrigin string
}

// Load reads configuration from the environment, applying defaults for the
// non-secret knobs. DatabaseURL and RedisURL are required and error if absent.
func Load() (Config, error) {
	c := Config{
		HTTPAddr:      getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		RedisURL:      os.Getenv("REDIS_URL"),
		OverpassURL:   getenv("OVERPASS_URL", "https://overpass-api.de/api/interpreter"),
		NominatimURL:  getenv("NOMINATIM_URL", "https://nominatim.openstreetmap.org"),
		WikidataURL:   getenv("WIKIDATA_URL", "https://query.wikidata.org/sparql"),
		AllowedOrigin: getenv("ALLOWED_ORIGIN", "http://localhost:3000"),
	}
	if c.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: DATABASE_URL is required")
	}
	if c.RedisURL == "" {
		return Config{}, fmt.Errorf("config: REDIS_URL is required")
	}
	return c, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
