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

// EqualNowFn asserts that a.Equals(b).
func EqualNowFn[T any](t testing.TB, a, b T) {
	t.Helper()

	if !EqualFn[T](t, a, b) {
		t.FailNow()
	}
}

// EqualFn asserts that "a" is equals to "b" by calling a.Equal(b).
// Returns true if assertion passes.
func EqualFn[T any](t testing.TB, a any, b T) bool {
	t.Helper()

	eqIf := a.(interface{ Equal(other T) bool })

	if eqIf.Equal(b) {
		return true
	}

	t.Errorf("Values should be equal, but they aren't:\nA: %v\nB: %v", a, b)
	return false
}
