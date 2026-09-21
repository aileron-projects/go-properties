<!-- markdownlint-disable MD033 MD041 -->

<div align="center">

[![Release](https://img.shields.io/github/v/release/aileron-projects/go-properties?sort=semver)](https://github.com/aileron-projects/go-properties/releases)
[![Reference](https://pkg.go.dev/badge/github.com/aileron-projects/go-properties.svg)](https://pkg.go.dev/github.com/aileron-projects/go-properties)
[![DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/aileron-projects/go-properties)
[![Test](https://github.com/aileron-projects/go-properties/actions/workflows/test.yaml/badge.svg)](https://github.com/aileron-projects/go-properties/actions/workflows/test.yaml)

[![Insights](https://badgen.net/badge/Insights/open%2Fsource%2Finsights/cyan)](https://deps.dev/go/github.com%2Faileron-projects%2Fgo-properties)
[![Insights](https://badgen.net/badge/Insights/OSS%2FInsight/orange)](https://ossinsight.io/analyze/aileron-projects/go-properties)

</div>

# go-properties

**Loading properties files for Go.**

`go-properties` is a lightweight Go library for loading and managing `.properties` configuration files.

Parse Java-style `.properties` files with support for quoting, escaping, multi-line values, and variable references (environment variables and cross-property lookups).
Includes a trie-based lookup structure and optional autoloading of profile-specific configuration files, inspired by Spring Boot's configuration model.

## Features

- Load properties files.
- Trie tree based property lookup.
- Autoload properties files.

## Usages

### Properties file syntax

**Properties key syntax:**

```txt
# 'a'-'z', 'A'-'Z', '0'-'9', '_', '.', '-', are allowed for keys.
# Whitespaces are allowed around the seperator '='.
foo = ok
foo=ok
foo_bar=ok
foo.bar=ok
foo-bar=ok
```

**Single line:**

```txt
# Single quotes and double quotes are removed if entire value is enclosed.
foo=bar          >> bar
foo="bar"        >> bar
foo='bar'        >> bar
foo='b"r'        >> b"r
foo="b'r"        >> b'r
```

**Multiple lines:**

```txt
# The "foo" will be "barbaz".
# Line breaks of LF and CRLF are removed.
# BOTH single quotes and double quotes can be used to enclose multiple lines.
foo="
bar
baz
"
```

**Comments:**

```txt
# Sharp '#' can be used for commenting.
# It must not be in the scope of single quotes and double quotes.
# It must have at least 1 white space before '#' if the comment is inlined.
# comment            >> Comment is appropriately parsed.
FOO=BAR # comment    >> Comment is appropriately parsed.
FOO=BAR# comment     >> '#' is not parsed as comment. It is considered as a part of the value.
```

**Escapes:**

```txt
# '\\' can be used for escaping characters by following the 3 rules.
# 1. '\\' always escapes special character of ', ", \\, #
# 2. '\\' is ignored when it is not in the scope of single quotes or double quotes.
# 3. '\\'n or "\n" is replaced to LF, '\\'r or "\r" to CR, '\\'t or "\t" to TAB.
foo=ba\"r      >> ba"r
foo=ba\'r      >> ba'r
foo="ba\"r"    >> ba"r
foo=ba\r       >> bar ('\' is not in a scope of single or double quotes.)
foo="ba\nr"    >> ba<LF>r (\n is, if in a scope of quotes, converted into a line break.)
foo="ba\rr"    >> ba<CR>r (\r is, if in a scope of quotes, converted into a carriage return.)
foo="ba\tr"    >> ba<TAB>r (\t is, if in a scope of quotes, converted into a tab space.)
```

**Referencing (Environmental variables and other properties):**

```txt
# Other properties and environmental variables can be referenced by ${} with following steps.
# 1. Lookup environmental variable.
# 2. Convert the key to environmental variable and lookup environmental variable.
# 3. Lookup property.
# For example, with the following example, value will be resolved by the order of
# LookupEnv("bar.baz") > LookupEnv(ToEnvName("bar.baz")) > LookupProperty("bar.baz").
# Here, ToEnvName transforms the key to upper cases and replaces '-' and '.' to '_'.
FOO=${bar.baz}
```

### Parse properties

Properties can be parsed with the methods:

- `Parse([]byte) (map[string]string, error)`
- `ParseReader(io.Reader) (map[string]string, error)`

For example:

```go
data := `
server.port = 8080
server.tls.cert = /tmp/cert.pem
server.tls.key  = /tmp/key.pem
`
props, err := properties.Parse([]byte(data))
if err != nil {
    panic(err)
}

fmt.Println(props["server.port"])     // "8080"
fmt.Println(props["server.tls.cert"]) // "/tmp/cert.pem"
fmt.Println(props["server.tls.key"])  // "/tmp/key.pem"
```

### Get/Set properties

`Properties` type provides useful methods.

- `Set(key string, value any)`: set properties from code
- `Get(key string) (any, error)`: get properties with a key
- `GetAs[T any](key string) (T, error)`: get properties as the specified type

Properties can be created from parsed properties in map.

Example:

```go
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

// Create a properties.
props := properties.NewProperties(mapProps)

// Set more properties
props.Set("server.listen", "0.0.0.0")

// Get properties
fmt.Println(props.Get("server.listen"))             // "0.0.0.0"
fmt.Println(props.GetAs[int]("server.port"))        // 8080
fmt.Println(props.GetAs[string]("server.tls.cert")) // "/tmp/cert.pem"
fmt.Println(props.GetAs[string]("server.tls.key"))  // "/tmp/key.pem"
fmt.Println(props.GetAs[bool]("debug.enabled"))     // true
```

### Autoload properties

`autoload` package provides autoloading of properties files.
It loads `application.properties` and `application-${profile}.properties` by default.
The `${profile}` can be set with the environmental variable `PROPERTIES_ACTIVE_PROFILE`.

Example:

```go
import (
    "fmt"
    "github.com/aileron-projects/go-properties/autoload"
)

func main() {
    props := autoload.Props
    fmt.Println(props.GetAs[string]("server.port"))
    fmt.Println(props.GetAs[string]("server.listen"))
    fmt.Println(props.GetAs[int]("server.read-timeout"))
}
```

Directory and file names can be set from code using the global variables.

```go
var (
    PropsDir       = "./"
    PropsPath      = "application.properties"
    ProfilePattern = "application-${profile}.properties"
    Profile        = os.Getenv("PROPERTIES_ACTIVE_PROFILE")
)
```

### Referencing environment variables

Environment variables can be referenced by `${...}`.
If the variable is not set, the `${...}` is left as it is.

`${...}` can be escaped by `\$\{...\}`.

```txt
reference = ${FOO}
escaped = \$\{BAR\}
```

### Referencing other property

Other properties can be referenced with `${...}` like environment variables.

Environment variable `FOO_BAR` is prior to the property `foo.bar` when resolving the references of `${...}`.

```txt
foo.bar = baz
reference = ${foo.bar}
```

## Docs & Examples

- GoDoc: <https://pkg.go.dev/github.com/aileron-projects/go-properties>
- Examples:
  - [example_test.go](./example_test.go)
  - autoloading: [examples/autoloading/](./examples/autoloading/)

## References

- [Externalized Configuration - SpringBoot](https://docs.spring.io/spring-boot/reference/features/external-config.html)
