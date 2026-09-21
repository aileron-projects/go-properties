package properties_test

import (
	"testing"

	"github.com/aileron-projects/go-properties"
	"github.com/aileron-projects/go-tester"
)

func TestLoad(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got, err := properties.Load("testdata/empty.txt")
		want := map[string]string{}
		tester.AssertDeepEqual(t, want, got)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("basic key syntax", func(t *testing.T) {
		got, err := properties.Load("testdata/key-syntax.txt")
		want := map[string]string{
			"foo.bar": "1",
			"foo-bar": "2",
			"foo_bar": "3",
		}
		tester.AssertDeepEqual(t, want, got)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("comment", func(t *testing.T) {
		got, err := properties.Load("testdata/comment.txt")
		want := map[string]string{"FOO": "foo"}
		tester.AssertDeepEqual(t, want, got)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("multiple files", func(t *testing.T) {
		got, err := properties.Load("testdata/single-line-foo.txt", "testdata/single-line-bar.txt")
		want := map[string]string{"FOO": "foo", "BAR": "bar"}
		tester.AssertDeepEqual(t, want, got)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("duplication", func(t *testing.T) {
		got, err := properties.Load("testdata/single-line-foo.txt", "testdata/single-line-bar.txt", "testdata/single-line-override.txt")
		want := map[string]string{"FOO": "foooo", "BAR": "baaar", "BAZ": "baz"}
		tester.AssertDeepEqual(t, want, got)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("referencing env", func(t *testing.T) {
		t.Setenv("BAR_BAZ", "ok")
		got, err := properties.Load("testdata/referencing-env.txt")
		want := map[string]string{"foo": "ok"}
		tester.AssertDeepEqual(t, want, got)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("referencing env property", func(t *testing.T) {
		t.Setenv("BAR_BAZ", "ok")
		got, err := properties.Load("testdata/referencing-property.txt")
		want := map[string]string{"bar.baz": "test", "foo": "ok"}
		tester.AssertDeepEqual(t, want, got)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("referencing property", func(t *testing.T) {
		got, err := properties.Load("testdata/referencing-property.txt")
		want := map[string]string{"bar.baz": "test", "foo": "test"}
		tester.AssertDeepEqual(t, want, got)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("referencing non exist", func(t *testing.T) {
		got, err := properties.Load("testdata/referencing-non-exist.txt")
		want := map[string]string{"foo": "${!FOO}"}
		tester.AssertDeepEqual(t, want, got)
		tester.AssertEqualErr(t, nil, err)
	})
	t.Run("properties parse error", func(t *testing.T) {
		got, err := properties.Load("testdata/invalid-char.txt")
		tester.AssertDeepEqual(t, nil, got)
		tester.AssertEqualErr(t, &properties.Error{Type: "parse"}, err)
	})
	t.Run("read file error", func(t *testing.T) {
		got, err := properties.Load("testdata/not-exist.txt")
		tester.AssertDeepEqual(t, nil, got)
		tester.AssertEqualErr(t, &properties.Error{Type: "load"}, err)
	})
}
