package properties_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/aileron-projects/go-properties"
	"github.com/aileron-projects/go-tester"
)

func TestNewProperties(t *testing.T) {
	t.Parallel()
	t.Run("nil map", func(t *testing.T) {
		p := properties.NewProperties(nil)
		v, err := p.Get("test")
		tester.AssertEqual(t, nil, v)
		tester.AssertEqualErr(t, properties.ErrNotFound, err)
	})
	t.Run("non-nil map", func(t *testing.T) {
		p := properties.NewProperties(map[string]string{"test": "value"})
		v, err := p.Get("test")
		tester.AssertEqual(t, any("value"), v)
		tester.AssertEqualErr(t, nil, err)
	})
}

func TestProperties(t *testing.T) {
	t.Parallel()
	t.Run("supported data types", func(t *testing.T) {
		p := properties.NewProperties(nil)
		p.Set("string", "test")
		p.Set("bool", "true")
		p.Set("int", "-123")
		p.Set("int8", "-123")
		p.Set("int16", "-123")
		p.Set("int32", "-123")
		p.Set("int64", "-123")
		p.Set("uint", "123")
		p.Set("uint8", "123")
		p.Set("uint16", "123")
		p.Set("uint32", "123")
		p.Set("uint64", "123")
		p.Set("float32", "123.45")
		p.Set("float64", "123.45")
		p.Set("complex64", "123+456i")
		p.Set("complex128", "123+456i")
		tester.AssertEqual(t, "test", mustGet[string](p, "string"))
		tester.AssertEqual(t, true, mustGet[bool](p, "bool"))
		tester.AssertEqual(t, -123, mustGet[int](p, "int"))
		tester.AssertEqual(t, -123, mustGet[int8](p, "int8"))
		tester.AssertEqual(t, -123, mustGet[int16](p, "int16"))
		tester.AssertEqual(t, -123, mustGet[int32](p, "int32"))
		tester.AssertEqual(t, -123, mustGet[int64](p, "int64"))
		tester.AssertEqual(t, 123, mustGet[uint](p, "uint"))
		tester.AssertEqual(t, 123, mustGet[uint8](p, "uint8"))
		tester.AssertEqual(t, 123, mustGet[uint16](p, "uint16"))
		tester.AssertEqual(t, 123, mustGet[uint32](p, "uint32"))
		tester.AssertEqual(t, 123, mustGet[uint64](p, "uint64"))
		tester.AssertEqual(t, 123.45, mustGet[float32](p, "float32"))
		tester.AssertEqual(t, 123.45, mustGet[float64](p, "float64"))
		tester.AssertEqual(t, 123+456i, mustGet[complex64](p, "complex64"))
		tester.AssertEqual(t, 123+456i, mustGet[complex128](p, "complex128"))
	})
	t.Run("unsupported string type", func(t *testing.T) {
		p := properties.NewProperties(nil)
		p.Set("writer", "test")
		w, err := p.GetAs[io.Writer]("writer")
		tester.AssertEqualErr(t, &properties.Error{Type: "convert"}, err)
		tester.AssertEqual(t, nil, w)
	})
	t.Run("type conversion success", func(t *testing.T) {
		p := properties.NewProperties(nil)
		var buf bytes.Buffer
		p.Set("buffer", &buf)
		w, err := p.GetAs[io.Writer]("buffer")
		tester.AssertEqualErr(t, nil, err)
		w.Write([]byte("test"))
		tester.AssertEqual(t, "test", buf.String())
	})
	t.Run("type conversion failed", func(t *testing.T) {
		p := properties.NewProperties(nil)
		var buf bytes.Buffer
		p.Set("buffer", &buf)
		c, err := p.GetAs[io.Closer]("buffer")
		tester.AssertEqualErr(t, &properties.Error{Type: "convert"}, err)
		tester.AssertEqual(t, nil, c)
	})
	t.Run("key not found", func(t *testing.T) {
		p := properties.NewProperties(nil)
		c, err := p.GetAs[string]("not-exist")
		tester.AssertEqualErr(t, properties.ErrNotFound, err)
		tester.AssertEqual(t, "", c)
	})
	t.Run("nested keys from shallow", func(t *testing.T) {
		p := properties.NewProperties(nil)
		p.Set("foo", "FOO")
		p.Set("foo.bar", "BAR")
		p.Set("foo.bar.baz", "BAZ")
		tester.AssertEqual(t, "FOO", mustGet[string](p, "foo"))
		tester.AssertEqual(t, "BAR", mustGet[string](p, "foo.bar"))
		tester.AssertEqual(t, "BAZ", mustGet[string](p, "foo.bar.baz"))
	})
	t.Run("nested keys from deep", func(t *testing.T) {
		p := properties.NewProperties(nil)
		p.Set("foo.bar.baz", "BAZ")
		p.Set("foo.bar", "BAR")
		p.Set("foo", "FOO")
		tester.AssertEqual(t, "FOO", mustGet[string](p, "foo"))
		tester.AssertEqual(t, "BAR", mustGet[string](p, "foo.bar"))
		tester.AssertEqual(t, "BAZ", mustGet[string](p, "foo.bar.baz"))
	})
	t.Run("nested key not found", func(t *testing.T) {
		p := properties.NewProperties(nil)
		p.Set("foo", "FOO")
		p.Set("foo.bar.baz", "BAZ")
		c, err := p.GetAs[string]("foo.bar")
		tester.AssertEqualErr(t, properties.ErrNotFound, err)
		tester.AssertEqual(t, "", c)
	})
}

func mustGet[T any](p *properties.Properties, key string) T {
	v, err := p.GetAs[T](key)
	if err != nil {
		panic(err)
	}
	return v
}
