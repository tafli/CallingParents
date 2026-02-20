package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration.
type Config struct {
	// ProPresenterHost is the hostname or IP of the ProPresenter machine.
	ProPresenterHost string
	// ProPresenterPort is the API port of ProPresenter (default 50001).
	ProPresenterPort string
	// ListenAddr is the address the server listens on (default :8080).
	ListenAddr string
	// ChildrenFile is the path to the JSON file containing children's names.
	ChildrenFile string
	// AuthToken is the bearer token for API authentication.
	// If empty, a random token is generated on startup.
	AuthToken string
	// MessageName is the name of the ProPresenter message template to trigger.
	MessageName string
	// AutoClearSeconds is the number of seconds after which a sent message is
	// automatically cleared. 0 disables auto-clear.
	AutoClearSeconds int
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	return Config{
		ProPresenterHost: getEnv("PROPRESENTER_HOST", "localhost"),
		ProPresenterPort: getEnv("PROPRESENTER_PORT", "50001"),
		ListenAddr:       getEnv("LISTEN_ADDR", ":8080"),
		ChildrenFile:     getEnv("CHILDREN_FILE", "children.json"),
		AuthToken:        os.Getenv("AUTH_TOKEN"),
		MessageName:      getEnv("MESSAGE_NAME", "Eltern rufen"),
		AutoClearSeconds: getEnvInt("AUTO_CLEAR_SECONDS", 30),
	}
}

// ProPresenterURL returns the base URL for the ProPresenter API.
func (c Config) ProPresenterURL() string {
	return "http://" + c.ProPresenterHost + ":" + c.ProPresenterPort
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}
