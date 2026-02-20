package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
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

	cfg, result, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Created {
		t.Error("expected Created=false for empty path")
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

	cfg, result, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Created {
		t.Error("expected Created=false for existing file")
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

	cfg, _, err := Load(path)
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

	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	cfg, result, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error for missing file: %v", err)
	}
	if !result.Created {
		t.Error("expected Created=true when file did not exist")
	}

	// Should use defaults
	if cfg.ProPresenterHost != "localhost" {
		t.Errorf("expected default ProPresenterHost=localhost, got %s", cfg.ProPresenterHost)
	}

	// File should exist on disk now
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected config file to be created at %s: %v", path, err)
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	clearEnv(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(path, []byte("{{invalid toml"), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, _, err := Load(path)
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

	cfg, _, err := Load(path)
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

func TestMergeNewKeys(t *testing.T) {
	clearEnv(t)

	// Config with only two keys — the rest should be merged.
	tomlContent := `propresenter_host = "10.0.0.5"
propresenter_port = "9999"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, result, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Created {
		t.Error("expected Created=false for existing file")
	}

	// Should have merged the missing keys.
	expected := []string{
		"listen_addr", "children_file", "message_name",
		"auto_clear_seconds", "activity_log", "auth_token",
	}
	if len(result.MergedKeys) != len(expected) {
		t.Fatalf("expected %d merged keys, got %d: %v", len(expected), len(result.MergedKeys), result.MergedKeys)
	}
	for i, key := range expected {
		if result.MergedKeys[i] != key {
			t.Errorf("merged key %d: expected %s, got %s", i, key, result.MergedKeys[i])
		}
	}

	// Backup should exist.
	if result.BackupPath == "" {
		t.Fatal("expected a backup path")
	}
	bakContent, err := os.ReadFile(result.BackupPath)
	if err != nil {
		t.Fatalf("failed to read backup: %v", err)
	}
	if string(bakContent) != tomlContent {
		t.Error("backup content does not match original")
	}

	// User values should be preserved.
	if cfg.ProPresenterHost != "10.0.0.5" {
		t.Errorf("expected ProPresenterHost=10.0.0.5, got %s", cfg.ProPresenterHost)
	}
	if cfg.ProPresenterPort != "9999" {
		t.Errorf("expected ProPresenterPort=9999, got %s", cfg.ProPresenterPort)
	}

	// The merged file should be valid TOML when re-parsed.
	var verifyConfig Config
	if _, err := toml.DecodeFile(path, &verifyConfig); err != nil {
		t.Fatalf("merged config is not valid TOML: %v", err)
	}
}

func TestMergeNoChangesWhenComplete(t *testing.T) {
	clearEnv(t)

	// Config that already has all keys (including commented-out ones).
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(generateDefaultConfig()), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, result, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.MergedKeys) != 0 {
		t.Errorf("expected no merged keys for complete config, got %v", result.MergedKeys)
	}
	if result.BackupPath != "" {
		t.Errorf("expected no backup for complete config, got %s", result.BackupPath)
	}
}

func TestMergePreservesUserValues(t *testing.T) {
	clearEnv(t)

	tomlContent := `propresenter_host = "custom-host"
propresenter_port = "12345"
listen_addr = ":9090"
children_file = "my-kids.json"
message_name = "Custom Msg"
auto_clear_seconds = 99
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, result, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only activity_log and auth_token should be merged (they were commented-out defaults).
	if len(result.MergedKeys) != 2 {
		t.Fatalf("expected 2 merged keys, got %d: %v", len(result.MergedKeys), result.MergedKeys)
	}

	// All custom values must be preserved.
	if cfg.ProPresenterHost != "custom-host" {
		t.Errorf("got ProPresenterHost=%s", cfg.ProPresenterHost)
	}
	if cfg.ProPresenterPort != "12345" {
		t.Errorf("got ProPresenterPort=%s", cfg.ProPresenterPort)
	}
	if cfg.ListenAddr != ":9090" {
		t.Errorf("got ListenAddr=%s", cfg.ListenAddr)
	}
	if cfg.ChildrenFile != "my-kids.json" {
		t.Errorf("got ChildrenFile=%s", cfg.ChildrenFile)
	}
	if cfg.MessageName != "Custom Msg" {
		t.Errorf("got MessageName=%s", cfg.MessageName)
	}
	if cfg.AutoClearSeconds != 99 {
		t.Errorf("got AutoClearSeconds=%d", cfg.AutoClearSeconds)
	}
}

func TestGenerateDefaultConfig(t *testing.T) {
	content := generateDefaultConfig()

	// Must start with the header.
	if !strings.Contains(content, "# Calling Parents") {
		t.Error("missing header in generated config")
	}

	// Must contain every key from allConfigBlocks.
	for _, block := range allConfigBlocks {
		if !strings.Contains(content, block.key) {
			t.Errorf("generated config missing key %s", block.key)
		}
	}
}
