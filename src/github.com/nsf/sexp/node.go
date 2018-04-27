package sexp

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// The main and only AST structure. All fields are self explanatory, however
// the way they are being formed needs explanation.
//
// A list node has empty value and non-nil children pointer, which is a
// nil-terminated list of children nodes.
//
// A scalar node has nil children pointer.
//
// Take a look at this example:
//
//   ((1 2) 3 4)
//
// will yield:
//
//   Node{Children:
//     Node{Children:
//       Node{Value: "1", Next:
//       Node{Value: "2"}}, Next:
//     Node{Value: "3", Next:
//     Node{Value: "4"}}}}
type Node struct {
	Location SourceLoc
	Value    string
	Children *Node
	Next     *Node
}

// Returns true if the node is a list (has children).
func (n *Node) IsList() bool {
	return n.Children != nil
}

// Return true if the node is a scalar (has no children).
func (n *Node) IsScalar() bool {
	return n.Children == nil
}

func (n *Node) String() string {
	return n.Value
}

// Returns the number of children nodes. Has O(N) complexity.
func (n *Node) NumChildren() int {
	i := 0
	c := n.Children
	for c != nil {
		i++
		c = c.Next
	}
	return i
}

// Returns Nth child node. If node is not a list, it will return an error.
func (n *Node) Nth(num int) (*Node, error) {
	if !n.IsList() {
		return nil, NewUnmarshalError(n, nil, "node is not a list")
	}

	i := 0
	for c := n.Children; c != nil; c = c.Next {
		if i == num {
			return c, nil
		}
		i++
	}

	num++
	return nil, NewUnmarshalError(n, nil,
		"cannot retrieve %d%s child node, %s",
		num, number_suffix(num),
		the_list_has_n_children(n.NumChildren()))
}

// Walk over children nodes, assuming they are key/value pairs. It returns error
// if the iterable node is not a list or if any of its children is not a
// key/value pair.
func (n *Node) IterKeyValues(f func(k, v *Node) error) error {
	for c := n.Children; c != nil; c = c.Next {
		if !c.IsList() {
			return NewUnmarshalError(c, nil,
				"node is not a list, expected key/value pair")
		}
		// don't check for error here, because it's obvious that if the
		// node is a list (and the definition of the list is `Children
		// != nil`), it has at least one child
		k, _ := c.Nth(0)
		v, err := c.Nth(1)
		if err != nil {
			return err
		}
		err = f(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

type Unmarshaler interface {
	UnmarshalSexp(n *Node) error
}

// Unmarshals all children nodes of the node to pointer values. Applies the
// same logic as Unmarshal. See description of the (*Node).Unmarshal method for
// more details.
func (n *Node) UnmarshalChildren(vals ...interface{}) (err error) {
	if len(vals) == 0 {
		return nil
	}

	// unmarshal all children of the node
	i := 0
	for c := n.Children; c != nil; c = c.Next {
		if i >= len(vals) {
			break
		}
		if vals[i] == nil {
			i++
			continue
		}
		if err := c.unmarshal(vals[i]); err != nil {
			return err
		}
		i++
	}

	// did we fullfil all the arguments?
	if i < len(vals) {
		if i == 0 {
			return NewUnmarshalError(n, nil,
				"node has no children, %d was requested",
				len(vals))
		}
		return NewUnmarshalError(n, nil,
			"node has %d children, %d was requested",
			i, len(vals))
	}

	return nil
}

// Unmarshals the node and its siblings to pointer values.
//
// The function expects pointers to values with arbitrary types. If one of the
// arguments is not a pointer it will panic.
//
// It supports unmarshaling to the following types:
//  - all number types: int{8,16,32,64}, uint{8,16,32,64}, float{32,64}
//  - bool
//  - string
//  - arrays and slices of all supported types
//  - empty interfaces{}
//  - maps
//  - structs
//  - pointers to any of the supported types (only one level of indirection)
//  - any type which implements Unmarshaler
//
// Here's some details on unmarshaling semantics:
//  (u)ints: unmarshaled using strconv.ParseInt/strconv.ParseUint with base 10
//           only
//  floats:  unmarshaled using strconv.ParseFloat
//  bool:    works strictly on two values "true" or "false"
//  string:  unmarshaled as is (keep in mind that lexer supports escape sequences)
//  arrays:  uses up to len(array) elements, if there is a smaller amount of
//           elements, the rest is zeroed
//  slices:  uses all elements appending them to the slice, however if the slice
//           was bigger than the amount of elements, it will reslice it to the
//           appropriate length
//  iface:   only empty interfaces are supported, it will unmarshal AST to
//           a []interface{} or a string
//  map:     when unmarshaling to the map, assumes this AST form:
//           `((key value) (key value) (key value))`, doesn't clear the map
//           before appending all the key value pairs
//  struct:  uses the same AST form as the map, where `key` means `field`,
//           supports `sexp` tags (see description below), will try to match
//           name specified in the tag, the field name and the field name
//           ignoring the case in that order
//
// Struct tags have the form: "name,opt,opt". Special tag "-" means "skip me".
// Supported options:
//  siblings: will use sibling nodes instead of children for unmarshaling
//            to an array or a slice.
//
// Important note: If the type implements Unmarshaler interface, it will use it
// instead of applying default unmarshaling strategies described above.
func (n *Node) Unmarshal(vals ...interface{}) (err error) {
	if len(vals) == 0 {
		return nil
	}

	// unmarshal the node itself
	if vals[0] != nil {
		if err := n.unmarshal(vals[0]); err != nil {
			return err
		}
	}

	// unmarshal node's siblings
	i := 1
	for s := n.Next; s != nil; s = s.Next {
		if i >= len(vals) {
			break
		}
		if vals[i] == nil {
			i++
			continue
		}
		if err := s.unmarshal(vals[i]); err != nil {
			return err
		}
		i++
	}

	// did we fullfil all the arguments?
	if i < len(vals) {
		if i == 1 {
			return NewUnmarshalError(n, nil,
				"node has no siblings, %d was requested",
				len(vals)-1)
		}
		return NewUnmarshalError(n, nil,
			"node has %d siblings, %d was requested",
			i-1, len(vals)-1)
	}

	return nil
}

type UnmarshalError struct {
	Type    reflect.Type
	Node    *Node
	message string
}

func NewUnmarshalError(n *Node, t reflect.Type, format string, args ...interface{}) *UnmarshalError {
	if len(args) == 0 {
		// simple hack to make it a bit faster in the case when no args
		// were provided
		return &UnmarshalError{
			Type:    t,
			Node:    n,
			message: format,
		}
	}
	return &UnmarshalError{
		Type:    t,
		Node:    n,
		message: fmt.Sprintf(format, args...),
	}
}

func (e *UnmarshalError) Error() string {
	args := []interface{}{e.message}
	format := "%s"
	if e.Node != nil {
		if e.Node.IsList() {
			format += " (list value)"
		} else {
			format += " (value: %q)"
			args = append(args, e.Node.Value)
		}
	}
	if e.Type != nil {
		format += " (type: %s)"
		args = append(args, e.Type)
	}

	return fmt.Sprintf(format, args...)
}

func (n *Node) unmarshal_error(t reflect.Type, format string, args ...interface{}) {
	panic(NewUnmarshalError(n, t, fmt.Sprintf(format, args...)))
}

func (n *Node) unmarshal_unmarshaler(v reflect.Value) bool {
	u, ok := v.Interface().(Unmarshaler)
	if !ok {
		// T doesn't work, try *T as well
		if v.Kind() != reflect.Ptr && v.CanAddr() {
			u, ok = v.Addr().Interface().(Unmarshaler)
			if ok {
				v = v.Addr()
			}
		}
	}
	if ok && (v.Kind() != reflect.Ptr || !v.IsNil()) {
		err := u.UnmarshalSexp(n)
		if err != nil {
			if ue, ok := err.(*UnmarshalError); ok {
				panic(ue)
			}
			n.unmarshal_error(v.Type(), err.Error())
		}
		return true
	}
	return false
}

func (n *Node) ensure_scalar(t reflect.Type) {
	if n.IsScalar() {
		return
	}

	n.unmarshal_error(t, "scalar value required")
}

func (n *Node) ensure_list(t reflect.Type) {
	if n.IsList() {
		return
	}

	n.unmarshal_error(t, "list value required")
}

func (n *Node) unmarshal_value(v reflect.Value, use_siblings bool) {
	t := v.Type()
	// we support one level of indirection at the moment
	if v.Kind() == reflect.Ptr {
		// if the pointer is nil, allocate a new element of the type it
		// points to
		if v.IsNil() {
			v.Set(reflect.New(t.Elem()))
		}
		v = v.Elem()
	}

	// try Unmarshaler interface
	if n.unmarshal_unmarshaler(v) {
		return
	}

	// fallback to default unmarshaling scheme
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// TODO: more string -> int conversion options (hex, binary, octal, etc.)
		n.ensure_scalar(t)
		num, err := strconv.ParseInt(n.Value, 10, 64)
		if err != nil {
			n.unmarshal_error(t, err.Error())
		}
		if v.OverflowInt(num) {
			n.unmarshal_error(t, "integer overflow")
		}
		v.SetInt(num)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// TODO: more string -> int conversion options (hex, binary, octal, etc.)
		n.ensure_scalar(t)
		num, err := strconv.ParseUint(n.Value, 10, 64)
		if err != nil {
			n.unmarshal_error(t, err.Error())
		}
		if v.OverflowUint(num) {
			n.unmarshal_error(t, "integer overflow")
		}
		v.SetUint(num)
	case reflect.Float32, reflect.Float64:
		n.ensure_scalar(t)
		num, err := strconv.ParseFloat(n.Value, 64)
		if err != nil {
			n.unmarshal_error(t, err.Error())
		}
		v.SetFloat(num)
	case reflect.Bool:
		n.ensure_scalar(t)
		switch n.Value {
		case "true":
			v.SetBool(true)
		case "false":
			v.SetBool(false)
		default:
			n.unmarshal_error(t, "undefined boolean value, use true|false")
		}
	case reflect.String:
		n.ensure_scalar(t)
		v.SetString(n.Value)
	case reflect.Array, reflect.Slice:
		if !use_siblings {
			n.ensure_list(t)
		}
		i := 0
		c := n.Children
		if use_siblings {
			c = n
		}
		for ; c != nil; c = c.Next {
			if i >= v.Len() {
				if v.Kind() == reflect.Array {
					break
				} else {
					v.Set(reflect.Append(v, reflect.Zero(t.Elem())))
				}
			}

			c.unmarshal_value(v.Index(i), false)
			i++
		}

		if i < v.Len() {
			if v.Kind() == reflect.Array {
				z := reflect.Zero(t.Elem())
				for n := v.Len(); i < n; i++ {
					v.Index(i).Set(z)
				}
			} else {
				v.SetLen(i)
			}
		}
	case reflect.Interface:
		if v.NumMethod() != 0 {
			n.unmarshal_error(t, "unsupported type")
		}

		v.Set(reflect.ValueOf(n.unmarshal_as_interface()))
	case reflect.Map:
		n.ensure_list(t)
		if v.IsNil() {
			v.Set(reflect.MakeMap(t))
		}

		keyv := reflect.New(t.Key()).Elem()
		valv := reflect.New(t.Elem()).Elem()
		err := n.IterKeyValues(func(key, val *Node) error {
			key.unmarshal_value(keyv, false)
			val.unmarshal_value(valv, false)
			v.SetMapIndex(keyv, valv)
			return nil
		})
		if err != nil {
			n.unmarshal_error(t, "%s", err)
		}
	case reflect.Struct:
		err := n.IterKeyValues(func(key, val *Node) error {
			var f reflect.StructField
			var ok bool
			var opts tag_options
			for i, n := 0, t.NumField(); i < n; i++ {
				var tagname string
				f = t.Field(i)
				tag := f.Tag.Get("sexp")
				if tag == "-" {
					continue
				}
				if f.Anonymous {
					continue
				}
				tagname, opts = parse_tag(tag)

				ok = tagname == key.Value
				if ok {
					break
				}
				ok = f.Name == key.Value
				if ok {
					break
				}
				ok = strings.EqualFold(f.Name, key.Value)
				if ok {
					break
				}
			}
			if ok {
				if f.PkgPath != "" {
					n.unmarshal_error(t, "writing to an unexported field")
				} else {
					v := v.FieldByIndex(f.Index)
					val.unmarshal_value(v, opts.contains("siblings"))
				}
			}
			return nil
		})
		if err != nil {
			n.unmarshal_error(t, "%s", err)
		}
	default:
		n.unmarshal_error(t, "unsupported type")
	}
}

func (n *Node) unmarshal_as_interface() interface{} {
	// interface parsing for sexp isn't really useful, the outcome is
	// []interface{} or string
	if n.IsList() {
		var s []interface{}
		for c := n.Children; c != nil; c = c.Next {
			s = append(s, c.unmarshal_as_interface())
		}
		return s
	}
	return n.Value
}

func (n *Node) unmarshal(v interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if _, ok := e.(*UnmarshalError); ok {
				err = e.(error)
			} else {
				panic(e)
			}
		}
	}()

	pv := reflect.ValueOf(v)
	if pv.Kind() != reflect.Ptr || pv.IsNil() {
		panic("Node.Unmarshal expects a non-nil pointer argument")
	}
	n.unmarshal_value(pv.Elem(), false)
	return nil
}
