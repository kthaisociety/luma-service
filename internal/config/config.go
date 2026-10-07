// Package config reads every setting from the environment, once, at boot.
package config

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

type Config struct {
	// Host is the bind address. Empty binds all interfaces, which the container needs. Set it to
	// 127.0.0.1 for local dev: a dev machine with a public IP exposes anything bound to 0.0.0.0.
	Host string
	Port string

	// DatabaseURL is a Postgres connection string, e.g.
	// postgres://luma:luma@127.0.0.1:5434/luma?sslmode=disable.
	DatabaseURL string

	// ServiceSecret is the shared secret callers send as X-Service-Secret. Every caller holds the
	// same value (secret/shared/luma-service-secret/<env> in OpenBao).
	ServiceSecret string

	// LumaAPIKey is a calendar API key: it decides which Luma calendar this service reads.
	LumaAPIKey string
	// LumaCommunityTierID is the membership tier served at /v1/community/members.
	LumaCommunityTierID string
}

func Load() *Config {
	return &Config{
		Host:                getEnv("HOST", ""),
		Port:                getEnv("PORT", "8000"),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		ServiceSecret:       getEnv("LUMA_SERVICE_SECRET", ""),
		LumaAPIKey:          getEnv("LUMA_API_KEY", ""),
		LumaCommunityTierID: getEnv("LUMA_COMMUNITY_TIER_ID", ""),
	}
}

// Validate fails if a required setting is missing, naming every missing variable at once.
func (c *Config) Validate() error {
	var missing []string
	for name, value := range map[string]string{
		"DATABASE_URL":           c.DatabaseURL,
		"LUMA_SERVICE_SECRET":    c.ServiceSecret,
		"LUMA_API_KEY":           c.LumaAPIKey,
		"LUMA_COMMUNITY_TIER_ID": c.LumaCommunityTierID,
	} {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		slices.Sort(missing)
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
