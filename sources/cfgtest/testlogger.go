package cfgtest

import (
	"runtime"
	"testing"
)

func LoggerFor(t *testing.T) func(msg string, depth int) {
	return func(msg string, depth int) {
		t.Helper()

		_, file, line, _ := runtime.Caller(depth)
		t.Logf("\n%s:%d %s", file, line, msg)
	}
}
