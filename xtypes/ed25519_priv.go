package xtypes

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"sync"

	"github.com/simplesurance/proteus/internal/consts"
	"github.com/simplesurance/proteus/types"
)

// Ed25519PrivateKey is a xtype for ed25519.PrivateKey.
type Ed25519PrivateKey struct {
	DefaultValue  ed25519.PrivateKey
	UpdateFn      func(ed25519.PrivateKey)
	Base64Encoder *base64.Encoding
	content       struct {
		value ed25519.PrivateKey
		mutex sync.Mutex
	}
}

var _ types.XType = &Ed25519PrivateKey{}
var _ types.Redactor = &Ed25519PrivateKey{}

// UnmarshalParam parses the input as a string.
func (d *Ed25519PrivateKey) UnmarshalParam(in *string) error {
	var privK ed25519.PrivateKey
	if in != nil {
		var err error
		privK, err = parseEd25519PrivateKey(*in, d.Base64Encoder)
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
func (d *Ed25519PrivateKey) Value() ed25519.PrivateKey {
	d.content.mutex.Lock()
	defer d.content.mutex.Unlock()

	if d.content.value == nil {
		return d.DefaultValue
	}

	return d.content.value
}

// ValueValid test if the provided parameter value is valid. Has no side
// effects.
func (d *Ed25519PrivateKey) ValueValid(s string) error {
	_, err := parseEd25519PrivateKey(s, d.Base64Encoder)
	return err
}

// GetDefaultValue will be used to read the default value when showing usage
// information.
func (d *Ed25519PrivateKey) GetDefaultValue() (string, error) {
	return "<secret>", nil
}

// RedactValue fully redacts the private key, to avoid leaking secrets.
func (d *Ed25519PrivateKey) RedactValue(string) string {
	return consts.RedactedPlaceholder
}

func parseEd25519PrivateKey(v string, base64Enc *base64.Encoding) (ed25519.PrivateKey, error) {
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
		return nil, fmt.Errorf("invalid PEM encoding for Ed25519 private key")
	}

	// support PKCS1 or PKCS8 based on PEM header
	switch pemBlock.Type {
	case "PRIVATE KEY":
		var privK any
		var err error
		privK, err = x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error decoding PEM block as PKCS8: %w", err)
		}

		edPrivK, ok := privK.(ed25519.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("expected key of type ed25519.PrivateKey, but got type: %T", privK)
		}

		return edPrivK, nil
	default:
		return nil, fmt.Errorf("PEM of type %q is not supported. Expected: %q",
			pemBlock.Type,
			"PRIVATE KEY")
	}

}
