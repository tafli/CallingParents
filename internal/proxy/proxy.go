package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// New creates a reverse proxy handler that forwards requests from /api/*
// to the ProPresenter API, stripping the /api prefix.
func New(proPresenterURL string) http.Handler {
	target, err := url.Parse(proPresenterURL)
	if err != nil {
		log.Fatalf("invalid ProPresenter URL %q: %v", proPresenterURL, err)
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host

			// Strip the /api prefix: /api/v1/messages -> /v1/messages
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api")
			if req.URL.RawPath != "" {
				req.URL.RawPath = strings.TrimPrefix(req.URL.RawPath, "/api")
			}
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("proxy error: %v", err)
			http.Error(w, "ProPresenter is not reachable", http.StatusBadGateway)
		},
	}

	return proxy
}
