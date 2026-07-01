// Package config loads runtime configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
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

	// Mail — email campaigns. MailProvider is "smtp" (default), "resend", or
	// "brevo". HTTP providers (resend/brevo) work over HTTPS where SMTP ports
	// are blocked. Campaigns are disabled unless a provider is configured.
	MailProvider string
	MailAPIKey   string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string // e.g. "Acme <hello@acme.com>"
	// CampaignDelayMS throttles sends; CampaignMax caps recipients per campaign.
	CampaignDelayMS int
	CampaignMax     int
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

		MailProvider:    getenv("MAIL_PROVIDER", "smtp"),
		MailAPIKey:      os.Getenv("MAIL_API_KEY"),
		SMTPHost:        os.Getenv("SMTP_HOST"),
		SMTPPort:        atoiDefault(os.Getenv("SMTP_PORT"), 587),
		SMTPUsername:    os.Getenv("SMTP_USERNAME"),
		SMTPPassword:    os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:        os.Getenv("SMTP_FROM"),
		CampaignDelayMS: atoiDefault(os.Getenv("CAMPAIGN_DELAY_MS"), 1000),
		CampaignMax:     atoiDefault(os.Getenv("CAMPAIGN_MAX"), 500),
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

func atoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil && v > 0 {
		return v
	}
	return def
}
