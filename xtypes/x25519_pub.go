package xtypes

import (
	"crypto/ecdh"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"sync"

	"github.com/simplesurance/proteus/types"
)

// X25519PublicKey is a xtype for *ecdh.PublicKey. The key format is expected
// to be on PKIX/PEM format, hex encoded (32 bytes), or raw bytes (optionally
// base64 encoded).
type X25519PublicKey struct {
	DefaultValue  *ecdh.PublicKey
	UpdateFn      func(*ecdh.PublicKey)
	Base64Encoder *base64.Encoding
	content       struct {
		value *ecdh.PublicKey
		mutex sync.Mutex
	}
}

var _ types.XType = &X25519PublicKey{}

// UnmarshalParam parses the input as a string.
func (d *X25519PublicKey) UnmarshalParam(in *string) error {
	var pubK *ecdh.PublicKey
	if in != nil && *in != "" {
		var err error
		pubK, err = parseX25519PublicKey(*in, d.Base64Encoder)
		if err != nil {
			return err
		}
	}

	d.content.mutex.Lock()
	d.content.value = pubK
	d.content.mutex.Unlock()

	if d.UpdateFn != nil {
		d.UpdateFn(d.Value())
	}

	return nil
}

// Value reads the current updated value, taking the default value into
// consideration. If the parameter is not marked as optional, this is
// guaranteed to be not nil.
func (d *X25519PublicKey) Value() *ecdh.PublicKey {
	d.content.mutex.Lock()
	defer d.content.mutex.Unlock()

	if d.content.value == nil {
		return d.DefaultValue
	}

	return d.content.value
}

// ValueValid test if the provided parameter value is valid. Has no side
// effects.
func (d *X25519PublicKey) ValueValid(s string) error {
	if s == "" {
		return types.ErrNoValue
	}
	_, err := parseX25519PublicKey(s, d.Base64Encoder)
	return err
}

// GetDefaultValue will be used to read the default value when showing usage
// information.
func (d *X25519PublicKey) GetDefaultValue() (string, error) {
	// TODO show the public key
	return "<secret>", nil
}

func parseX25519PublicKey(v string, base64Enc *base64.Encoding) (*ecdh.PublicKey, error) {
	var data []byte
	var err error

	if base64Enc != nil {
		data, err = base64Enc.DecodeString(v)
		if err != nil {
			return nil, fmt.Errorf("not a valid base64: %w", err)
		}
	} else {
		data = []byte(v)
	}

	// Try PEM first
	pemBlock, _ := pem.Decode(data)
	if pemBlock != nil {
		if pemBlock.Type != "PUBLIC KEY" {
			return nil, fmt.Errorf("PEM of type %q is not supported. Expected: %q",
				pemBlock.Type,
				"PUBLIC KEY")
		}
		pubK, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error decoding PEM block as ANS.1 public key: %w", err)
		}
		xPubK, ok := pubK.(*ecdh.PublicKey)
		if !ok || xPubK.Curve() != ecdh.X25519() {
			return nil, fmt.Errorf("expected key of type *ecdh.PublicKey (X25519), but got type: %T", pubK)
		}
		return xPubK, nil
	}

	// If not PEM, try hex (as seen in reference code)
	if len(v) == 64 {
		raw, err := hex.DecodeString(v)
		if err == nil && len(raw) == 32 {
			return ecdh.X25519().NewPublicKey(raw)
		}
	}

	// Try raw bytes if it's 32 bytes (maybe it was base64 encoded raw bytes)
	if len(data) == 32 {
		return ecdh.X25519().NewPublicKey(data)
	}

	return nil, fmt.Errorf("invalid X25519 public key: expected PEM, 64-char hex, or 32-byte raw key")
}

