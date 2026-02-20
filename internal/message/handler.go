package message

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tafli/CallingParents/internal/activitylog"
)

// Handler provides HTTP endpoints that proxy message operations to ProPresenter.
// The PWA sends only a child's name; the handler knows the message template.
type Handler struct {
	proPresenterURL  string
	messageName      string
	autoClearSeconds int
	client           *http.Client
	logger           *activitylog.Logger
}

// New creates a Handler that talks to ProPresenter at the given base URL
// using the given message template name.
func New(proPresenterURL, messageName string, autoClearSeconds int, logger *activitylog.Logger) *Handler {
	return &Handler{
		proPresenterURL:  strings.TrimRight(proPresenterURL, "/"),
		messageName:      messageName,
		autoClearSeconds: autoClearSeconds,
		client:           &http.Client{Timeout: 10 * time.Second},
		logger:           logger,
	}
}

// sendRequest is the expected JSON body for POST /message/send.
type sendRequest struct {
	Name string `json:"name"`
}

// HandleSend triggers the ProPresenter message with the given child's name.
func (h *Handler) HandleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		http.Error(w, "name must not be empty", http.StatusBadRequest)
		return
	}

	msgID := url.PathEscape(h.messageName)
	ppURL := fmt.Sprintf("%s/v1/message/%s/trigger", h.proPresenterURL, msgID)

	body := fmt.Sprintf(`[{"name":"Name","text":{"text":"%s"}}]`, escapeJSON(name))

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	ppReq, err := http.NewRequestWithContext(ctx, http.MethodPost, ppURL, strings.NewReader(body))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	ppReq.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(ppReq)
	if err != nil {
		http.Error(w, "ProPresenter ist nicht erreichbar", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		http.Error(w, "ProPresenter hat die Nachricht abgelehnt", http.StatusServiceUnavailable)
		return
	}

	h.logger.Log("send", name)
	w.WriteHeader(http.StatusNoContent)
}

// HandleClear clears the ProPresenter message.
func (h *Handler) HandleClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	msgID := url.PathEscape(h.messageName)
	ppURL := fmt.Sprintf("%s/v1/message/%s/clear", h.proPresenterURL, msgID)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	ppReq, err := http.NewRequestWithContext(ctx, http.MethodGet, ppURL, nil)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp, err := h.client.Do(ppReq)
	if err != nil {
		http.Error(w, "ProPresenter ist nicht erreichbar", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		http.Error(w, "ProPresenter konnte die Nachricht nicht löschen", http.StatusServiceUnavailable)
		return
	}

	h.logger.Log("clear", "")
	w.WriteHeader(http.StatusNoContent)
}

// HandleTest tests the connection to ProPresenter by listing messages.
func (h *Handler) HandleTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ppURL := fmt.Sprintf("%s/v1/messages", h.proPresenterURL)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	ppReq, err := http.NewRequestWithContext(ctx, http.MethodGet, ppURL, nil)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp, err := h.client.Do(ppReq)
	if err != nil {
		http.Error(w, "ProPresenter ist nicht erreichbar", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		http.Error(w, "ProPresenter-Verbindung fehlgeschlagen", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	// Forward the ProPresenter response body.
	var messages json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		json.NewEncoder(w).Encode([]any{})
		return
	}
	json.NewEncoder(w).Encode(messages)
}

// escapeJSON escapes a string for safe embedding in a JSON string literal.
func escapeJSON(s string) string {
	b, _ := json.Marshal(s)
	// json.Marshal wraps in quotes: "value" — strip them.
	return string(b[1 : len(b)-1])
}

// configResponse is the JSON body returned by HandleConfig.
type configResponse struct {
	AutoClearSeconds int `json:"autoClearSeconds"`
}

// HandleConfig returns client-relevant configuration as JSON.
func (h *Handler) HandleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configResponse{
		AutoClearSeconds: h.autoClearSeconds,
	})
}
