package assert

import "testing"

// TrueNow asserts that the value is true, terminating the test immediately if
// isn't.
func TrueNow(t testing.TB, b bool, msg string) {
	if !True(t, b, msg) {
		t.FailNow()
	}
}

// True asserts that the value is true. Returns true if assertion passes.
func True(t testing.TB, b bool, msg string) bool {
	t.Helper()
	if b {
		return true
	}

	t.Fatalf("Value is not true: %s", msg)
	return false
}
