package assert

import (
	"testing"
)

// NotNilNow asserts that the value is not nil, terminating the test if it
// isn't.
func NotNilNow(t testing.TB, v any) {
	t.Helper()

	if !NotNil(t, v) {
		t.FailNow()
	}
}

// NotNil asserts that the value is not nil. Returns true if assertion passes.
func NotNil(t testing.TB, v any) bool {
	t.Helper()

	if v != nil {
		return true
	}

	t.Error("Value is nil")
	return false
}
