package xtypes

import (
	"crypto/ecdh"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"sync"

	"github.com/simplesurance/proteus/internal/consts"
	"github.com/simplesurance/proteus/types"
)

// X25519PrivateKey is a xtype for *ecdh.PrivateKey. The key format is expected
// to be on PKCS8/PEM format, hex encoded (32 bytes), or raw bytes (optionally
// base64 encoded). The full value is fully redacted to avoid leaking secrets.
type X25519PrivateKey struct {
	DefaultValue  *ecdh.PrivateKey
	UpdateFn      func(*ecdh.PrivateKey)
	Base64Encoder *base64.Encoding
	content       struct {
		value *ecdh.PrivateKey
		mutex sync.Mutex
	}
}

var _ types.XType = &X25519PrivateKey{}
var _ types.Redactor = &X25519PrivateKey{}

// UnmarshalParam parses the input as a string.
func (d *X25519PrivateKey) UnmarshalParam(in *string) error {
	var privK *ecdh.PrivateKey
	if in != nil && *in != "" {
		var err error
		privK, err = parseX25519PrivateKey(*in, d.Base64Encoder)
		if err != nil {
			return err
		}
	}

	d.content.mutex.Lock()
	d.content.value = privK
	d.content.mutex.Unlock()

	if d.UpdateFn != nil {
		d.UpdateFn(d.Value())
	}

	return nil
}

// Value reads the current updated value, taking the default value into
// consideration. If the parameter is not marked as optional, this is
// guaranteed to be not nil.
func (d *X25519PrivateKey) Value() *ecdh.PrivateKey {
	d.content.mutex.Lock()
	defer d.content.mutex.Unlock()

	if d.content.value == nil {
		return d.DefaultValue
	}

	return d.content.value
}

// ValueValid test if the provided parameter value is valid. Has no side
// effects.
func (d *X25519PrivateKey) ValueValid(s string) error {
	if s == "" {
		return types.ErrNoValue
	}
	_, err := parseX25519PrivateKey(s, d.Base64Encoder)
	return err
}

// GetDefaultValue will be used to read the default value when showing usage
// information.
func (d *X25519PrivateKey) GetDefaultValue() (string, error) {
	return "<secret>", nil
}

// RedactValue fully redacts the private key, to avoid leaking secrets.
func (d *X25519PrivateKey) RedactValue(string) string {
	return consts.RedactedPlaceholder
}

func parseX25519PrivateKey(v string, base64Enc *base64.Encoding) (*ecdh.PrivateKey, error) {
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
		if pemBlock.Type != "PRIVATE KEY" {
			return nil, fmt.Errorf("PEM of type %q is not supported. Expected: %q",
				pemBlock.Type,
				"PRIVATE KEY")
		}
		privK, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error decoding PEM block as PKCS8: %w", err)
		}
		xPrivK, ok := privK.(*ecdh.PrivateKey)
		if !ok || xPrivK.Curve() != ecdh.X25519() {
			return nil, fmt.Errorf("expected key of type *ecdh.PrivateKey (X25519), but got type: %T", privK)
		}
		return xPrivK, nil
	}

	// If not PEM, try hex (as seen in reference code)
	if len(v) == 64 {
		raw, err := hex.DecodeString(v)
		if err == nil && len(raw) == 32 {
			return ecdh.X25519().NewPrivateKey(raw)
		}
	}

	// Try raw bytes if it's 32 bytes (maybe it was base64 encoded raw bytes)
	if len(data) == 32 {
		return ecdh.X25519().NewPrivateKey(data)
	}

	return nil, fmt.Errorf("invalid X25519 private key: expected PEM, 64-char hex, or 32-byte raw key")
}
