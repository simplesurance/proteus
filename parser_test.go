//go:build unittest || !integrationtest
// +build unittest !integrationtest

package proteus_test

import (
	"math"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/internal/assert"
	"github.com/simplesurance/proteus/plog"
	"github.com/simplesurance/proteus/sources/cfgtest"
	"github.com/simplesurance/proteus/types"
	"github.com/simplesurance/proteus/xtypes"
)

func TestDefaultValueAllTypes(t *testing.T) {
	testWriter := testWriter{t}
	now := time.Now()

	testProvider := cfgtest.New(types.ParamValues{})

	localhost, _ := url.Parse("https://localhost")

	cfg := struct {
		Str      string                `param:",optional"`
		I        int                   `param:",optional"`
		I8       int8                  `param:",optional"`
		I16      int16                 `param:",optional"`
		I32      int32                 `param:",optional"`
		I64      int64                 `param:",optional"`
		U        uint                  `param:",optional"`
		UI8      uint8                 `param:",optional"`
		UI16     uint16                `param:",optional"`
		UI32     uint32                `param:",optional"`
		UI64     uint64                `param:",optional"`
		Bool     bool                  `param:",optional"`
		Time     time.Time             `param:",optional"`
		Duration time.Duration         `param:",optional"`
		XStr     *xtypes.String        `param:",optional"`
		XBool    *xtypes.Bool          `param:",optional"`
		XOneOf   *xtypes.OneOf         `param:",optional"`
		XURL     *xtypes.URL           `param:",optional"`
		XRSAPriv *xtypes.RSAPrivateKey `param:",optional"`
	}{
		Str:      "str",
		I:        math.MinInt,
		I8:       math.MinInt8,
		I16:      math.MinInt16,
		I32:      math.MinInt32,
		I64:      math.MinInt64,
		U:        math.MaxUint,
		UI8:      math.MaxUint8,
		UI16:     math.MaxUint16,
		UI32:     math.MaxUint32,
		UI64:     math.MaxUint64,
		Bool:     true,
		Time:     now,
		Duration: time.Hour,
		XStr: &xtypes.String{
			DefaultValue: "def dyn",
		},
		XBool: &xtypes.Bool{
			DefaultValue: true,
		},
		XOneOf: &xtypes.OneOf{
			DefaultValue: "sol",
			Choices:      []string{"do", "re", "mi", "fa", "sol", "la", "si"},
		},
		XURL: &xtypes.URL{
			DefaultValue: localhost,
		},
	}

	parsed, err := proteus.MustParse(&cfg,
		proteus.WithProviders(testProvider),
		proteus.WithLogger(plog.TestLogger(t)))
	if err != nil {
		t.Logf("Unexpected error parsing configuration: %+v", err)
		parsed.WriteError(testWriter, err)
		t.FailNow()
	}

	parsed.Usage(testWriter)

	assert.Equal(t, cfg.Str, "str")
	assert.Equal(t, cfg.I, int(math.MinInt))
	assert.Equal(t, cfg.I8, int8(math.MinInt8))
	assert.Equal(t, cfg.I16, int16(math.MinInt16))
	assert.Equal(t, cfg.I32, int32(math.MinInt32))
	assert.Equal(t, cfg.I64, int64(math.MinInt64))
	assert.Equal(t, cfg.UI8, uint8(math.MaxUint8))
	assert.Equal(t, cfg.U, uint(math.MaxUint))
	assert.Equal(t, cfg.UI16, uint16(math.MaxUint16))
	assert.Equal(t, cfg.UI32, uint32(math.MaxUint32))
	assert.Equal(t, cfg.UI64, uint64(math.MaxUint64))
	assert.Equal(t, true, cfg.Bool)
	assert.Equal(t, "def dyn", cfg.XStr.Value())
	assert.Equal(t, "sol", cfg.XOneOf.Value())
	assert.Equal(t, true, cfg.XBool.Value())
	assert.Equal(t, localhost, cfg.XURL.Value())
}

// TestEmbeddingParameters asserts that embedding structs result in the values
// being flat.
func TestEmbeddingParameters(t *testing.T) {
	testWriter := testWriter{t}

	type embeddedConfig struct {
		EmbParam string
	}

	testAppCfg := struct {
		embeddedConfig
		TestSet struct {
			embeddedConfig
			FSValue string
		}
		DevMode bool
	}{}

	testProvider := cfgtest.New(types.ParamValues{
		"": map[string]string{
			"devmode":  "true",
			"embparam": "emb1",
		},
		"testset": map[string]string{
			"fsvalue":  "test value",
			"embparam": "emb2",
		},
	})

	defer testProvider.Stop()

	parsed, err := proteus.MustParse(&testAppCfg,
		proteus.WithProviders(testProvider),
		proteus.WithLogger(plog.TestLogger(t)))
	if err != nil {
		t.Logf("Unexpected error parsing configuration: %+v", err)
		parsed.WriteError(testWriter, err)
		t.FailNow()
	}

	parsed.Usage(testWriter)

	assert.Equal(t, true, testAppCfg.DevMode)
	assert.Equal(t, "emb1", testAppCfg.EmbParam)
	assert.Equal(t, "test value", testAppCfg.TestSet.FSValue)
	assert.Equal(t, "emb2", testAppCfg.TestSet.EmbParam)
}

// TestEmbeddingParamSet asserts that structs including paramsets can be
// embedded as parameter.
func TestEmbeddingParamSet(t *testing.T) {
	testWriter := testWriter{t}

	type httpConfig struct {
		HTTP struct {
			IP   string
			Port uint16
		}
	}

	type logConfig struct {
		Log struct {
			FileName string
		}
	}

	testAppCfg := struct {
		httpConfig
		logConfig
	}{}

	testProvider := cfgtest.New(types.ParamValues{
		"http": map[string]string{
			"ip":   "127.0.0.1",
			"port": "42",
		},
		"log": map[string]string{
			"filename": "/dev/null",
		},
	})

	defer testProvider.Stop()

	parsed, err := proteus.MustParse(&testAppCfg,
		proteus.WithProviders(testProvider),
		proteus.WithLogger(plog.TestLogger(t)))
	if err != nil {
		t.Logf("Unexpected error parsing configuration: %+v", err)
		parsed.WriteError(testWriter, err)
		t.FailNow()
	}

	assert.Equal(t, "127.0.0.1", testAppCfg.HTTP.IP)
	assert.Equal(t, 42, testAppCfg.HTTP.Port)
	assert.Equal(t, "/dev/null", testAppCfg.Log.FileName)
}

func TestParseWithTrim(t *testing.T) {
	testWriter := testWriter{t}

	params := struct {
		TestString string
		TestInt    int
		TestURL    *xtypes.URL
	}{}

	testProvider := cfgtest.New(types.ParamValues{
		"": map[string]string{
			"teststring": `
			value
			`,

			"testint": `

			42
			`,

			"testurl": `
			https://localhost
			`,
		},
	})

	defer testProvider.Stop()

	parsed, err := proteus.MustParse(&params,
		proteus.WithProviders(testProvider),
		proteus.WithLogger(plog.TestLogger(t)),
		proteus.WithValueFormatting(proteus.ValueFormattingOptions{
			TrimSpace: true,
		}))
	if err != nil {
		t.Logf("Unexpected error parsing configuration: %+v", err)
		parsed.WriteError(testWriter, err)
		t.FailNow()
	}

	assert.Equal(t, "value", params.TestString)
	assert.Equal(t, 42, params.TestInt)
	assert.Equal(t, "https://localhost", params.TestURL.Value().String())
}

func TestTimeAndDuration(t *testing.T) {
	testWriter := testWriter{t}
	now := time.Now()
	dur := time.Hour + time.Duration(1)

	params := struct {
		T time.Time
		D time.Duration
	}{}

	testProvider := cfgtest.New(types.ParamValues{"": map[string]string{
		"t": now.Format(time.RFC3339Nano),
		"d": dur.String(),
	}})

	defer testProvider.Stop()

	parsed, err := proteus.MustParse(&params,
		proteus.WithProviders(testProvider),
		proteus.WithLogger(plog.TestLogger(t)))
	if err != nil {
		t.Logf("Unexpected error parsing configuration: %+v", err)
		parsed.WriteError(testWriter, err)
		t.FailNow()
	}

	assert.EqualFn(t, now, params.T)
	assert.Equal(t, dur, params.D)

	sb := strings.Builder{}
	parsed.Dump(&sb)
	t.Log(sb.String())
}

func TestIgnoreInvalidMember(t *testing.T) {
	testWriter := testWriter{t}

	params := struct {
		A int    `param:",optional"`
		B func() `param:"-"`
	}{}

	testProvider := cfgtest.New(types.ParamValues{"": map[string]string{}})

	defer testProvider.Stop()

	parsed, err := proteus.MustParse(&params,
		proteus.WithProviders(testProvider),
		proteus.WithLogger(plog.TestLogger(t)))
	if err != nil {
		t.Logf("Unexpected error parsing configuration: %+v", err)
		parsed.WriteError(testWriter, err)
		t.FailNow()
	}

	sb := strings.Builder{}
	parsed.Dump(&sb)
	t.Log(sb.String())

	t.Logf("Configuration field successfully ignored")
}

type testWriter struct {
	t *testing.T
}

func (t testWriter) Write(v []byte) (int, error) {
	t.t.Logf("%s", v)
	return len(v), nil
}
