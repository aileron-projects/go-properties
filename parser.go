package properties

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"regexp"
	"strconv"
)

// Parse parses properties from the given bytes.
// Typically Parse parses key-values from files such as .properties file.
// Parse resolves embedded environmental variables in the b.
//
// Input specifications:
//
//	Key syntax:
//		# 'a'-'z', 'A'-'Z', '0'-'9', '_', '.', '-', are allowed for keys.
//		# Whitespaces are allowed around the seperator '='.
//		foo = ok
//		foo=ok
//		foo_bar=ok
//		foo.bar=ok
//		foo-bar=ok
//
//	Single line:
//		# Single quotes and double quotes are removed if entire value is enclosed.
//		foo=bar          >> bar
//		foo="bar"        >> bar
//		foo='bar'        >> bar
//		foo='b"r'        >> b"r
//		foo="b'r"        >> b'r
//
//	Multiple lines:
//		# The "foo" will be "barbaz".
//		# Line breaks of LF and CRLF are removed.
//		# BOTH single quotes and double quotes can be used to enclose multiple lines.
//		foo="
//		bar
//		baz
//		"
//
//	Comments:
//		# Sharp '#' can be used for commenting.
//		# It must not be in the scope of single quotes and double quotes.
//		# It must have at least 1 white space before '#' if the comment is inlined.
//		# comment            >> Comment is appropriately parsed.
//		FOO=BAR # comment    >> Comment is appropriately parsed.
//		FOO=BAR# comment     >> '#' is not parsed as comment. It is considered as a part of the value.
//
//	Escapes:
//		# '\\' can be used for escaping characters by following the 3 rules.
//		# 1. '\\' always escapes special character of ', ", \\, #
//		# 2. '\\' is ignored when it is not in the scope of single quotes or double quotes.
//		# 3. '\\'n or "\n" is replaced to LF, '\\'r or "\r" to CR, '\\'t or "\t" to TAB.
//		foo=ba\"r      >> ba"r
//		foo=ba\'r      >> ba'r
//		foo="ba\"r"    >> ba"r
//		foo=ba\r       >> bar ('\' is not in a scope of single or double quotes.)
//		foo="ba\nr"    >> ba<LF>r (\n is, if in a scope of quotes, converted into a line break.)
//		foo="ba\rr"    >> ba<CR>r (\r is, if in a scope of quotes, converted into a carriage return.)
//		foo="ba\tr"    >> ba<TAB>r (\t is, if in a scope of quotes, converted into a tab space.)
//
//	Referencing:
//		# Other properties and environmental variables can be referenced by ${} with following steps.
//		# 1. Lookup environmental variable.
//		# 2. Convert the key to environmental variable and lookup environmental variable.
//		# 3. Lookup property.
//		# For example, with the following example, value will be resolved by the order of
//		# LookupEnv("bar.baz") > LookupEnv(ToEnvName("bar.baz")) > LookupProperty("bar.baz").
//		# Here, ToEnvName transforms the key to upper cases and replaces '-' and '.' to '_'.
//		FOO=${bar.baz}
func Parse(b []byte) (map[string]string, error) {
	return ParseReader(bytes.NewReader(b))
}

// Parse parses properties from the reader.
// See [Parse] for more details.
func ParseReader(r io.Reader) (map[string]string, error) {
	envs, err := parse(r)
	if err != nil {
		return nil, err
	}
	return envs, nil
}

func parse(r io.Reader) (map[string]string, error) {
	reader := bufio.NewReader(r)
	props := map[string]string{}
	inSingleQuote, inDoubleQuote := false, false
	multilineName, multilineValue := "", ""
	current := 0 // Line number.
	eof := false // End of file flag.
	for !eof {
		current += 1
		line, err := reader.ReadBytes('\n')
		eof = (err == io.EOF)
		if !eof && err != nil {
			return nil, err
		}
		line = bytes.TrimSpace(line)
		line = subst(line, props) // Replace references and environmental variables if any.

		if inSingleQuote || inDoubleQuote {
			var val string
			val, inSingleQuote, inDoubleQuote = scanValue(line, inSingleQuote, inDoubleQuote)
			multilineValue += val
			if !inSingleQuote && !inDoubleQuote {
				props[multilineName] = multilineValue
				multilineName = ""  // Reset variable.
				multilineValue = "" // Reset variable.
			}
			if eof {
				err := errors.New("quotation not closed `" + multilineName + "`")
				return nil, &Error{Inner: err, Type: "parse", Msg: "line " + strconv.Itoa(current)}
			}
		} else {
			name, rest, err := scanName(line)
			if err != nil {
				return nil, &Error{Inner: err, Type: "parse", Msg: "line " + strconv.Itoa(current)}
			}
			if name == "" {
				continue // Maybe comment line.
			}
			var val string
			val, inSingleQuote, inDoubleQuote = scanValue(rest, inSingleQuote, inDoubleQuote)
			if inSingleQuote || inDoubleQuote {
				multilineName = name
				multilineValue = val
			} else {
				props[name] = val
			}
		}
	}
	return props, nil
}

// scanKey scans a line and looks for environmental variable key name.
// If the line is comment, it returns empty key.
// If a key found, it returns the key and the rest of the line after '='.
//
// Example input patterns:
//   - FOO=bar
//   - FOO="bar"
//   - FOO="bar" #comment
//   - FOO="bar
//   - #comment
func scanName(b []byte) (name string, rest []byte, err error) {
	if len(b) == 0 || b[0] == '#' {
		return "", nil, nil
	}
	if b[0] == '=' {
		return "", nil, errors.New("key not found")
	}
	before, after, found := bytes.Cut(b, []byte("="))
	if !found {
		return "", nil, errors.New("key-value seperator `=` not found")
	}
	before = bytes.TrimSpace(before)
	after = bytes.TrimSpace(after)
	for _, c := range before {
		switch {
		case '0' <= c && c <= '9':
		case 'a' <= c && c <= 'z':
		case 'A' <= c && c <= 'Z':
		case c == '.' || c == '-' || c == '_':
		default:
			return "", nil, errors.New("invalid character `" + string(c) + "`")
		}
	}
	return string(before), after, nil
}

// scanValue scans value line.
// It returns the value and the flag of
// in-single-quote or in-double-quote.
// isq and idq don't become true simultaneously.
//
// Example input patterns:
//
//	foobar        <-- Non quoted value
//	"foobar"      <-- Quoted value
//	"foobar       <-- Quoted value. Double quote is not closed.
//	'foobar       <-- Quoted value. Single quote is not closed.
//	"foo"'bar     <-- Quoted value. Single quote is not closed.
//	"foo"'bar'    <-- Non quoted value. All quotes are closed.
//	foo#comment   <-- Non quoted value. The '#' is treated as a value.
//	foo #comment  <-- Commented value. Comments are ignored.
//	#comment      <-- Commented value. Comments are ignored.
func scanValue(b []byte, sq, dq bool) (val string, isq, idq bool) {
	v := make([]byte, 0, len(b))
	escaped := false
	inSingleQuote := sq
	inDoubleQuote := dq
	for _, c := range b {
		if escaped {
			if c == '\\' || c == '\'' || c == '"' || c == '#' {
				v = append(v, c)
				escaped = false
				continue
			}
			if inSingleQuote || inDoubleQuote {
				switch c {
				case 'n':
					v = append(v, '\n')
					escaped = false
					continue
				case 't':
					v = append(v, '\t')
					escaped = false
					continue
				case 'r':
					v = append(v, '\r')
					escaped = false
					continue
				}
			}
			v = append(v, c)
			escaped = false
			continue
		}
		switch c {
		case '\\':
			escaped = true
		case '\'':
			if inDoubleQuote {
				v = append(v, c)
			} else {
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if inSingleQuote {
				v = append(v, c)
			} else {
				inDoubleQuote = !inDoubleQuote
			}
		case '#':
			if !inSingleQuote && !inDoubleQuote && (len(v) > 1 && v[len(v)-1] == ' ') {
				v = bytes.TrimRight(v, " ")
			}
			return string(v), inSingleQuote, inDoubleQuote
		default:
			v = append(v, c)
		}
	}
	if escaped {
		v = append(v, '\\')
	}
	return string(v), inSingleQuote, inDoubleQuote
}

var (
	envRe    = regexp.MustCompile(`\$\{\s*[^{}]*?\s*\}`)
	envEscRe = regexp.MustCompile(`\\\$\\\{\s*.*?\s*\\\}`)
)

func subst(b []byte, priority map[string]string) (sb []byte) {
	sb = b
	sb = envRe.ReplaceAllFunc(sb, func(expression []byte) []byte {
		b := expression[2 : len(expression)-1]
		b = bytes.TrimSpace(b)
		if v, ok := os.LookupEnv(string(b)); ok {
			return []byte(v)
		}
		if v, ok := os.LookupEnv(string(toEnvName(b))); ok {
			return []byte(v)
		}
		if v, ok := priority[string(b)]; ok {
			return []byte(v)
		}
		return expression
	})
	sb = envEscRe.ReplaceAllFunc(sb, func(b []byte) []byte {
		b = b[2 : len(b)-1] // Reuse the buffer.
		b[0], b[1], b[len(b)-1] = '$', '{', '}'
		return b
	})
	return sb
}

func toEnvName(b []byte) []byte {
	bb := bytes.ToUpper(b)
	bb = bytes.ReplaceAll(bb, []byte("-"), []byte("_"))
	bb = bytes.ReplaceAll(bb, []byte("."), []byte("_"))
	return bb
}
