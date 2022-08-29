package plog_test

import (
	"encoding/json"
	"runtime"
	"testing"

	"github.com/simplesurance/proteus/internal/assert"
	"github.com/simplesurance/proteus/plog"
)

func TestMarshaling(t *testing.T) {
	cases := []plog.Entry{
		{
			Severity: plog.SevDebug,
			Message:  "debug",
			Caller:   plog.ReadCaller(0),
		},
		{
			Message: "smallest",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Message, func(t *testing.T) {
			j, err := json.MarshalIndent(tc, "", "  ")
			assert.NoErrorNow(t, err)
			t.Logf("%s %v", j, err)

			var have plog.Entry
			err = json.Unmarshal(j, &have)
			assert.NoErrorNow(t, err)

			assert.Equal(t, have, tc)
		})
	}
}

func TestLog(t *testing.T) {
	var loggedEntry plog.Entry
	var logger plog.Logger = func(e plog.Entry) {
		loggedEntry = e
	}

	// the next two lines must be exactly one after the other
	_, thisFile, thisLineNumber, ok := runtime.Caller(0)
	logger.E("test error message")

	assert.NotNilNow(t, loggedEntry.Caller)
	assert.True(t, ok, "runtime.Caller returned !ok")

	assert.Equal(t, thisFile, loggedEntry.Caller.File)
	assert.Equal(t, thisLineNumber+1, loggedEntry.Caller.LineNumber)
}

func TestSkipCaller(t *testing.T) {
	var loggedEntry plog.Entry
	var logger plog.Logger = func(e plog.Entry) {
		loggedEntry = e
	}

	logf := func(m string) {
		// The log entry should not register the following file/line number;
		// It should register instead its caller.
		logger.E("test error", plog.SkipCallers(1))
	}

	// the next two lines must be exactly one after the other
	_, thisFile, thisLineNumber, ok := runtime.Caller(0)
	logf("test message") // must record this line

	assert.NotNilNow(t, loggedEntry.Caller)
	assert.TrueNow(t, ok, "runtime.Caller() returned !ok")

	assert.Equal(t, thisFile, loggedEntry.Caller.File)
	assert.Equal(t, thisLineNumber+1, loggedEntry.Caller.LineNumber)
}
