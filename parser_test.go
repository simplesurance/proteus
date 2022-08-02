//go:build unittest || !integrationtest
// +build unittest !integrationtest

package proteus_test

import (
	"math"
	"net/url"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/sources/cfgtest"
	"github.com/simplesurance/proteus/types"
	"github.com/simplesurance/proteus/xtypes"
)

func TestDefaultValueAllTypes(t *testing.T) {
	testWriter := testWriter{t}

	testSource := cfgtest.New(t, types.ParamValues{})

	localhost, _ := url.Parse("https://localhost")

	cfg := struct {
		Str        string                `param:",optional"`
		I          int                   `param:",optional"`
		I8         int8                  `param:",optional"`
		I16        int16                 `param:",optional"`
		I32        int32                 `param:",optional"`
		I64        int64                 `param:",optional"`
		UI8        uint8                 `param:",optional"`
		UI16       uint16                `param:",optional"`
		UI32       uint32                `param:",optional"`
		UI64       uint64                `param:",optional"`
		Bool       bool                  `param:",optional"`
		DynStr     *xtypes.String        `param:",optional"`
		DynBool    *xtypes.Bool          `param:",optional"`
		DynOneOf   *xtypes.OneOf         `param:",optional"`
		DynURL     *xtypes.URL           `param:",optional"`
		DynRSAPriv *xtypes.RSAPrivateKey `param:",optional"`
	}{
		Str:  "str",
		I:    math.MinInt,
		I8:   math.MinInt8,
		I16:  math.MinInt16,
		I32:  math.MinInt32,
		I64:  math.MinInt64,
		UI8:  math.MaxUint8,
		UI16: math.MaxUint16,
		UI32: math.MaxUint32,
		UI64: math.MaxUint64,
		Bool: true,
		DynStr: &xtypes.String{
			DefaultValue: "def dyn",
		},
		DynBool: &xtypes.Bool{
			DefaultValue: true,
		},
		DynOneOf: &xtypes.OneOf{
			DefaultValue: "sol",
			Choices:      []string{"do", "re", "mi", "fa", "sol", "la", "si"},
		},
		DynURL: &xtypes.URL{
			DefaultValue: localhost,
		},
	}

	parsed, err := proteus.MustParse(&cfg,
		proteus.WithSources(testSource),
		proteus.WithLogger(newTestLogger(t)))
	if err != nil {
		t.Logf("Unexpected error parsing configuration: %+v", err)
		parsed.ErrUsage(testWriter, err)
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
	assert.Equal(t, cfg.UI16, uint16(math.MaxUint16))
	assert.Equal(t, cfg.UI32, uint32(math.MaxUint32))
	assert.Equal(t, cfg.UI64, uint64(math.MaxUint64))
	assert.Equal(t, true, cfg.Bool)
	assert.Equal(t, "def dyn", cfg.DynStr.Value())
	assert.Equal(t, "sol", cfg.DynOneOf.Value())
	assert.Equal(t, true, cfg.DynBool.Value())
	assert.Equal(t, localhost, cfg.DynURL.Value())
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

	testSource := cfgtest.New(t, types.ParamValues{
		"": map[string]string{
			"devmode":  "true",
			"embparam": "emb1",
		},
		"testset": map[string]string{
			"fsvalue":  "test value",
			"embparam": "emb2",
		},
	})

	defer testSource.Stop()

	parsed, err := proteus.MustParse(&testAppCfg,
		proteus.WithSources(testSource),
		proteus.WithLogger(newTestLogger(t)))
	if err != nil {
		t.Logf("Unexpected error parsing configuration: %+v", err)
		parsed.ErrUsage(testWriter, err)
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

	testSource := cfgtest.New(t, types.ParamValues{
		"http": map[string]string{
			"ip":   "127.0.0.1",
			"port": "42",
		},
		"log": map[string]string{
			"filename": "/dev/null",
		},
	})

	defer testSource.Stop()

	parsed, err := proteus.MustParse(&testAppCfg,
		proteus.WithSources(testSource),
		proteus.WithLogger(newTestLogger(t)))
	if err != nil {
		t.Logf("Unexpected error parsing configuration: %+v", err)
		parsed.ErrUsage(testWriter, err)
		t.FailNow()
	}

	assert.Equal(t, "127.0.0.1", testAppCfg.HTTP.IP)
	assert.EqualValues(t, 42, testAppCfg.HTTP.Port)
	assert.Equal(t, "/dev/null", testAppCfg.Log.FileName)
}

func newTestLogger(t *testing.T) proteus.Logger {
	return func(msg string, skip int) {
		_, file, line, _ := runtime.Caller(skip)
		t.Logf("%s (%s:%d)", msg, file, line)
	}
}

type testWriter struct {
	t *testing.T
}

func (t testWriter) Write(v []byte) (int, error) {
	t.t.Logf("%s", v)
	return len(v), nil
}
