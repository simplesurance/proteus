package proteus_test

import (
	"bytes"
	"sync/atomic"
	"testing"
	"time"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/sources/cfgtest"
	"github.com/simplesurance/proteus/types"
	"github.com/simplesurance/proteus/xtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDynamic asserts that the expected behavior is being observed when
// a provider updates a value regarding the result of the Value() function
// and the callback function. The test is written in a specific way to
// maximize the chance of possible race conditions to be detected by
// "go test -race".
func TestDynamic(t *testing.T) {
	wantedValues := []string{
		"initial value",
		"updated value",
		"second updated value",
		"final updated value",
	}
	var callbackInvoked int32

	params := struct {
		X *xtypes.String
	}{
		X: &xtypes.String{
			UpdateFn: func(s string) {
				// Each time this callback function is called we expect
				// that is called with one of the values from the wanted
				// values, in order.
				callIx := atomic.AddInt32(&callbackInvoked, 1) - 1
				t.Logf("callback invoked callIx=%d value=%s", callIx, s)
				require.True(t, int(callIx) < len(wantedValues))
				assert.Equal(t, wantedValues[callIx], s)
			},
		},
	}

	provider := cfgtest.New(types.ParamValues{
		"": map[string]string{"x": wantedValues[0]},
	})

	parsed, err := proteus.MustParse(&params, proteus.WithProviders(provider))
	if err != nil {
		buffer := bytes.Buffer{}
		parsed.ErrUsage(&buffer, err)
		t.Error(buffer.String())
	}

	require.Equal(t, wantedValues[0], params.X.Value())

	time.Sleep(time.Second)

	for ix, value := range wantedValues[1:] {
		t.Logf("Done waiting for callback; requesting dynamic value update with value %q", value)

		// Update the value is assert that the new value is visible. The test code
		// here will update the value in one routine while reading it on a busy
		// loop to allow the race detector to find concurrency issues.
		go func() {
			provider.Update("", "x", value)
		}()

		start := time.Now()
		for i := 0; ; i++ {
			if params.X.Value() == value {
				t.Logf("Got new value with i=%d", i)
				break
			}

			if time.Since(start) > 2*time.Second {
				t.Fatalf("timeout waiting for Value() to return test value ix=%d value=%q", ix+1, value)
			}
		}

		// give time for any spurious invocation of the callback function to happen
		time.Sleep(time.Second)
	}

	// each value the parameter ever had must result in exactly one call to
	// the callback function
	assert.EqualValues(t, len(wantedValues), atomic.LoadInt32(&callbackInvoked))
}
