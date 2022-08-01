package plog

import (
	"encoding/json"
	"testing"
)

// TestLogger is a logger that sends the log entries to a test logger.
func TestLogger(t *testing.T) Logger {
	return func(e Entry) {
		t.Helper()

		j, _ := json.MarshalIndent(e, "", "  ")
		t.Logf("%s", j)
	}
}
