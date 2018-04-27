package sexp

import (
	"fmt"
)

func panic_if_error(err error) {
	if err != nil {
		panic(err)
	}
}

func number_suffix(n int) string {
	if n >= 10 && n <= 20 {
		return "th"
	}
	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	}
	return "th"
}

func the_list_has_n_children(n int) string {
	switch n {
	case 1:
		return "the list has 1 child only"
	}
	return fmt.Sprintf("the list has %d children only", n)
}
