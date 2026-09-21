package properties

import (
	"io"
	"maps"
	"os"
)

// Load loads properties from files.
// Duplicated keys are always overwritten.
// See [Parse] for file formats and [Resolve] for file syntax.
func Load(files ...string) (map[string]string, error) {
	readers := make([]io.Reader, 0, len(files))
	for _, f := range files {
		r, err := os.Open(f)
		if err != nil {
			return nil, &Error{Inner: err, Type: "load", Msg: f}
		}
		readers = append(readers, r)
		defer r.Close()
	}
	return LoadReaders(readers...)
}

// LoadReaders loads envirproperties from files.
// Duplicated keys are always overwritten.
func LoadReaders(readers ...io.Reader) (map[string]string, error) {
	m := map[string]string{}
	for _, r := range readers {
		props, err := parse(r)
		if err != nil {
			return nil, err
		}
		maps.Copy(m, props)
	}
	return m, nil
}
