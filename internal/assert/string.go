package assert

import (
	"strings"
	"testing"
)

// StringContains asserts that substr is within s. Returns
// true if assertion passes.
func StringContains(t testing.TB, s, substr string) bool {
	t.Helper()

	if strings.Contains(s, substr) {
		return true
	}

	t.Errorf("'substr' is not contained in 's':\ns: %s\nsubstr: %s", s, substr)
	return false
}
