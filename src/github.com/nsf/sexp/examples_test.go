package sexp_test

import (
	"fmt"
	"github.com/nsf/sexp"
	"strings"
)

func ExampleNode_Unmarshal() {
	const example_sexp = `
		(position 5   10    4.7)
		(target   10  -2.4  30.3)
	`

	var example struct {
		Pos [3]float32 `sexp:"position,siblings"`
		Tgt [3]float32 `sexp:"target,siblings"`
	}

	ast, err := sexp.Parse(strings.NewReader(example_sexp), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ast.Unmarshal(&example)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(example.Pos)
	fmt.Println(example.Tgt)

	// Output:
	// [5 10 4.7]
	// [10 -2.4 30.3]
}

func ExampleBeautify() {
	const example_sexp = `
		(correct syntax)
		( ; oops, no enclosing ')' here
	`
	var ctx sexp.SourceContext
	f := ctx.AddFile("example.sexp", -1)
	_, err := sexp.Parse(strings.NewReader(example_sexp), f)
	if err != nil {
		// we know the contents of the only source file used, let's
		// just return it:
		getcont := func(string) []byte {
			return []byte(example_sexp)
		}
		fmt.Println(sexp.Beautify(err, getcont, &ctx, false))
	}
	// Output:
	// example.sexp:3:3: error: missing matching sequence delimiter ')'
	// 		( ; oops, no enclosing ')' here
	// 		↑
}
