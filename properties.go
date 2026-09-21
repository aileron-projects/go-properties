package properties

import (
	"reflect"
	"strconv"
	"strings"
)

// NewProperties creates new instance of [Properties].
// If props is not empty, properties will be set to the
// returned Properties.
func NewProperties(props map[string]string) *Properties {
	p := &Properties{
		root: &treeNode{},
	}
	for k, v := range props {
		p.Set(k, v)
	}
	return p
}

// node is the trie tree node.
type treeNode struct {
	value    any
	children map[string]*treeNode
	hasValue bool
}

// Properties provides priperties.
type Properties struct {
	root *treeNode
}

// Set sets the property.
func (p *Properties) Set(key string, value any) {
	node := p.root
	for seg := range strings.SplitSeq(key, ".") {
		if node.children == nil {
			node.children = map[string]*treeNode{}
		}
		if nn, ok := node.children[seg]; ok {
			node = nn
			continue
		}
		newNode := &treeNode{}
		node.children[seg] = newNode
		node = newNode
	}
	node.hasValue = true
	node.value = value
}

// Get returns a property value if exists.
// Get returns [ErrNotFound] if the key was not found.
func (p *Properties) Get(key string) (any, error) {
	node := p.root
	for seg := range strings.SplitSeq(key, ".") {
		if len(node.children) == 0 {
			return nil, ErrNotFound
		}
		node = node.children[seg]
	}
	if !node.hasValue {
		return nil, ErrNotFound
	}
	return node.value, nil
}

// Get returns a property value with the specified type T if exists.
func (p *Properties) GetAs[T any](key string) (T, error) {
	v, err := p.Get(key)
	if err != nil {
		return *new(T), err
	}
	return parseAny[T](v)
}

func parseAny[T any](v any) (T, error) {
	s, ok := v.(string)
	if ok {
		vv, err := parseString(s, reflect.TypeFor[T]().Kind())
		if err != nil {
			from := "from string"
			to := " to " + reflect.TypeFor[T]().String()
			return *new(T), &Error{Inner: err, Type: "convert", Msg: "cannot convert data type." + from + to}
		}
		return vv.(T), nil
	}
	vv, ok := v.(T)
	if !ok {
		from := " from " + reflect.TypeOf(v).String()
		to := " to " + reflect.TypeFor[T]().String()
		return *new(T), &Error{Type: "convert", Msg: "cannot convert data type." + from + to}
	}
	return vv, nil
}

func parseString(v string, kind reflect.Kind) (any, error) {
	switch kind {
	case reflect.String:
		return v, nil
	case reflect.Bool:
		return strconv.ParseBool(v)
	case reflect.Int:
		return strconv.Atoi(v)
	case reflect.Int8:
		vv, err := strconv.ParseInt(v, 10, 8)
		return int8(vv), err
	case reflect.Int16:
		vv, err := strconv.ParseInt(v, 10, 16)
		return int16(vv), err
	case reflect.Int32:
		vv, err := strconv.ParseInt(v, 10, 32)
		return int32(vv), err
	case reflect.Int64:
		vv, err := strconv.ParseInt(v, 10, 64)
		return vv, err
	case reflect.Uint:
		vv, err := strconv.ParseUint(v, 10, 32)
		return uint(vv), err
	case reflect.Uint8:
		vv, err := strconv.ParseUint(v, 10, 8)
		return uint8(vv), err
	case reflect.Uint16:
		vv, err := strconv.ParseUint(v, 10, 16)
		return uint16(vv), err
	case reflect.Uint32:
		vv, err := strconv.ParseUint(v, 10, 32)
		return uint32(vv), err
	case reflect.Uint64:
		vv, err := strconv.ParseUint(v, 10, 64)
		return vv, err
	case reflect.Float32:
		vv, err := strconv.ParseFloat(v, 32)
		return float32(vv), err
	case reflect.Float64:
		vv, err := strconv.ParseFloat(v, 64)
		return vv, err
	case reflect.Complex64:
		vv, err := strconv.ParseComplex(v, 32)
		return complex64(vv), err
	case reflect.Complex128:
		vv, err := strconv.ParseComplex(v, 64)
		return vv, err
	}
	return nil, ErrConvert
}
