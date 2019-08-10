// Package swriter generates formatted s-expression output.
package swriter

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

var (
	alphaLower         = "abcdefghijklmnopqrstuvwxyz"
	alphaUpper         = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	num                = "0123456789"
	special            = "._+-/:*%${}"
	allowedStringChars = alphaLower + alphaUpper + num + special
)

// SExpWriter generates output in the form of S-expressions.
type SExpWriter struct {
	writer *bufio.Writer

	indent        int
	needSeparator bool
}

// NewSExpWriter creates a writer object for emitting serialized S-expressions.
func NewSExpWriter(w io.Writer) (*SExpWriter, error) {
	out := SExpWriter{
		writer: bufio.NewWriterSize(w, 1024),
	}
	return &out, nil
}

// StartList starts an s-expression list.
func (w *SExpWriter) StartList(newBlock bool) {
	if newBlock {
		w.indentNewline()
	}
	if w.needSeparator {
		w.writer.WriteRune(' ')
		w.needSeparator = false
	}
	w.writer.WriteRune('(')
	w.indent += 2
}

// Newlines emits the given number of newlines.
func (w *SExpWriter) Newlines(num int) {
	for i := 0; i < num-1; i++ {
		w.writer.WriteRune('\n')
	}
	if num > 0 {
		w.indentNewline()
	}
}

// Separator spaces out the output.
func (w *SExpWriter) Separator() {
	w.writer.WriteRune('\n')
	w.indentNewline()
}

// StringScalar writes a scalar string value to the next position,
// quoting if necessary.
func (w *SExpWriter) StringScalar(in string) {
	if w.needSeparator {
		w.writer.WriteRune(' ')
	}

	if w.needsQuoting(in) {
		w.writer.WriteRune('"')
		w.writer.WriteString(in)
		w.writer.WriteRune('"')
	} else {
		w.writer.WriteString(in)
	}
	w.needSeparator = true
}

// IntScalar writes a int string value to the next position.
func (w *SExpWriter) IntScalar(in int) {
	if w.needSeparator {
		w.writer.WriteRune(' ')
	}
	w.writer.WriteString(fmt.Sprint(in))
	w.needSeparator = true
}

func (w *SExpWriter) needsQuoting(in string) bool {
	if in == "" {
		return true
	}

outer:
	for _, c := range in {
		for _, a := range allowedStringChars {
			if c == a {
				continue outer
			}
		}
		return true
	}
	return false
}

// CloseList closes the outermost list.
func (w *SExpWriter) CloseList(newline bool) error {
	if w.indent <= 0 {
		return errors.New("no open list")
	}

	w.indent--
	if newline {
		w.Newlines(1)
	}
	w.indent--

	w.writer.WriteRune(')')
	return w.writer.Flush()
}

func (w *SExpWriter) indentNewline() {
	if w.indent > 0 {
		w.writer.WriteRune('\n')
		w.needSeparator = false
	}
	for i := 0; i < w.indent-1; i++ {
		w.writer.WriteRune(' ')
	}
	w.needSeparator = true
}
