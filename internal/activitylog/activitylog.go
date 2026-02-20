package activitylog

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Entry represents a single activity log line.
type Entry struct {
	Time   string `json:"time"`
	Action string `json:"action"`
	Name   string `json:"name,omitempty"`
}

// Logger appends activity entries as JSON lines to a file.
// It is safe for concurrent use.
type Logger struct {
	mu   sync.Mutex
	file *os.File
}

// New opens (or creates) the log file for appending and returns a Logger.
func New(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Logger{file: f}, nil
}

// Log writes a timestamped entry to the log file.
func (l *Logger) Log(action, name string) {
	if l == nil {
		return
	}
	e := Entry{
		Time:   time.Now().Format(time.RFC3339),
		Action: action,
		Name:   name,
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	data, err := json.Marshal(e)
	if err != nil {
		return
	}
	data = append(data, '\n')
	l.file.Write(data)
}

// Close closes the underlying file.
func (l *Logger) Close() error {
	if l == nil {
		return nil
	}
	return l.file.Close()
}
