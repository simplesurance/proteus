package xtypes_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"reflect"
	"testing"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/internal/assert"
	"github.com/simplesurance/proteus/sources/cfgtest"
	"github.com/simplesurance/proteus/types"
	"github.com/simplesurance/proteus/xtypes"
)

func TestECDSAPublicKey(t *testing.T) {
	_, publicKeyStr := generateTestECPubKey(t)
	defaultKey, _ := generateTestECPubKey(t)

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
					"optionalkey": publicKeyStr,
					"requiredkey": publicKeyStr,
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
					"requiredkey": publicKeyStr,
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
					"requiredkey": publicKeyStr,
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
					"requiredkey": publicKeyStr,
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
				OptionalKey *xtypes.ECDSAPubKey `param:",optional"`
				RequiredKey *xtypes.ECDSAPubKey
			}{}

			if tt.useDefault {
				cfg.OptionalKey = &xtypes.ECDSAPubKey{DefaultValue: defaultKey}
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
					assert.True(t, reflect.DeepEqual(defaultKey, cfg.OptionalKey.Value()), "default key should be used")
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

func generateTestECPubKey(t *testing.T) (*ecdsa.PublicKey, string) {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA private key: %v", err)
	}
	derBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal ECDSA public key: %v", err)
	}
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derBytes,
	}
	return &privateKey.PublicKey, string(pem.EncodeToMemory(pemBlock))
}
