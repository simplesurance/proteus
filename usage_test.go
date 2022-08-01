package proteus_test

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/sources/cfgenv"
)

func ExampleParsed_Usage() {
	params := struct {
		Server string `param_desc:"Name of the server to connect"`
		Port   uint16 `param:",optional" param_desc:"Port to conect"`
	}{
		Port: 5432,
	}

	parsed, _ := proteus.MustParse(&params, proteus.WithSources(cfgenv.New("TEST")))

	buffer := bytes.Buffer{}
	parsed.Usage(&buffer)

	scanner := bufio.NewScanner(&buffer)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	// Output:
	// Syntax:
	// ./proteus.test \
	//     <-server string> \
	//     [-port uint16]
	//
	// PARAMETERS
	// - server:string
	//   Name of the server to connect
	// - port:uint16 default=5432
	//   Port to conect
}
