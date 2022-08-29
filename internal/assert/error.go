package assert

import (
	"testing"
)

// NoErrorNow asserts that the provided value is not an error, terminating
// the test if this is not true.
func NoErrorNow(t testing.TB, err error) {
	t.Helper()

	if !NoError(t, err) {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// NoError asserts that the provided value is not an error. Returns true if
// assertion passes.
func NoError(t testing.TB, err error) bool {
	t.Helper()

	if err == nil {
		return true
	}

	t.Errorf("Unexpected error: %v", err)
	return false
}

// ErrorNow asserts that the provided value is an error, terminating
// the test if this is not true.
func ErrorNow(t testing.TB, err error) {
	t.Helper()

	if !Error(t, err) {
		t.FailNow()
	}
}

// Error asserts that the provided value is an error. Returns true if assertion
// passes.
func Error(t testing.TB, err error) bool {
	t.Helper()

	if err != nil {
		return true
	}

	t.Error("Expected error, but got nil")
	return false
}

// PanicsNow asserts that the provided function panics, terminating the test
// immediately if it doesn't.
func PanicsNow(t testing.TB, f func()) {
	t.Helper()

	var r any
	func() {
		defer func() {
			r = recover()
		}()

		f()
	}()

	if r != nil {
		return
	}

	t.Fatalf("Function was expected to panic, but it didn't")
}
