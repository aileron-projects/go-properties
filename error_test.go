package properties

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/aileron-projects/go-tester"
)

func TestError(t *testing.T) {
	t.Parallel()
	t.Run("unwrap", func(t *testing.T) {
		err := &Error{Inner: io.EOF}
		inner := err.Unwrap()
		tester.AssertEqualErr(t, io.EOF, inner)
	})
	t.Run("error message", func(t *testing.T) {
		err := &Error{Inner: io.EOF, Type: "type", Msg: "msg"}
		msg := err.Error()
		tester.AssertEqual(t, "go-properties/properties: type: msg [EOF]", msg)
	})
	t.Run("empty message", func(t *testing.T) {
		err := &Error{Inner: io.EOF, Type: "type", Msg: ""}
		msg := err.Error()
		tester.AssertEqual(t, "go-properties/properties: type: [EOF]", msg)
	})
	t.Run("nil error", func(t *testing.T) {
		var err *Error
		tester.AssertEqual(t, false, err.Is(nil))
	})
	t.Run("nil target", func(t *testing.T) {
		err := &Error{Type: "type", Msg: "aaa", Inner: nil}
		tester.AssertEqual(t, false, err.Is(nil))
	})
	t.Run("errors equal", func(t *testing.T) {
		target := &Error{Type: "type", Msg: "aaa", Inner: nil}
		err := &Error{Type: "type", Msg: "bbb", Inner: io.EOF}
		tester.AssertEqual(t, true, errors.Is(err, target))
	})
	t.Run("errors not equal", func(t *testing.T) {
		target := &Error{Type: "foo"}
		err := &Error{Type: "bar"}
		tester.AssertEqual(t, false, errors.Is(err, target))
	})
	t.Run("wrapped error equal", func(t *testing.T) {
		target := &Error{Type: "type"}
		inner := &Error{Type: "type"}
		err := fmt.Errorf("outer error [%w]", inner)
		tester.AssertEqual(t, true, errors.Is(err, target))
	})
	t.Run("wrapped error not equal", func(t *testing.T) {
		target := &Error{Type: "type"}
		err := fmt.Errorf("outer error [%w]", io.EOF)
		tester.AssertEqual(t, false, errors.Is(err, target))
	})
	t.Run("wrapped errors equal", func(t *testing.T) {
		target := &Error{Type: "type"}
		inner := &Error{Type: "type"}
		err := fmt.Errorf("outer error [%w] [%w]", io.EOF, inner)
		tester.AssertEqual(t, true, errors.Is(err, target))
	})
	t.Run("wrapped error not equal", func(t *testing.T) {
		target := &Error{Type: "type"}
		err := fmt.Errorf("outer error [%w] [%w]", io.EOF, io.ErrUnexpectedEOF)
		tester.AssertEqual(t, false, errors.Is(err, target))
	})
}
