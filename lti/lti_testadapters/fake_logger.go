package lti_testadapters

import (
	"fmt"
	"strings"
	"sync"

	"github.com/kvizdos/lti-server/lti/lti_ports"
)

var _ lti_ports.Logger = (*FakeLogger)(nil)

// FakeLogger is a thread-safe test logger that records every message
// and the key/value pairs passed to it for later assertions.
type FakeLogger struct {
	mu      sync.Mutex
	entries []LogEntry
}

// LogEntry represents one structured log call.
type LogEntry struct {
	Level string
	Msg   string
	KVs   []any
}

// NewFakeLogger returns a fresh FakeLogger.
func NewFakeLogger() *FakeLogger {
	return &FakeLogger{}
}

func (f *FakeLogger) Info(msg string, kv ...any)  { f.add("INFO", msg, kv...) }
func (f *FakeLogger) Warn(msg string, kv ...any)  { f.add("WARN", msg, kv...) }
func (f *FakeLogger) Debug(msg string, kv ...any) { f.add("DEBUG", msg, kv...) }
func (f *FakeLogger) Error(msg string, kv ...any) { f.add("ERROR", msg, kv...) }

func (f *FakeLogger) add(level, msg string, kv ...any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries = append(f.entries, LogEntry{
		Level: level,
		Msg:   msg,
		KVs:   append([]any(nil), kv...), // copy
	})
}

// Entries returns a snapshot of all logged entries.
func (f *FakeLogger) Entries() []LogEntry {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]LogEntry, len(f.entries))
	copy(out, f.entries)
	return out
}

// Count returns how many total log calls were recorded.
func (f *FakeLogger) Count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.entries)
}

// Last returns the last logged entry (or nil if none).
func (f *FakeLogger) Last() *LogEntry {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.entries) == 0 {
		return nil
	}
	e := f.entries[len(f.entries)-1]
	return &e
}

// ContainsMessage checks if any log message contains the given substring.
func (f *FakeLogger) ContainsMessage(substr string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, e := range f.entries {
		if strings.Contains(e.Msg, substr) {
			return true
		}
	}
	return false
}

// Dump prints all logs to stdout — handy for debugging failed tests.
func (f *FakeLogger) Dump() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, e := range f.entries {
		fmt.Printf("[%s] %s %v\n", e.Level, e.Msg, e.KVs)
	}
}
