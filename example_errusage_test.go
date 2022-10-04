package proteus_test

import (
	"os"

	"github.com/simplesurance/proteus"
)

func ExampleParsed_WriteError() {
	params := struct {
		Latitude  float64
		Longitude float64
	}{}

	parsed, err := proteus.MustParse(&params)
	if err != nil {
		// "parsed" is never nil; it can be used to parse the error
		parsed.WriteError(os.Stderr, err)
		os.Exit(1)
	}

	// use parameters
}
