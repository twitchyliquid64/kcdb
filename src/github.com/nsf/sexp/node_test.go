package sexp

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func must_contain(t *testing.T, err, what string) {
	re := regexp.MustCompile(what)
	if !re.MatchString(err) {
		t.Errorf(`expected: "%s", got: "%s"`, what, err)
	} else {
		t.Logf(`ok: %s`, err)
	}
}

func error_must_contain(t *testing.T, err error, what string) {
	if err == nil {
		t.Errorf("non-nil error expected")
		return
	}
	must_contain(t, err.Error(), what)
}

func test_unmarshal_generic(t *testing.T, data string, f func(*Node, ...interface{}) error, v ...interface{}) {
	root, err := Parse(strings.NewReader(data), nil)
	if err != nil {
		t.Error(err)
		return
	}

	err = f(root, v...)
	if err != nil {
		t.Error(err)
	}
}

func test_unmarshal(t *testing.T, data string, v ...interface{}) {
	test_unmarshal_generic(t, data, (*Node).Unmarshal, v...)
}

func test_unmarshal_children(t *testing.T, data string, v ...interface{}) {
	test_unmarshal_generic(t, data, (*Node).UnmarshalChildren, v...)
}

const countries = `
;; a list of arbitrary countries
(countries (
	Spain
	Russia ; I live here :-D
	Japan
	China
	England
	Germany
	France
	Sweden
	Iraq
	Iran
	Indonesia
	India
	USA
	Canada
	Brazil
))
`

// just to test Unmarshaler interface
type smiley string

func (s *smiley) UnmarshalSexp(n *Node) error {
	if !n.IsScalar() {
		return NewUnmarshalError(n, reflect.TypeOf(s),
			"scalar value required")
	}
	*s = smiley(n.Value + " :-D")
	return nil
}

// always fails
type neversmiley string

func (s *neversmiley) UnmarshalSexp(n *Node) error {
	return NewUnmarshalError(n, reflect.TypeOf(s), ":-( Y U NO HAPPY?")
}

type neversmiley2 string

func (s *neversmiley2) UnmarshalSexp(n *Node) error {
	return errors.New("inevitable failure")
}

func TestUnmarshal(t *testing.T) {
	var a [3]int8
	test_unmarshal(t, "5 10 -15", &a)
	t.Logf("%d %d %d", a[0], a[1], a[2])

	var b [3]uint16
	test_unmarshal(t, "1024 750 300", &b)
	t.Logf("%d %d %d", b[0], b[1], b[2])

	var m map[string][]string
	test_unmarshal(t, countries, &m)
	for _, country := range m["countries"] {
		t.Logf("%q", country)
	}

	var s []smiley
	test_unmarshal(t, `what if we try`, &s)
	for _, s := range s {
		t.Logf("%q", s)
	}
}

func test_unmarshal_error(t *testing.T, source, what string, args ...interface{}) {
	ast, err := Parse(strings.NewReader(source), nil)
	if err != nil {
		t.Error(err)
	}
	err = ast.Unmarshal(args...)
	error_must_contain(t, err, what)
}

func TestUnmarshalErrors(t *testing.T) {
	var (
		a [3]uint8
		b neversmiley
		c [1]neversmiley2
		d [][]string
		e []bool
		f map[string]int
		g chan int
		h [3]int8
		i **int
		j [1]float64
		k interface {
			String() string
		}
	)

	expect_panic(func() {
		test_unmarshal_error(t, "1 2 3", "", a)
	}, func(v interface{}) {
		if s, ok := v.(string); ok {
			must_contain(t, s, "Node.Unmarshal expects a non-nil pointer")
		} else {
			t.Errorf("unexpected panic: %s", v)
		}
	})
	test_unmarshal_error(t, "123", "Y U NO HAPPY", &b)
	test_unmarshal_error(t, "123", "inevitable failure", &c)
	test_unmarshal_error(t, "123", "list value required", &d)
	test_unmarshal_error(t, "(true (1 2 3) false)", "scalar value required", &e)
	test_unmarshal_error(t, "256", "integer overflow", &a)
	test_unmarshal_error(t, "abc", "invalid syntax", &a)
	test_unmarshal_error(t, "trUe", "undefined boolean", &e)
	test_unmarshal_error(t, "(a 1) (b 2) (c 3) oops", "node is not a list", &f)
	test_unmarshal_error(t, "hello", "unsupported type", &g)
	test_unmarshal_error(t, "-129", "integer overflow", &h)
	test_unmarshal_error(t, "abc", "invalid syntax", &h)
	test_unmarshal_error(t, "11", "unsupported type", &i) // only one level of indirection
	test_unmarshal_error(t, "3.1415f", "invalid syntax", &j)
	test_unmarshal_error(t, "xxx", "unsupported type", &k)
}

func TestNodeNth(t *testing.T) {
	root, err := Parse(strings.NewReader("0 1 2 3"), nil)
	if err != nil {
		t.Error(err)
		return
	}

	assert := func(cond bool, msg interface{}) {
		if !cond {
			t.Error(msg)
		}
	}
	n, err := root.Nth(0)
	assert(err == nil, err)
	assert(n != nil, "non-nil node expected")
	assert(n.Value == "0", "0 expected")

	_, err = n.Nth(234)
	error_must_contain(t, err, "is not a list")

	_, err = root.Nth(3)
	assert(err == nil, err)

	_, err = root.Nth(4)
	error_must_contain(t, err, "cannot retrieve 5th")
}

func TestNodeIterKeyValues(t *testing.T) {
	root, err := Parse(strings.NewReader(`
		(
			(x y)
			(z w)
			(a 1)
			(b 2)
			(c 3)
		)
		(
			(x y)
			(z)
		)
		(
			not-a-list
		)
	`), nil)
	if err != nil {
		t.Error(err)
		return
	}
	nth_must := func(n *Node, err error) *Node {
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	list1 := nth_must(root.Nth(0))
	list2 := nth_must(root.Nth(1))
	list3 := nth_must(root.Nth(2))

	items := []struct{ key, value string }{
		{"x", "y"},
		{"z", "w"},
		{"a", "1"},
		{"b", "2"},
		{"c", "3"},
	}
	i := 0
	err = list1.IterKeyValues(func(k, v *Node) error {
		if k.Value != items[i].key {
			t.Errorf("%q != %q", k.Value, items[i].key)
		}
		if v.Value != items[i].value {
			t.Errorf("%q != %q", v.Value, items[i].value)
		}
		i++
		return nil
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = list2.IterKeyValues(func(k, v *Node) error { return nil })
	error_must_contain(t, err, "cannot retrieve 2nd")

	err = list3.IterKeyValues(func(k, v *Node) error { return nil })
	error_must_contain(t, err, "node is not a list")
}
