package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
)

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
// overrides. If the file does not exist, defaults are used.
func Load(path string) (Config, error) {
	cfg := defaults()

	if path != "" {
		if _, err := os.Stat(path); err == nil {
			if _, err := toml.DecodeFile(path, &cfg); err != nil {
				return Config{}, fmt.Errorf("reading config %s: %w", path, err)
			}
		}
	}

	// Environment variables override TOML values.
	applyEnvOverrides(&cfg)

	return cfg, nil
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
