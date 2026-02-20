package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Unset any env vars that might be set.
	os.Unsetenv("PROPRESENTER_HOST")
	os.Unsetenv("PROPRESENTER_PORT")
	os.Unsetenv("LISTEN_ADDR")

	cfg := Load()

	if cfg.ProPresenterHost != "localhost" {
		t.Errorf("expected ProPresenterHost=localhost, got %s", cfg.ProPresenterHost)
	}
	if cfg.ProPresenterPort != "50001" {
		t.Errorf("expected ProPresenterPort=50001, got %s", cfg.ProPresenterPort)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected ListenAddr=:8080, got %s", cfg.ListenAddr)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("PROPRESENTER_HOST", "192.168.1.100")
	os.Setenv("PROPRESENTER_PORT", "9999")
	os.Setenv("LISTEN_ADDR", ":3000")
	defer func() {
		os.Unsetenv("PROPRESENTER_HOST")
		os.Unsetenv("PROPRESENTER_PORT")
		os.Unsetenv("LISTEN_ADDR")
	}()

	cfg := Load()

	if cfg.ProPresenterHost != "192.168.1.100" {
		t.Errorf("expected ProPresenterHost=192.168.1.100, got %s", cfg.ProPresenterHost)
	}
	if cfg.ProPresenterPort != "9999" {
		t.Errorf("expected ProPresenterPort=9999, got %s", cfg.ProPresenterPort)
	}
	if cfg.ListenAddr != ":3000" {
		t.Errorf("expected ListenAddr=:3000, got %s", cfg.ListenAddr)
	}
}

func TestProPresenterURL(t *testing.T) {
	cfg := Config{
		ProPresenterHost: "10.0.0.5",
		ProPresenterPort: "50001",
	}
	expected := "http://10.0.0.5:50001"
	if got := cfg.ProPresenterURL(); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}
