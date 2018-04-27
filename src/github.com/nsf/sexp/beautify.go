package sexp

import (
	"bytes"
	"fmt"
	"unicode/utf8"
)

const (
	color_white_bold = "\033[1;37m"
	color_red_bold   = "\033[1;31m"
	color_green_bold = "\033[1;32m"
	color_none       = "\033[0m"
)

// Returns a prettified version of the `err`.
//
// The arguments need explanation, if you get a parser error (returned by one
// of the Parse functions) or unmarshaling error (returned by one of the Node
// methods), it can be prettified given that you have access to the
// SourceContext used in parsing and the source data.
//
// You need to provide a closure `getcont` which will return contents of the
// source file for a given filename argument. The reason for this complicated
// interface is because SourceContext supports multiple files and it's not
// necessarly clear where the error is.
//
// Colors argument specified whether you want to use colors or not. It applies
// typical terminal escape sequnces to the resulting string in case if the
// argument is true.
//
// It will prettify only ParseError or UnmarshalError errors, if something else
// is given it will return error.Error() output.
func Beautify(err error, getcont func(string) []byte, ctx *SourceContext, colors bool) string {
	var loc SourceLoc
	switch e := err.(type) {
	case *ParseError:
		loc = e.Location
	case *UnmarshalError:
		loc = e.Node.Location
	default:
		return e.Error()
	}

	locex := ctx.Decode(loc)
	contents := getcont(locex.Filename)
	col := utf8.RuneCount(contents[locex.LineOffset:locex.Offset]) + 1

	linecont := contents[locex.LineOffset:]
	end := bytes.Index(linecont, []byte("\n"))
	if end != -1 {
		linecont = linecont[:end]
	}

	var buf bytes.Buffer
	if colors {
		fmt.Fprintf(&buf, "%s%s:%d:%d: %serror: %s%s%s\n",
			color_white_bold, locex.Filename, locex.Line, col, color_red_bold,
			color_white_bold, err, color_none)
	} else {
		fmt.Fprintf(&buf, "%s:%d:%d: error: %s\n",
			locex.Filename, locex.Line, col, err)
	}
	fmt.Fprintf(&buf, "%s\n", linecont)
	for i := locex.LineOffset; i < locex.Offset; {
		r, size := utf8.DecodeRune(linecont)
		linecont = linecont[size:]
		i += size

		if r == '\t' {
			buf.WriteByte('\t')
		} else {
			buf.WriteByte(' ')
		}
	}
	if colors {
		buf.WriteString(color_green_bold)
	}
	buf.WriteString("↑")
	if colors {
		buf.WriteString(color_none)
	}
	return buf.String()
}
