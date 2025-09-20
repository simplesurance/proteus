//go:build unittest || !integrationtest
// +build unittest !integrationtest

package proteus_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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

func TestRSAPrivateKey(t *testing.T) {
	_, privateKeyStr := generateTestKey(t)
	defaultKey, _ := generateTestKey(t)

	tests := []struct {
		name          string
		params        types.ParamValues
		shouldErr     bool
		optionalIsNil bool
		useDefault    bool
	}{
		{
			name: "valid key for optional and required",
			params: types.ParamValues{
				"": {
					"optionalkey": privateKeyStr,
					"requiredkey": privateKeyStr,
				},
			},
			shouldErr:     false,
			optionalIsNil: false,
		},
		{
			name: "empty string for optional key",
			params: types.ParamValues{
				"": {
					"optionalkey": "",
					"requiredkey": privateKeyStr,
				},
			},
			shouldErr:     false,
			optionalIsNil: true,
		},
		{
			name: "empty string for optional key with default",
			params: types.ParamValues{
				"": {
					"optionalkey": "",
					"requiredkey": privateKeyStr,
				},
			},
			shouldErr:     false,
			optionalIsNil: false,
			useDefault:    true,
		},
		{
			name: "no value for optional key",
			params: types.ParamValues{
				"": {
					"requiredkey": privateKeyStr,
				},
			},
			shouldErr:     false,
			optionalIsNil: true,
		},
		{
			name: "empty string for required key",
			params: types.ParamValues{
				"": {
					"requiredkey": "",
				},
			},
			shouldErr: true,
		},
		{
			name:      "no value for required key",
			params:    types.ParamValues{"": {}},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := struct {
				OptionalKey *xtypes.RSAPrivateKey `param:",optional"`
				RequiredKey *xtypes.RSAPrivateKey
			}{}

			if tt.useDefault {
				cfg.OptionalKey = &xtypes.RSAPrivateKey{DefaultValue: defaultKey}
			}

			testProvider := cfgtest.New(tt.params)
			defer testProvider.Stop()

			_, err := proteus.MustParse(&cfg,
				proteus.WithProviders(testProvider))

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.useDefault {
					assert.Equal(t, defaultKey, cfg.OptionalKey.Value())
				} else if tt.optionalIsNil {
					assert.Equal(t, nil, cfg.OptionalKey.Value())
				} else {
					assert.NotNil(t, cfg.OptionalKey.Value())
				}

				if _, ok := tt.params[""]["requiredkey"]; ok && tt.params[""]["requiredkey"] != "" {
					assert.NotNil(t, cfg.RequiredKey.Value())
				}
			}
		})
	}
}

func generateTestKey(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA private key: %v", err)
	}
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	return privateKey, string(pem.EncodeToMemory(privateKeyPEM))
}

func TestOptionalBasicTypes(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		params     types.ParamValues
		shouldErr  bool
		expectInt  int
		expectBool bool
		expectDur  time.Duration
		expectTime time.Time
	}{
		{
			name: "no value for optional params",
			params: types.ParamValues{
				"": {
					"req": "1",
				},
			},
			shouldErr:  false,
			expectInt:  42,
			expectBool: true,
			expectDur:  time.Hour,
			expectTime: now,
		},
		{
			name: "empty string for optional params",
			params: types.ParamValues{
				"": {
					"i":   "",
					"b":   "",
					"d":   "",
					"t":   "",
					"req": "2",
				},
			},
			shouldErr:  false,
			expectInt:  42,
			expectBool: true,
			expectDur:  time.Hour,
			expectTime: now,
		},
		{
			name: "valid values for optional params",
			params: types.ParamValues{
				"": {
					"i":   "123",
					"b":   "false",
					"d":   "10s",
					"t":   "2024-01-01T00:00:00Z",
					"req": "3",
				},
			},
			shouldErr:  false,
			expectInt:  123,
			expectBool: false,
			expectDur:  10 * time.Second,
			expectTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "empty string for required param",
			params: types.ParamValues{
				"": {
					"req": "",
				},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := struct {
				I   int           `param:",optional"`
				B   bool          `param:",optional"`
				D   time.Duration `param:",optional"`
				T   time.Time     `param:",optional"`
				Req int
			}{
				I:   42,
				B:   true,
				D:   time.Hour,
				T:   now,
				Req: 99,
			}

			testProvider := cfgtest.New(tt.params)
			defer testProvider.Stop()

			_, err := proteus.MustParse(&cfg, proteus.WithProviders(testProvider))

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectInt, cfg.I)
				assert.Equal(t, tt.expectBool, cfg.B)
				assert.Equal(t, tt.expectDur, cfg.D)
				assert.EqualFn(t, tt.expectTime, cfg.T)
			}
		})
	}
}

func TestOptionalXTypes(t *testing.T) {
	tests := []struct {
		name       string
		params     types.ParamValues
		shouldErr  bool
		expectInt  int
		expectBool bool
		expectJSON string
	}{
		{
			name:       "no value for optional xtypes",
			params:     types.ParamValues{"": {"req": "1"}},
			shouldErr:  false,
			expectInt:  88,
			expectBool: true,
			expectJSON: `{"a":"b"}`,
		},
		{
			name: "empty string for optional xtypes",
			params: types.ParamValues{
				"": {
					"xi":  "",
					"xb":  "",
					"xj":  "",
					"req": "2",
				},
			},
			shouldErr:  false,
			expectInt:  88,
			expectBool: true,
			expectJSON: `{"a":"b"}`,
		},
		{
			name: "valid values for optional xtypes",
			params: types.ParamValues{
				"": {
					"xi":  "-5",
					"xb":  "false",
					"xj":  `[1,2]`, // raw json
					"req": "3",
				},
			},
			shouldErr:  false,
			expectInt:  -5,
			expectBool: false,
			expectJSON: `[1,2]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := struct {
				XI  *xtypes.Integer[int] `param:",optional"`
				XB  *xtypes.Bool         `param:",optional"`
				XJ  *xtypes.RawJSON      `param:",optional"`
				Req int
			}{
				XI:  &xtypes.Integer[int]{DefaultValue: 88},
				XB:  &xtypes.Bool{DefaultValue: true},
				XJ:  &xtypes.RawJSON{DefaultValue: []byte(`{"a":"b"}`)},
				Req: 99,
			}

			testProvider := cfgtest.New(tt.params)
			defer testProvider.Stop()

			_, err := proteus.MustParse(&cfg, proteus.WithProviders(testProvider))

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectInt, cfg.XI.Value())
				assert.Equal(t, tt.expectBool, cfg.XB.Value())
				assert.Equal(t, tt.expectJSON, string(cfg.XJ.Value()))
			}
		})
	}
}
