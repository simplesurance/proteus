package assert

import (
	"testing"

	"golang.org/x/exp/constraints"
)

// SmallerNow asserts that a < b, terminating test immediately otherwise.
func SmallerNow[T constraints.Ordered](t testing.TB, a, b T) {
	t.Helper()

	if !Smaller(t, a, b) {
		t.FailNow()
	}
}

// Smaller asserts that a < b. Returns true if assertion passes.
func Smaller[T constraints.Ordered](t testing.TB, a, b T) bool {
	t.Helper()

	if a < b {
		return true
	}

	t.Fatalf("A should be smaller than B:\n%v\n%v", a, b)
	return false
}
