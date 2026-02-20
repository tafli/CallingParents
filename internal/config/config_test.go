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
	os.Unsetenv("CHILDREN_FILE")
	os.Unsetenv("AUTH_TOKEN")

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
	if cfg.ChildrenFile != "children.json" {
		t.Errorf("expected ChildrenFile=children.json, got %s", cfg.ChildrenFile)
	}
	if cfg.AuthToken != "" {
		t.Errorf("expected AuthToken=empty, got %s", cfg.AuthToken)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("PROPRESENTER_HOST", "192.168.1.100")
	os.Setenv("PROPRESENTER_PORT", "9999")
	os.Setenv("LISTEN_ADDR", ":3000")
	os.Setenv("AUTH_TOKEN", "my-secret-token")
	defer func() {
		os.Unsetenv("PROPRESENTER_HOST")
		os.Unsetenv("PROPRESENTER_PORT")
		os.Unsetenv("LISTEN_ADDR")
		os.Unsetenv("AUTH_TOKEN")
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
	if cfg.AuthToken != "my-secret-token" {
		t.Errorf("expected AuthToken=my-secret-token, got %s", cfg.AuthToken)
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
