package xtypes

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"sync"

	"github.com/simplesurance/proteus/types"
)

// ECDSAPubKey is a xtype for values of type *ecdsa.PublicKey.
type ECDSAPubKey struct {
	DefaultValue  *ecdsa.PublicKey
	UpdateFn      func(*ecdsa.PublicKey)
	Base64Encoder *base64.Encoding
	content       struct {
		value *ecdsa.PublicKey
		mutex sync.Mutex
	}
}

var _ types.XType = &ECDSAPubKey{}

// UnmarshalParam parses the input as a string.
func (d *ECDSAPubKey) UnmarshalParam(in *string) error {
	var pubK *ecdsa.PublicKey
	if in != nil {
		var err error
		pubK, err = parseECPubKey(*in, d.Base64Encoder)
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
func (d *ECDSAPubKey) Value() *ecdsa.PublicKey {
	d.content.mutex.Lock()
	defer d.content.mutex.Unlock()

	if d.content.value == nil {
		return d.DefaultValue
	}

	return d.content.value
}

// ValueValid test if the provided parameter value is valid. Has no side
// effects.
func (d *ECDSAPubKey) ValueValid(s string) error {
	_, err := parseECPubKey(s, d.Base64Encoder)
	return err
}

// GetDefaultValue will be used to read the default value when showing usage
// information.
func (d *ECDSAPubKey) GetDefaultValue() (string, error) {
	// FIXME show the public key
	return "<secret>", nil
}

func parseECPubKey(v string, base64Enc *base64.Encoding) (*ecdsa.PublicKey, error) {
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
		return nil, fmt.Errorf("invalid PEM encoded public key")
	}

	// support PKCS1 or PKCS8 based on PEM header
	switch pemBlock.Type {
	case "PUBLIC KEY":
		var pubK any
		var err error
		pubK, err = x509.ParsePKIXPublicKey(pemBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error decoding PEM block as ANS.1 public key: %w", err)
		}

		ecpubK, ok := pubK.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("expected key of type *ecdsa.pubateKey, but got type: %T", pubK)
		}

		return ecpubK, nil
	default:
		return nil, fmt.Errorf("PEM of type %q is not supported. Expected: %q",
			pemBlock.Type,
			"PUBLIC KEY")
	}

}
