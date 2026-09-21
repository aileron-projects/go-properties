package properties_test

import (
	"bytes"
	"testing"

	"github.com/aileron-projects/go-properties"
	"github.com/aileron-projects/go-tester"
)

func TestParse(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		txt := ``
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("comment", func(t *testing.T) {
		txt := `
		# comment line
		foo=foo # inline comment
		`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"foo": "foo"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("basic key syntax", func(t *testing.T) {
		txt := `
		abcdefghijklmnopqrstuvwxyz = ok
		ABCDEFGHIJKLMNOPQRSTUVWXYZ = OK
		1234567890 = ok
		foo.bar = aaa # dot is allowed
		foo-bar = bbb # hiphen is allowed
		foo_bar = ccc # underscore is allowed
		`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{
			"abcdefghijklmnopqrstuvwxyz": "ok",
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ": "OK",
			"1234567890":                 "ok",
			"foo.bar":                    "aaa",
			"foo-bar":                    "bbb",
			"foo_bar":                    "ccc",
		}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("char escape", func(t *testing.T) {
		txt := `FOO=\f\o\o`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"FOO": "foo"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("quotations", func(t *testing.T) {
		txt := `
		NONE=none
		SINGLE='single'
		DOUBLE="double"
		`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"NONE": "none", "SINGLE": "single", "DOUBLE": "double"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("quotations in quotations", func(t *testing.T) {
		txt := `
		SINGLE='single and "double"'
		DOUBLE="'single' and double"
		`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"SINGLE": "single and \"double\"", "DOUBLE": "'single' and double"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("quotations escape", func(t *testing.T) {
		txt := `
		SINGLE='single \'escape\''
		DOUBLE="double \"escape\""
		`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"SINGLE": "single 'escape'", "DOUBLE": "double \"escape\""}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("quotations sequence", func(t *testing.T) {
		txt := `
		SEQ1='Single'"Double"
		SEQ2='Single' and "Double"
		`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"SEQ1": "SingleDouble", "SEQ2": "Single and Double"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("multiline", func(t *testing.T) {
		txt := `
		MULTI1='
		line1
		line2
		'
		MULTI2="
		line1
		line2
		"
		`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"MULTI1": "line1line2", "MULTI2": "line1line2"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("multiline with line break", func(t *testing.T) {
		txt := `
		MULTI1='
		line1\n
		line2
		'
		MULTI2="
		line1\n
		line2
		"
		`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"MULTI1": "line1\nline2", "MULTI2": "line1\nline2"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("replace LF", func(t *testing.T) {
		txt := `FOO="alice\nbob"`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"FOO": "alice\nbob"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("replace CR", func(t *testing.T) {
		txt := `FOO="alice\rbob"`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"FOO": "alice\rbob"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("replace TAB", func(t *testing.T) {
		txt := `FOO="alice\tbob"`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"FOO": "alice\tbob"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("end with escape", func(t *testing.T) {
		txt := `FOO=foo\`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"FOO": "foo\\"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("duplication", func(t *testing.T) {
		txt := `
		FOO=bar
		FOO=foo
		`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"FOO": "foo"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("property reference", func(t *testing.T) {
		txt := `
		greeting.target=alice
		greeting=hello ${greeting.target}
		`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"greeting.target": "alice", "greeting": "hello alice"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("environmental variable", func(t *testing.T) {
		t.Setenv("TARGET", "alice")
		txt := `greeting=hello ${TARGET}`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"greeting": "hello alice"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("environmental variable replacing", func(t *testing.T) {
		t.Setenv("FOO_BAR_BAZ_TEST", "alice")
		txt := `greeting=hello ${foo.bar-baz.test}`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"greeting": "hello alice"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("escape substitution", func(t *testing.T) {
		txt := `FOO=\$\{FOO\}`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"FOO": "${FOO}"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("key not found", func(t *testing.T) {
		txt := `=foo`
		m, err := properties.Parse([]byte(txt))
		tester.AssertDeepEqual(t, nil, m)
		tester.AssertEqualErr(t, &properties.Error{Type: "parse"}, err)
	})
	t.Run("invalid char", func(t *testing.T) {
		txt := `***=foo`
		m, err := properties.Parse([]byte(txt))
		tester.AssertDeepEqual(t, nil, m)
		tester.AssertEqualErr(t, &properties.Error{Type: "parse"}, err)
	})
	t.Run("invalid line format", func(t *testing.T) {
		txt := `foo`
		m, err := properties.Parse([]byte(txt))
		tester.AssertDeepEqual(t, nil, m)
		tester.AssertEqualErr(t, &properties.Error{Type: "parse"}, err)
	})
	t.Run("quotation not closed", func(t *testing.T) {
		txt := `
		MULTI='
		line1
		line2
		`
		m, err := properties.Parse([]byte(txt))
		tester.AssertDeepEqual(t, nil, m)
		tester.AssertEqualErr(t, &properties.Error{Type: "parse"}, err)
	})
	t.Run("env subst error", func(t *testing.T) {
		txt := `FOO=${!FOO}`
		m, err := properties.Parse([]byte(txt))
		want := map[string]string{"FOO": "${!FOO}"}
		tester.AssertDeepEqual(t, want, m)
		tester.AssertEqualErr(t, nil, err)
	})
}

func TestParseReader(t *testing.T) {
	t.Run("read error", func(t *testing.T) {
		txt := `
			FOO=foo
			BAR=bar
		`
		r := tester.MaxErrorReader(bytes.NewReader([]byte(txt)), 5)
		m, err := properties.ParseReader(r)
		tester.AssertDeepEqual(t, nil, m)
		tester.AssertEqualErr(t, tester.ErrMaxRead, err)
	})
}
