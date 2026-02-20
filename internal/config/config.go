package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
)

// defaultConfigContent is written to disk when no config file exists.
const defaultConfigContent = `# Calling Parents — Configuration

# Hostname or IP of the ProPresenter machine.
propresenter_host = "localhost"

# ProPresenter API port (default in ProPresenter: 50001).
propresenter_port = "50001"

# Address and port this server listens on.
listen_addr = ":8080"

# Path to the JSON file with children's names (see children.json.example).
children_file = "children.json"

# ProPresenter message template name (must match the message name in ProPresenter).
message_name = "Eltern rufen"

# Seconds after which a displayed message is automatically cleared.
# Set to 0 to disable auto-clear.
auto_clear_seconds = 30

# Path to activity log file (JSONL format, append-only).
# Records send/clear events with timestamps. Leave empty to disable.
# activity_log = "activity.jsonl"

# Bearer token for API authentication.
# If not set, a random token is generated on each startup (printed in QR code).
# Set this for a stable token that survives restarts.
# auth_token = ""
`

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
// The returned bool indicates whether a new config file was created.
func Load(path string) (Config, bool, error) {
	cfg := defaults()
	created := false

	if path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Write a default config file so the user has something to edit.
			if writeErr := os.WriteFile(path, []byte(defaultConfigContent), 0644); writeErr != nil {
				return Config{}, false, fmt.Errorf("creating default config %s: %w", path, writeErr)
			}
			created = true
		} else if err == nil {
			if _, err := toml.DecodeFile(path, &cfg); err != nil {
				return Config{}, false, fmt.Errorf("reading config %s: %w", path, err)
			}
		}
	}

	// Environment variables override TOML values.
	applyEnvOverrides(&cfg)

	return cfg, created, nil
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
