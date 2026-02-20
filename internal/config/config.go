package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// configBlock defines a single config key with its descriptive comment block.
// Used for generating the default config file and appending missing keys to existing files.
type configBlock struct {
	key  string
	text string
}

// allConfigBlocks lists every known config key with its comment+default block.
// Order here determines the order they appear when appended.
var allConfigBlocks = []configBlock{
	{"propresenter_host", "# Hostname or IP of the ProPresenter machine.\npropresenter_host = \"localhost\"\n"},
	{"propresenter_port", "# ProPresenter API port (default in ProPresenter: 50001).\npropresenter_port = \"50001\"\n"},
	{"listen_addr", "# Address and port this server listens on.\nlisten_addr = \":8080\"\n"},
	{"children_file", "# Path to the JSON file with children's names (see children.json.example).\nchildren_file = \"children.json\"\n"},
	{"message_name", "# ProPresenter message template name (must match the message name in ProPresenter).\nmessage_name = \"Eltern rufen\"\n"},
	{"auto_clear_seconds", "# Seconds after which a displayed message is automatically cleared.\n# Set to 0 to disable auto-clear.\nauto_clear_seconds = 30\n"},
	{"activity_log", "# Path to activity log file (JSONL format, append-only).\n# Records send/clear events with timestamps. Leave empty to disable.\n# activity_log = \"activity.jsonl\"\n"},
	{"auth_token", "# Bearer token for API authentication.\n# If not set, a random token is generated on each startup (printed in QR code).\n# Set this for a stable token that survives restarts.\n# auth_token = \"\"\n"},
}

// generateDefaultConfig builds the full default config file content from allConfigBlocks.
func generateDefaultConfig() string {
	var b strings.Builder
	b.WriteString("# Calling Parents — Configuration\n\n")
	for _, block := range allConfigBlocks {
		b.WriteString(block.text)
		b.WriteString("\n")
	}
	return b.String()
}

// LoadResult contains the outcome of loading configuration.
type LoadResult struct {
	// Created is true when a new default config file was written.
	Created bool
	// MergedKeys lists config keys that were appended to an existing file.
	MergedKeys []string
	// BackupPath is the path of the backup file, if one was created.
	BackupPath string
}

// Config holds the application configuration.
type Config struct {
	// ProPresenterHost is the hostname or IP of the ProPresenter machine.
	ProPresenterHost string `toml:"propresenter_host"`
	// ProPresenterPort is the API port of ProPresenter (default 50001).
	ProPresenterPort string `toml:"propresenter_port"`
	// ListenAddr is the address the server listens on (default :8080).
	ListenAddr string `toml:"listen_addr"`
	// ChildrenFile is the path to the JSON file containing children's names.
	ChildrenFile string `toml:"children_file"`
	// AuthToken is the bearer token for API authentication.
	// If empty, a random token is generated on startup.
	AuthToken string `toml:"auth_token"`
	// MessageName is the name of the ProPresenter message template to trigger.
	MessageName string `toml:"message_name"`
	// AutoClearSeconds is the number of seconds after which a sent message is
	// automatically cleared. 0 disables auto-clear.
	AutoClearSeconds int `toml:"auto_clear_seconds"`
	// ActivityLog is the path to the activity log JSONL file.
	// If empty, activity logging is disabled.
	ActivityLog string `toml:"activity_log"`
}

// Load reads configuration from a TOML file, then applies environment variable
// overrides. If the file does not exist, a default config file is created.
// If the file exists but is missing new keys, the file is backed up and the
// missing keys are appended with their default values.
func Load(path string) (Config, LoadResult, error) {
	cfg := defaults()
	result := LoadResult{}

	if path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Write a default config file so the user has something to edit.
			if writeErr := os.WriteFile(path, []byte(generateDefaultConfig()), 0644); writeErr != nil {
				return Config{}, result, fmt.Errorf("creating default config %s: %w", path, writeErr)
			}
			result.Created = true
		} else if err == nil {
			meta, decErr := toml.DecodeFile(path, &cfg)
			if decErr != nil {
				return Config{}, result, fmt.Errorf("reading config %s: %w", path, decErr)
			}
			// Check for missing keys and merge them.
			merged, backupPath, mergeErr := mergeNewKeys(path, meta)
			if mergeErr != nil {
				return Config{}, result, fmt.Errorf("merging config %s: %w", path, mergeErr)
			}
			result.MergedKeys = merged
			result.BackupPath = backupPath
		}
	}

	// Environment variables override TOML values.
	applyEnvOverrides(&cfg)

	return cfg, result, nil
}

// mergeNewKeys checks for config keys that are not present in the user's file.
// If any are found, it backs up the file and appends the missing blocks.
// Keys that are commented out in the default template (activity_log, auth_token)
// are detected by scanning the raw file content for both active and commented forms.
func mergeNewKeys(path string, meta toml.MetaData) ([]string, string, error) {
	// Build set of keys present in the decoded TOML.
	defined := make(map[string]bool)
	for _, key := range meta.Keys() {
		defined[key.String()] = true
	}

	// Also scan the raw file for commented-out keys (e.g. "# auth_token =").
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("reading file: %w", err)
	}
	content := string(raw)
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Check if a commented line contains a known key assignment.
			trimmed = strings.TrimLeft(trimmed, "# ")
			for _, block := range allConfigBlocks {
				if strings.HasPrefix(trimmed, block.key+" ") || strings.HasPrefix(trimmed, block.key+"=") {
					defined[block.key] = true
				}
			}
		}
	}

	// Find missing keys.
	var missing []configBlock
	for _, block := range allConfigBlocks {
		if !defined[block.key] {
			missing = append(missing, block)
		}
	}
	if len(missing) == 0 {
		return nil, "", nil
	}

	// Back up the existing file.
	backupPath := path + ".bak"
	if err := os.WriteFile(backupPath, raw, 0644); err != nil {
		return nil, "", fmt.Errorf("creating backup %s: %w", backupPath, err)
	}

	// Append missing blocks.
	var appendText strings.Builder
	appendText.WriteString("\n# --- New options (added automatically) ---\n\n")
	var merged []string
	for _, block := range missing {
		appendText.WriteString(block.text)
		appendText.WriteString("\n")
		merged = append(merged, block.key)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, "", fmt.Errorf("opening config for append: %w", err)
	}
	defer f.Close()
	if _, err := f.WriteString(appendText.String()); err != nil {
		return nil, "", fmt.Errorf("appending new keys: %w", err)
	}

	return merged, backupPath, nil
}

// defaults returns a Config with sensible default values.
func defaults() Config {
	return Config{
		ProPresenterHost: "localhost",
		ProPresenterPort: "50001",
		ListenAddr:       ":8080",
		ChildrenFile:     "children.json",
		MessageName:      "Eltern rufen",
		AutoClearSeconds: 30,
	}
}

// applyEnvOverrides sets config fields from environment variables if present.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("PROPRESENTER_HOST"); v != "" {
		cfg.ProPresenterHost = v
	}
	if v := os.Getenv("PROPRESENTER_PORT"); v != "" {
		cfg.ProPresenterPort = v
	}
	if v := os.Getenv("LISTEN_ADDR"); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("CHILDREN_FILE"); v != "" {
		cfg.ChildrenFile = v
	}
	if v := os.Getenv("AUTH_TOKEN"); v != "" {
		cfg.AuthToken = v
	}
	if v := os.Getenv("MESSAGE_NAME"); v != "" {
		cfg.MessageName = v
	}
	if v := os.Getenv("AUTO_CLEAR_SECONDS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.AutoClearSeconds = i
		}
	}
	if v := os.Getenv("ACTIVITY_LOG"); v != "" {
		cfg.ActivityLog = v
	}
}

// ProPresenterURL returns the base URL for the ProPresenter API.
func (c Config) ProPresenterURL() string {
	return "http://" + c.ProPresenterHost + ":" + c.ProPresenterPort
}
