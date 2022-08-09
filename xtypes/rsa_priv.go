package xtypes

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"sync"

	"github.com/simplesurance/proteus/internal/consts"
	"github.com/simplesurance/proteus/types"
)

// RSAPrivateKey is a xtype for values of type *rsa.PrivateKey. The
// key format is expected to be on PKGCS8/PEM format, optionally base64
// encoded. The full value is fully redacted to avoid leaking secrets.
type RSAPrivateKey struct {
	DefaultValue  *rsa.PrivateKey
	UpdateFn      func(*rsa.PrivateKey)
	Base64Encoder *base64.Encoding
	content       struct {
		value *rsa.PrivateKey
		mutex sync.Mutex
	}
}

var _ types.XType = &RSAPrivateKey{}
var _ types.Redactor = &RSAPrivateKey{}

// UnmarshalParam parses the input as a string.
func (d *RSAPrivateKey) UnmarshalParam(in *string) error {
	var privK *rsa.PrivateKey
	if in != nil {
		var err error
		privK, err = parseRSAPriv(*in, d.Base64Encoder)
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
func (d *RSAPrivateKey) Value() *rsa.PrivateKey {
	d.content.mutex.Lock()
	defer d.content.mutex.Unlock()

	if d.content.value == nil {
		return d.DefaultValue
	}

	return d.content.value
}

// ValueValid test if the provided parameter value is valid. Has no side
// effects.
func (d *RSAPrivateKey) ValueValid(s string) error {
	_, err := parseRSAPriv(s, d.Base64Encoder)
	return err
}

// GetDefaultValue will be used to read the default value when showing usage
// information.
func (d *RSAPrivateKey) GetDefaultValue() (string, error) {
	return "<secret>", nil
}

// RedactValue fully redacts the secret to avoid leaking it.
func (d *RSAPrivateKey) RedactValue(string) string {
	return consts.RedactedPlaceholder
}

func parseRSAPriv(v string, base64Enc *base64.Encoding) (*rsa.PrivateKey, error) {
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
		return nil, fmt.Errorf("invalid PEM encoding for RSA private key")
	}

	// support PKCS1 or PKCS8 based on PEM header
	switch pemBlock.Type {
	case "RSA PRIVATE KEY":
		rsaPrivK, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error decoding PEM block as PKCS1: %w", err)
		}

		return rsaPrivK, nil
	case "PRIVATE KEY":
		var privK any
		var err error
		privK, err = x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error decoding PEM block as PKCS8: %w", err)
		}

		rsaPrivK, ok := privK.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("expected key of type *rsa.PrivateKey, but got type: %T", privK)
		}

		return rsaPrivK, nil
	default:
		return nil, fmt.Errorf("PEM of type %q is not supported. Expected: %q or %q",
			pemBlock.Type,
			"RSA PRIVATE KEY",
			"PRIVATE KEY")
	}

}
