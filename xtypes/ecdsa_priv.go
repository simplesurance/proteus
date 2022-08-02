package xtypes

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"sync"

	"github.com/simplesurance/proteus/internal/consts"
	"github.com/simplesurance/proteus/types"
)

// ECDSAPrivateKey is a xtype for *ecdsa.PrivateKey.
type ECDSAPrivateKey struct {
	DefaultValue  *ecdsa.PrivateKey
	UpdateFn      func(*ecdsa.PrivateKey)
	Base64Encoder *base64.Encoding
	content       struct {
		value *ecdsa.PrivateKey
		mutex sync.Mutex
	}
}

var _ types.XType = &ECDSAPrivateKey{}
var _ types.Redactor = &ECDSAPrivateKey{}

// UnmarshalParam parses the input as a string.
func (d *ECDSAPrivateKey) UnmarshalParam(in *string) error {
	var privK *ecdsa.PrivateKey
	if in != nil {
		var err error
		privK, err = parseECPrivKey(*in, d.Base64Encoder)
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
func (d *ECDSAPrivateKey) Value() *ecdsa.PrivateKey {
	d.content.mutex.Lock()
	defer d.content.mutex.Unlock()

	if d.content.value == nil {
		return d.DefaultValue
	}

	return d.content.value
}

// ValueValid test if the provided parameter value is valid. Has no side
// effects.
func (d *ECDSAPrivateKey) ValueValid(s string) error {
	_, err := parseECPrivKey(s, d.Base64Encoder)
	return err
}

// GetDefaultValue will be used to read the default value when showing usage
// information.
func (d *ECDSAPrivateKey) GetDefaultValue() (string, error) {
	return "<secret>", nil
}

// RedactValue fully redacts the private key, to avoid leaking secrets.
func (d *ECDSAPrivateKey) RedactValue(string) string {
	return consts.RedactedPlaceholder
}

func parseECPrivKey(v string, base64Enc *base64.Encoding) (*ecdsa.PrivateKey, error) {
	var pemData []byte

	// use base64 encoding, if requested
	if base64Enc != nil {
		var err error
		pemData, err = base64Enc.DecodeString(v)
		if err != nil {
			return nil, fmt.Errorf("not a valid base64: %w", err)
		}
	} else {
		pemData = []byte(v)
	}

	// parse PEM
	pemBlock, _ := pem.Decode(pemData)
	if pemBlock == nil {
		return nil, fmt.Errorf("invalid PEM encoding for ECDSA private key")
	}

	// support PKCS1 or PKCS8 based on PEM header
	switch pemBlock.Type {
	case "EC PRIVATE KEY":
		ecPrivK, err := x509.ParseECPrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error decoding PEM block as PKCS1: %w", err)
		}

		return ecPrivK, nil
	case "PRIVATE KEY":
		var privK any
		var err error
		privK, err = x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error decoding PEM block as PKCS8: %w", err)
		}

		ecPrivK, ok := privK.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("expected key of type *ecdsa.PrivateKey, but got type: %T", privK)
		}

		return ecPrivK, nil
	default:
		return nil, fmt.Errorf("PEM of type %q is not supported. Expected: %q or %q",
			pemBlock.Type,
			"ECDSA PRIVATE KEY",
			"PRIVATE KEY")
	}

}
