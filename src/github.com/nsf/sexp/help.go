package sexp

func DontPanic(f func() error) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if ue, ok := e.(*UnmarshalError); ok {
				err = ue
				return
			}
			panic(e)
		}
	}()
	return f()
}

// A simple helper structure inspired by the simplejson-go API. Use Help
// function to actually acquire it from the given *Node.
type Helper struct {
	node *Node
	err *UnmarshalError
}

func Help(node *Node) Helper {
	if node == nil {
		err := NewUnmarshalError(nil, nil, "nil node")
		return Helper{nil, err}
	}
	return Helper{node, nil}
}

func (h Helper) IsValid() bool {
	return h.node != nil
}

func (h Helper) Next() Helper {
	if h.node == nil {
		return h
	}
	if h.node.Next == nil {
		err := NewUnmarshalError(h.node, nil,
			"a sibling of the node was requested, but it has none")
		return Helper{nil, err}
	}
	return Helper{h.node.Next, nil}
}

func (h Helper) Child(n int) Helper {
	if h.node == nil {
		return h
	}
	c := h.node.Children
	if c == nil {
		err := NewUnmarshalError(h.node, nil,
			"cannot retrieve %d%s child node, node is not a list",
			n+1, number_suffix(n+1))
		return Helper{nil, err}
	}
	for i := 0; i < n; i++ {
		c = c.Next
		if c == nil {
			err := NewUnmarshalError(h.node, nil,
				"cannot retrieve %d%s child node, %s",
				n+1, number_suffix(n+1),
				the_list_has_n_children(h.node.NumChildren()))
			return Helper{nil, err}
		}
	}
	return Helper{c, nil}
}

func (h Helper) IsList() bool {
	if h.node == nil {
		return false
	}
	return h.node.IsList()
}

func (h Helper) IsScalar() bool {
	if h.node == nil {
		return false
	}
	return h.node.IsScalar()
}

func (h Helper) Bool() (bool, error) {
	if h.node == nil {
		return false, h.err
	}
	var v bool
	err := h.node.Unmarshal(&v)
	if err != nil {
		return false, err
	}
	return v, nil
}

func (h Helper) Int() (int, error) {
	if h.node == nil {
		return 0, h.err
	}
	var v int
	err := h.node.Unmarshal(&v)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (h Helper) Float64() (float64, error) {
	if h.node == nil {
		return 0, h.err
	}
	var v float64
	err := h.node.Unmarshal(&v)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (h Helper) String() (string, error) {
	if h.node == nil {
		return "", h.err
	}
	var v string
	err := h.node.Unmarshal(&v)
	if err != nil {
		return "", err
	}
	return v, nil
}

func (h Helper) Node() (*Node, error) {
	if h.node == nil {
		return nil, h.err
	}
	return h.node, nil
}

func (h Helper) MustBool() bool {
	v, err := h.Bool()
	if err != nil {
		panic(err)
	}
	return v
}

func (h Helper) MustInt() int {
	v, err := h.Int()
	if err != nil {
		panic(err)
	}
	return v
}

func (h Helper) MustFloat64() float64 {
	v, err := h.Float64()
	if err != nil {
		panic(err)
	}
	return v
}

func (h Helper) MustString() string {
	v, err := h.String()
	if err != nil {
		panic(err)
	}
	return v
}

func (h Helper) MustNode() *Node {
	v, err := h.Node()
	if err != nil {
		panic(err)
	}
	return v
}
