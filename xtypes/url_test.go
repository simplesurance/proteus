//go:build unittest || !integrationtest
// +build unittest !integrationtest

package xtypes_test

import (
	"bytes"
	"errors"
	"net/url"
	"testing"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/internal/assert"
	"github.com/simplesurance/proteus/sources/cfgtest"
	"github.com/simplesurance/proteus/types"
	"github.com/simplesurance/proteus/xtypes"
)

func TestSimpleURL(t *testing.T) {
	const fullURLString = "https://user:pass@asdf.com/x?a=b#segment"

	params := struct {
		URL *xtypes.URL
	}{}

	provider := cfgtest.New(types.ParamValues{
		"": map[string]string{
			"url": fullURLString,
		},
	})

	parsed, err := proteus.MustParse(&params, proteus.WithProviders(provider))
	assert.NoErrorNow(t, err)

	buffer := bytes.Buffer{}
	parsed.Dump(&buffer)
	t.Log("DUMP OF PROVIDED PARAMETERS\n" + buffer.String())

	buffer = bytes.Buffer{}
	parsed.Usage(&buffer)
	t.Log("USAGE INFORMATION\n" + buffer.String())

	assert.Equal(t, fullURLString, params.URL.Value().String())
}

func TestDefaultURL(t *testing.T) {
	const defaultURLString = "https://default"
	defaultURL, _ := url.Parse(defaultURLString)

	params := struct {
		URL *xtypes.URL `param:",optional"`
	}{
		URL: &xtypes.URL{DefaultValue: defaultURL},
	}

	provider := cfgtest.New(types.ParamValues{
		"": map[string]string{},
	})

	parsed, err := proteus.MustParse(&params, proteus.WithProviders(provider))
	assert.NoErrorNow(t, err)

	buffer := bytes.Buffer{}
	parsed.Dump(&buffer)
	t.Log("DUMP OF PROVIDED PARAMETERS\n" + buffer.String())

	buffer = bytes.Buffer{}
	parsed.Usage(&buffer)
	t.Log("USAGE INFORMATION\n" + buffer.String())

	assert.EqualNow(t, defaultURLString, params.URL.Value().String())
}

func TestEmptyURL(t *testing.T) {
	params := struct {
		URL *xtypes.URL
	}{
		URL: &xtypes.URL{ValidateFn: func(_ *url.URL) error { return nil }},
	}

	provider := cfgtest.New(types.ParamValues{
		"": map[string]string{"url": ""},
	})

	parsed, err := proteus.MustParse(&params, proteus.WithProviders(provider))
	assert.NoErrorNow(t, err)

	buffer := bytes.Buffer{}
	parsed.Dump(&buffer)
	t.Log("DUMP OF PROVIDED PARAMETERS\n" + buffer.String())

	buffer = bytes.Buffer{}
	parsed.Usage(&buffer)
	t.Log("USAGE INFORMATION\n" + buffer.String())

	assert.Equal(t, "", params.URL.Value().String())
}

func TestCustomValidator(t *testing.T) {
	var errMsg = "only http is accepted"

	validFn := func(v *url.URL) error {
		if v.Scheme != "https" {
			return errors.New(errMsg)
		}

		return nil
	}

	params := struct {
		URL *xtypes.URL
	}{
		URL: &xtypes.URL{ValidateFn: validFn},
	}

	cases := []struct {
		name      string
		haveURL   string
		wantError bool
	}{
		{
			name:      "ValidURL",
			haveURL:   "https://sisu.sh",
			wantError: false,
		},
		{
			name:      "ValidURL",
			haveURL:   "http://sisu.sh",
			wantError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			provider := cfgtest.New(types.ParamValues{
				"": map[string]string{
					"url": tc.haveURL,
				},
			})

			parsed, err := proteus.MustParse(&params, proteus.WithProviders(provider))

			// parsed is always not-null to allow querying the
			// configuration of the application even when the
			// provided parameters are incorrect

			buffer := bytes.Buffer{}
			parsed.Dump(&buffer)
			t.Log("DUMP OF PROVIDED PARAMETERS\n" + buffer.String())

			buffer = bytes.Buffer{}
			parsed.Usage(&buffer)
			t.Log("USAGE INFORMATION\n" + buffer.String())

			if !tc.wantError {
				assert.NoErrorNow(t, err)
				return
			}

			assert.ErrorNow(t, err)
			assert.StringContains(t, err.Error(), errMsg)
		})
	}
}
