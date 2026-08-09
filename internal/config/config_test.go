package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadFromEnvDefaults(t *testing.T) {
	os.Unsetenv("LISTEN_ADDR")
	os.Unsetenv("ROMM_URL")
	os.Unsetenv("ROMM_USER")
	os.Unsetenv("ROMM_PASS")
	os.Unsetenv("ROMM_TOKEN")
	os.Unsetenv("PUID")
	os.Unsetenv("PGID")
	os.Unsetenv("TZ")
	os.Unsetenv("SCRAPE_TIMEOUT")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ListenAddr != ":8585" {
		t.Errorf("expected :8585, got %s", cfg.ListenAddr)
	}
	if cfg.PUID != 1000 {
		t.Errorf("expected PUID 1000, got %d", cfg.PUID)
	}
	if cfg.PGID != 1000 {
		t.Errorf("expected PGID 1000, got %d", cfg.PGID)
	}
	if cfg.ScrapeTimeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", cfg.ScrapeTimeout)
	}
}

func TestLoadFromEnvCustom(t *testing.T) {
	t.Setenv("LISTEN_ADDR", ":9090")
	t.Setenv("ROMM_URL", "http://romm.example.com/")
	t.Setenv("ROMM_TOKEN", "rmm_12345")
	t.Setenv("PUID", "1001")
	t.Setenv("PGID", "1002")
	t.Setenv("SCRAPE_TIMEOUT", "15s")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ListenAddr != ":9090" {
		t.Errorf("expected :9090, got %s", cfg.ListenAddr)
	}
	if cfg.RommURL != "http://romm.example.com" {
		t.Errorf("expected trimmed url http://romm.example.com, got %s", cfg.RommURL)
	}
	if cfg.RommToken != "rmm_12345" {
		t.Errorf("expected rmm_12345, got %s", cfg.RommToken)
	}
	if cfg.PUID != 1001 {
		t.Errorf("expected PUID 1001, got %d", cfg.PUID)
	}
	if cfg.PGID != 1002 {
		t.Errorf("expected PGID 1002, got %d", cfg.PGID)
	}
	if cfg.ScrapeTimeout != 15*time.Second {
		t.Errorf("expected timeout 15s, got %v", cfg.ScrapeTimeout)
	}
}

func TestLoadFromEnvInvalidPUID(t *testing.T) {
	t.Setenv("PUID", "-5")
	_, err := LoadFromEnv()
	if err == nil {
		t.Errorf("expected error for invalid PUID")
	}
}

func TestLoadFromEnvInvalidTimeout(t *testing.T) {
	t.Setenv("SCRAPE_TIMEOUT", "invalid")
	_, err := LoadFromEnv()
	if err == nil {
		t.Errorf("expected error for invalid SCRAPE_TIMEOUT")
	}
}
