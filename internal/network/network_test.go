package network

import (
	"strings"
	"testing"
)

func TestLanURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		listenAddr string
		wantPrefix string
		wantSuffix string
	}{
		{
			name:       "default port",
			listenAddr: ":8080",
			wantPrefix: "http://",
			wantSuffix: ":8080",
		},
		{
			name:       "custom port",
			listenAddr: ":3000",
			wantPrefix: "http://",
			wantSuffix: ":3000",
		},
		{
			name:       "explicit bind address",
			listenAddr: "0.0.0.0:9090",
			wantPrefix: "http://",
			wantSuffix: ":9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			url := LanURL(tt.listenAddr)

			if !strings.HasPrefix(url, tt.wantPrefix) {
				t.Errorf("LanURL(%q) = %q, want prefix %q", tt.listenAddr, url, tt.wantPrefix)
			}
			if !strings.HasSuffix(url, tt.wantSuffix) {
				t.Errorf("LanURL(%q) = %q, want suffix %q", tt.listenAddr, url, tt.wantSuffix)
			}
		})
	}
}

func TestLanURLReturnsValidURL(t *testing.T) {
	t.Parallel()

	url := LanURL(":8080")

	if !strings.HasPrefix(url, "http://") {
		t.Fatalf("expected http:// prefix, got %q", url)
	}

	host := strings.TrimPrefix(url, "http://")
	host = strings.Split(host, ":")[0]

	if host == "" {
		t.Fatal("host part is empty")
	}
}
