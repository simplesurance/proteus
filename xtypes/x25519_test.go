package xtypes_test

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"testing"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/internal/assert"
	"github.com/simplesurance/proteus/sources/cfgtest"
	"github.com/simplesurance/proteus/types"
	"github.com/simplesurance/proteus/xtypes"
)

func TestX25519Keys(t *testing.T) {
	priv, privPEM := generateTestX25519PrivKey(t)
	pub, pubPEM := generateTestX25519PubKey(t, priv)

	cfg := struct {
		Priv *xtypes.X25519PrivateKey
		Pub  *xtypes.X25519PubKey
	}{}

	testProvider := cfgtest.New(types.ParamValues{
		"": map[string]string{
			"priv": privPEM,
			"pub":  pubPEM,
		},
	})
	defer testProvider.Stop()

	_, err := proteus.MustParse(&cfg, proteus.WithProviders(testProvider))
	assert.NoErrorNow(t, err)

	assert.True(t, priv.Equal(cfg.Priv.Value()), "private key should match")
	assert.True(t, pub.Equal(cfg.Pub.Value()), "public key should match")
}

func TestX25519KeysHex(t *testing.T) {
	priv, _ := generateTestX25519PrivKey(t)
	pub := priv.PublicKey()

	privHex := hex.EncodeToString(priv.Bytes())
	pubHex := hex.EncodeToString(pub.Bytes())

	cfg := struct {
		Priv *xtypes.X25519PrivateKey
		Pub  *xtypes.X25519PubKey
	}{}

	testProvider := cfgtest.New(types.ParamValues{
		"": map[string]string{
			"priv": privHex,
			"pub":  pubHex,
		},
	})
	defer testProvider.Stop()

	_, err := proteus.MustParse(&cfg, proteus.WithProviders(testProvider))
	assert.NoErrorNow(t, err)

	assert.True(t, priv.Equal(cfg.Priv.Value()), "private key should match (hex)")
	assert.True(t, pub.Equal(cfg.Pub.Value()), "public key should match (hex)")
}

func TestX25519KeysBase64Raw(t *testing.T) {
	priv, _ := generateTestX25519PrivKey(t)
	pub := priv.PublicKey()

	privB64 := base64.StdEncoding.EncodeToString(priv.Bytes())
	pubB64 := base64.StdEncoding.EncodeToString(pub.Bytes())

	cfg := struct {
		Priv *xtypes.X25519PrivateKey
		Pub  *xtypes.X25519PubKey
	}{
		Priv: &xtypes.X25519PrivateKey{Base64Encoder: base64.StdEncoding},
		Pub:  &xtypes.X25519PubKey{Base64Encoder: base64.StdEncoding},
	}

	testProvider := cfgtest.New(types.ParamValues{
		"": map[string]string{
			"priv": privB64,
			"pub":  pubB64,
		},
	})
	defer testProvider.Stop()

	_, err := proteus.MustParse(&cfg, proteus.WithProviders(testProvider))
	assert.NoErrorNow(t, err)

	assert.True(t, priv.Equal(cfg.Priv.Value()), "private key should match (b64 raw)")
	assert.True(t, pub.Equal(cfg.Pub.Value()), "public key should match (b64 raw)")
}

func TestX25519KeysBase64PEM(t *testing.T) {
	priv, privPEM := generateTestX25519PrivKey(t)
	pub, pubPEM := generateTestX25519PubKey(t, priv)

	privB64 := base64.StdEncoding.EncodeToString([]byte(privPEM))
	pubB64 := base64.StdEncoding.EncodeToString([]byte(pubPEM))

	cfg := struct {
		Priv *xtypes.X25519PrivateKey
		Pub  *xtypes.X25519PubKey
	}{
		Priv: &xtypes.X25519PrivateKey{Base64Encoder: base64.StdEncoding},
		Pub:  &xtypes.X25519PubKey{Base64Encoder: base64.StdEncoding},
	}

	testProvider := cfgtest.New(types.ParamValues{
		"": map[string]string{
			"priv": privB64,
			"pub":  pubB64,
		},
	})
	defer testProvider.Stop()

	_, err := proteus.MustParse(&cfg, proteus.WithProviders(testProvider))
	assert.NoErrorNow(t, err)

	assert.True(t, priv.Equal(cfg.Priv.Value()), "private key should match (b64 PEM)")
	assert.True(t, pub.Equal(cfg.Pub.Value()), "public key should match (b64 PEM)")
}

func TestX25519ValueValid(t *testing.T) {
	_, privPEM := generateTestX25519PrivKey(t)
	priv := &xtypes.X25519PrivateKey{}

	assert.NoError(t, priv.ValueValid(privPEM))
	assert.NoError(t, priv.ValueValid(hex.EncodeToString(make([]byte, 32))))
	assert.Error(t, priv.ValueValid("not a key"))
	assert.Error(t, priv.ValueValid(""))

	_, pubPEM := generateTestX25519PubKey(t, nil)
	pub := &xtypes.X25519PubKey{}

	assert.NoError(t, pub.ValueValid(pubPEM))
	assert.NoError(t, pub.ValueValid(hex.EncodeToString(make([]byte, 32))))
	assert.Error(t, pub.ValueValid("not a key"))
	assert.Error(t, pub.ValueValid(""))
}

func generateTestX25519PrivKey(t *testing.T) (*ecdh.PrivateKey, string) {
	t.Helper()
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate X25519 private key: %v", err)
	}
	derBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("failed to marshal X25519 private key: %v", err)
	}
	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: derBytes,
	}
	return priv, string(pem.EncodeToMemory(pemBlock))
}

func generateTestX25519PubKey(t *testing.T, priv *ecdh.PrivateKey) (*ecdh.PublicKey, string) {
	t.Helper()
	var pub *ecdh.PublicKey
	if priv != nil {
		pub = priv.PublicKey()
	} else {
		newPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
		if err != nil {
			t.Fatalf("failed to generate X25519 private key: %v", err)
		}
		pub = newPriv.PublicKey()
	}

	derBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatalf("failed to marshal X25519 public key: %v", err)
	}
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derBytes,
	}
	return pub, string(pem.EncodeToMemory(pemBlock))
}

func TestX25519PubKey_GetDefaultValue(t *testing.T) {
	pub, _ := generateTestX25519PubKey(t, nil)
	pubHex := hex.EncodeToString(pub.Bytes())

	t.Run("nil default", func(t *testing.T) {
		xt := &xtypes.X25519PubKey{DefaultValue: nil}
		val, err := xt.GetDefaultValue()
		assert.NoError(t, err)
		assert.Equal(t, "", val)
	})

	t.Run("standard encoding (hex)", func(t *testing.T) {
		xt := &xtypes.X25519PubKey{DefaultValue: pub}
		val, err := xt.GetDefaultValue()
		assert.NoError(t, err)
		assert.Equal(t, pubHex, val)
	})

	t.Run("with base64 encoder", func(t *testing.T) {
		xt := &xtypes.X25519PubKey{
			DefaultValue:  pub,
			Base64Encoder: base64.StdEncoding,
		}
		val, err := xt.GetDefaultValue()
		assert.NoError(t, err)
		assert.Equal(t, base64.StdEncoding.EncodeToString(pub.Bytes()), val)
	})
}
