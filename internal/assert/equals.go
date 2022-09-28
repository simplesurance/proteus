package assert

import (
	"testing"
)

// EqualNow asserts that a == b, terminating test immediately otherwise.
func EqualNow[T comparable](t testing.TB, a, b T) {
	t.Helper()

	if !Equal(t, a, b) {
		t.FailNow()
	}
}

// Equal asserts that a == b. Returns true if assertion passes.
func Equal[T comparable](t testing.TB, a, b T) bool {
	t.Helper()

	if a == b {
		return true
	}

	t.Errorf("Values should be equal, but they aren't:\nA: %v\nB: %v", a, b)
	return false
}
