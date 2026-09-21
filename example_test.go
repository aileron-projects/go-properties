package properties_test

import (
	"fmt"
	"os"

	"github.com/aileron-projects/go-properties"
)

func Example() {
	data := `
server.port = 8080
server.tls.cert = /tmp/cert.pem
server.tls.key  = /tmp/key.pem
`
	props, err := properties.Parse([]byte(data))
	if err != nil {
		panic(err)
	}

	fmt.Println("server.port:", props["server.port"])
	fmt.Println("server.tls.cert:", props["server.tls.cert"])
	fmt.Println("server.tls.key:", props["server.tls.key"])
	// Output:
	// server.port: 8080
	// server.tls.cert: /tmp/cert.pem
	// server.tls.key: /tmp/key.pem
}

func ExampleProperties() {
	data := `
server.port = 8080
server.tls.cert = /tmp/cert.pem
server.tls.key  = /tmp/key.pem
debug.enabled = true
`

	mapProps, err := properties.Parse([]byte(data))
	if err != nil {
		panic(err)
	}

	props := properties.NewProperties(mapProps)

	// Set more properties
	props.Set("server.listen", "0.0.0.0")

	// Get properties
	fmt.Println(props.Get("server.listen"))
	fmt.Println(props.GetAs[int]("server.port"))
	fmt.Println(props.GetAs[string]("server.tls.cert"))
	fmt.Println(props.GetAs[string]("server.tls.key"))
	fmt.Println(props.GetAs[bool]("debug.enabled"))
	// Output:
	// 0.0.0.0 <nil>
	// 8080 <nil>
	// /tmp/cert.pem <nil>
	// /tmp/key.pem <nil>
	// true <nil>
}

func ExampleParse_referenceProperty() {
	data := `
foo.bar = baz
reference = ${foo.bar}
`
	props, err := properties.Parse([]byte(data))
	if err != nil {
		panic(err)
	}

	fmt.Println("reference:", props["reference"])
	// Output:
	// reference: baz
}

func ExampleParse_referencePropertyEnv() {
	os.Setenv("FOO_BAR", "BAZ")
	defer func() { os.Unsetenv("FOO_BAR") }()

	data := `
foo.bar = baz
reference = ${foo.bar}
`
	props, err := properties.Parse([]byte(data))
	if err != nil {
		panic(err)
	}

	// Environmental variable is prior to the property in file.
	fmt.Println("reference:", props["reference"])
	// Output:
	// reference: BAZ
}

func ExampleParse_referenceEnv() {
	os.Setenv("EXAMPLE", "example")
	defer func() { os.Unsetenv("EXAMPLE") }()

	data := `reference = ${EXAMPLE}`

	props, err := properties.Parse([]byte(data))
	if err != nil {
		panic(err)
	}

	// Environmental variable is prior to the property in file.
	fmt.Println("reference:", props["reference"])
	// Output:
	// reference: example
}
