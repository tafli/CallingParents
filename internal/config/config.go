package config

import "os"

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
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	return Config{
		ProPresenterHost: getEnv("PROPRESENTER_HOST", "localhost"),
		ProPresenterPort: getEnv("PROPRESENTER_PORT", "50001"),
		ListenAddr:       getEnv("LISTEN_ADDR", ":8080"),
		ChildrenFile:     getEnv("CHILDREN_FILE", "children.json"),
		AuthToken:        os.Getenv("AUTH_TOKEN"),
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
