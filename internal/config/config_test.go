package config

import (
	"os"
	"path/filepath"
	"testing"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"PROPRESENTER_HOST", "PROPRESENTER_PORT", "LISTEN_ADDR",
		"CHILDREN_FILE", "AUTH_TOKEN", "MESSAGE_NAME",
		"AUTO_CLEAR_SECONDS", "ACTIVITY_LOG",
	} {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
	if cfg.MessageName != "Eltern rufen" {
		t.Errorf("expected MessageName=Eltern rufen, got %s", cfg.MessageName)
	}
	if cfg.AutoClearSeconds != 30 {
		t.Errorf("expected AutoClearSeconds=30, got %d", cfg.AutoClearSeconds)
	}
	if cfg.ActivityLog != "" {
		t.Errorf("expected ActivityLog=empty, got %s", cfg.ActivityLog)
	}
}

func TestLoadFromTOML(t *testing.T) {
	clearEnv(t)

	tomlContent := `
propresenter_host = "192.168.1.50"
propresenter_port = "9999"
listen_addr = ":3000"
children_file = "kids.json"
auth_token = "toml-secret"
message_name = "Custom Message"
auto_clear_seconds = 60
activity_log = "log.jsonl"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ProPresenterHost != "192.168.1.50" {
		t.Errorf("expected ProPresenterHost=192.168.1.50, got %s", cfg.ProPresenterHost)
	}
	if cfg.ProPresenterPort != "9999" {
		t.Errorf("expected ProPresenterPort=9999, got %s", cfg.ProPresenterPort)
	}
	if cfg.ListenAddr != ":3000" {
		t.Errorf("expected ListenAddr=:3000, got %s", cfg.ListenAddr)
	}
	if cfg.ChildrenFile != "kids.json" {
		t.Errorf("expected ChildrenFile=kids.json, got %s", cfg.ChildrenFile)
	}
	if cfg.AuthToken != "toml-secret" {
		t.Errorf("expected AuthToken=toml-secret, got %s", cfg.AuthToken)
	}
	if cfg.MessageName != "Custom Message" {
		t.Errorf("expected MessageName=Custom Message, got %s", cfg.MessageName)
	}
	if cfg.AutoClearSeconds != 60 {
		t.Errorf("expected AutoClearSeconds=60, got %d", cfg.AutoClearSeconds)
	}
	if cfg.ActivityLog != "log.jsonl" {
		t.Errorf("expected ActivityLog=log.jsonl, got %s", cfg.ActivityLog)
	}
}

func TestEnvOverridesToml(t *testing.T) {
	tomlContent := `
propresenter_host = "192.168.1.50"
message_name = "TOML Name"
auto_clear_seconds = 10
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	clearEnv(t)
	t.Setenv("PROPRESENTER_HOST", "10.0.0.1")
	t.Setenv("AUTO_CLEAR_SECONDS", "120")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Env should override TOML
	if cfg.ProPresenterHost != "10.0.0.1" {
		t.Errorf("expected env override ProPresenterHost=10.0.0.1, got %s", cfg.ProPresenterHost)
	}
	if cfg.AutoClearSeconds != 120 {
		t.Errorf("expected env override AutoClearSeconds=120, got %d", cfg.AutoClearSeconds)
	}
	// TOML value should remain where no env set
	if cfg.MessageName != "TOML Name" {
		t.Errorf("expected TOML MessageName=TOML Name, got %s", cfg.MessageName)
	}
}

func TestLoadMissingFile(t *testing.T) {
	clearEnv(t)

	cfg, err := Load("/nonexistent/config.toml")
	if err != nil {
		t.Fatalf("unexpected error for missing file: %v", err)
	}

	// Should fall back to defaults
	if cfg.ProPresenterHost != "localhost" {
		t.Errorf("expected default ProPresenterHost=localhost, got %s", cfg.ProPresenterHost)
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	clearEnv(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(path, []byte("{{invalid toml"), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid TOML, got nil")
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

func TestLoadPartialTOML(t *testing.T) {
	clearEnv(t)

	tomlContent := `
propresenter_host = "10.0.0.5"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Specified value
	if cfg.ProPresenterHost != "10.0.0.5" {
		t.Errorf("expected ProPresenterHost=10.0.0.5, got %s", cfg.ProPresenterHost)
	}
	// Defaults for unspecified
	if cfg.ProPresenterPort != "50001" {
		t.Errorf("expected default ProPresenterPort=50001, got %s", cfg.ProPresenterPort)
	}
	if cfg.AutoClearSeconds != 30 {
		t.Errorf("expected default AutoClearSeconds=30, got %d", cfg.AutoClearSeconds)
	}
}
