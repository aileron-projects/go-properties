package properties

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrConvert  = errors.New("cannot convert data type from string")
)

// Error is the property related error.
type Error struct {
	Inner error  // Inner is the inner error.
	Type  string // Type is the error type.
	Msg   string // Msg is the error message.
}

func (e *Error) Unwrap() error {
	return e.Inner
}

func (e *Error) Error() string {
	s := "go-properties/properties: " + e.Type + ":"
	if e.Msg != "" {
		s += " " + e.Msg
	}
	if e.Inner != nil {
		s = s + " [" + e.Inner.Error() + "]"
	}
	return s
}

func (e *Error) Is(target error) bool {
	ee, ok := target.(*Error)
	if ok {
		return e.Type == ee.Type
	}
	return false
}
