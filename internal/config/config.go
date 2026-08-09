package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all exporter configuration.
type Config struct {
	ListenAddr    string
	RommURL       string
	RommUser      string
	RommPass      string
	RommToken     string
	PUID          int
	PGID          int
	TZ            string
	ScrapeTimeout time.Duration
}

// LoadFromEnv loads configuration options from environment variables.
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		ListenAddr:    getEnv("LISTEN_ADDR", ":8585"),
		RommURL:       strings.TrimSuffix(getEnv("ROMM_URL", ""), "/"),
		RommUser:      getEnv("ROMM_USER", ""),
		RommPass:      getEnv("ROMM_PASS", ""),
		RommToken:     getEnv("ROMM_TOKEN", ""),
		PUID:          1000,
		PGID:          1000,
		TZ:            getEnv("TZ", ""),
		ScrapeTimeout: 10 * time.Second,
	}

	if puidStr := getEnv("PUID", ""); puidStr != "" {
		val, err := strconv.Atoi(puidStr)
		if err != nil || val < 0 {
			return nil, fmt.Errorf("invalid PUID value %q: must be a non-negative integer", puidStr)
		}
		cfg.PUID = val
	}

	if pgidStr := getEnv("PGID", ""); pgidStr != "" {
		val, err := strconv.Atoi(pgidStr)
		if err != nil || val < 0 {
			return nil, fmt.Errorf("invalid PGID value %q: must be a non-negative integer", pgidStr)
		}
		cfg.PGID = val
	}

	if timeoutStr := getEnv("SCRAPE_TIMEOUT", ""); timeoutStr != "" {
		d, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid SCRAPE_TIMEOUT value %q: %w", timeoutStr, err)
		}
		cfg.ScrapeTimeout = d
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return fallback
}
