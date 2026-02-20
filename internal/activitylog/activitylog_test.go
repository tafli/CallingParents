package activitylog

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestLogWritesJSONL(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "test.jsonl")
	logger, err := New(path)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.Log("send", "Paul")
	logger.Log("clear", "")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), string(data))
	}

	var entry1 Entry
	if err := json.Unmarshal([]byte(lines[0]), &entry1); err != nil {
		t.Fatalf("failed to parse line 1: %v", err)
	}
	if entry1.Action != "send" {
		t.Errorf("expected action=send, got %q", entry1.Action)
	}
	if entry1.Name != "Paul" {
		t.Errorf("expected name=Paul, got %q", entry1.Name)
	}
	if entry1.Time == "" {
		t.Error("expected non-empty time")
	}

	var entry2 Entry
	if err := json.Unmarshal([]byte(lines[1]), &entry2); err != nil {
		t.Fatalf("failed to parse line 2: %v", err)
	}
	if entry2.Action != "clear" {
		t.Errorf("expected action=clear, got %q", entry2.Action)
	}
	if entry2.Name != "" {
		t.Errorf("expected empty name for clear, got %q", entry2.Name)
	}
}

func TestLogNilLoggerIsSafe(t *testing.T) {
	t.Parallel()

	var logger *Logger
	// Should not panic
	logger.Log("send", "Paul")
	if err := logger.Close(); err != nil {
		t.Errorf("expected nil error from nil logger, got %v", err)
	}
}

func TestLogConcurrentWrites(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "concurrent.jsonl")
	logger, err := New(path)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Log("send", "Child")
		}()
	}
	wg.Wait()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open log file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		var e Entry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			t.Fatalf("invalid JSON on line %d: %v", count+1, err)
		}
		count++
	}
	if count != 50 {
		t.Errorf("expected 50 entries, got %d", count)
	}
}

func TestNewCreatesFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "new.jsonl")

	logger, err := New(path)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected log file to be created")
	}
}

func TestLogAppendsToExistingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "append.jsonl")

	// Write a first entry
	logger1, err := New(path)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	logger1.Log("send", "Anna")
	logger1.Close()

	// Re-open and write another
	logger2, err := New(path)
	if err != nil {
		t.Fatalf("failed to re-open logger: %v", err)
	}
	logger2.Log("send", "Ben")
	logger2.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines after append, got %d", len(lines))
	}
}
