package sexp

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

var text1 = `
Lorem ipsum | dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis
nostrud exercitation | ullamco laboris nisi ut aliquip ex ea commodo consequat.
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore
eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt
in culpa qui officia | deserunt | mollit anim id est laborum.`[1:]

var text2 = `
|On the other hand, we denounce with righteous indignation and dislike men who
are so beguiled and demoralized by the charms of pleasure of the moment, so
blinded by desire, that they cannot foresee the pain and trouble that are bound
|to ensue; and equal blame belongs to those who fail in their duty through
weakness of will, which is the same as saying through shrinking from toil and
pain. These cases are perfectly simple and easy to distinguish. In a free
hour, when our power of choice is untrammelled and when nothing prevents our
being able to do what we like best, every pleasure is to be welcomed and every
pain avoided. But in certain circumstances and owing to the claims of duty or
the obligations of business it will frequently occur that pleasures have to
be repudiated and annoyances accepted. The wise man therefore always holds in
these matters to this principle of selection: he rejects pleasures to secure
other greater pleasures, or else he endures pains to avoid worse pains.|`[1:]

func expect_panic(f func(), onrecover func(v interface{})) {
	defer func() {
		onrecover(recover())
	}()
	f()
}

func TestSourceContextForPanics(t *testing.T) {
	var ctx SourceContext
	expect_panic(func() {
		ctx.Decode(0)
	}, func(v interface{}) {
		if v == nil {
			t.Fatal("expected panic")
		}
		t.Logf("%s", v)
	})

	expect_panic(func() {
		ctx.AddFile("1.txt", -1)
		ctx.AddFile("2.txt", 50)
	}, func(v interface{}) {
		if v == nil {
			t.Fatal("expected panic")
		}
		t.Logf("%s", v)
	})

}

func read_file(ctx *SourceContext, filename string, r io.Reader) []SourceLoc {
	f := ctx.AddFile(filename, -1)
	br := bufio.NewReader(r)
	offset := 0
	locs := []SourceLoc(nil)
	for {
		r, s, err := br.ReadRune()
		if err != nil {
			f.Finalize(offset + s)
			return locs
		}
		switch r {
		case '\n':
			offset += s
			f.AddLine(offset)
		case '|':
			locs = append(locs, f.Encode(offset))
			fallthrough
		default:
			offset += s
		}
	}
	panic("unreachable")
}

func TestSourceLocation(t *testing.T) {
	var ctx SourceContext
	var locs []SourceLoc

	locs = append(locs, read_file(&ctx, "1.txt", strings.NewReader(text1))...)
	locs = append(locs, read_file(&ctx, "2.txt", strings.NewReader(text2))...)
	locs = append(locs, read_file(&ctx, "3.txt", strings.NewReader(text1))...)
	locs = append(locs, read_file(&ctx, "4.txt", strings.NewReader(text2))...)
	lengths := [][2]int{
		{ctx.files[0].length, len(text1)},
		{ctx.files[1].length, len(text2)},
		{ctx.files[2].length, len(text1)},
		{ctx.files[3].length, len(text2)},
	}
	for i, l := range lengths {
		if l[0] != l[1] {
			t.Fatalf("[%d] lengths should match, got: %d != %d", i, l[0], l[1])
		}
	}

	if len(locs) != 14 {
		t.Fatalf("there should be 12 '|' characters in the stream")
	}

	goldlocs := []SourceLocEx{
		{"1.txt", 1, 0, 12},
		{"1.txt", 3, 157, 178},
		{"1.txt", 6, 393, 414},
		{"1.txt", 6, 393, 425},
		{"2.txt", 1, 0, 0},
		{"2.txt", 4, 235, 235},
		{"2.txt", 13, 927, 998},
		{"3.txt", 1, 0, 12},
		{"3.txt", 3, 157, 178},
		{"3.txt", 6, 393, 414},
		{"3.txt", 6, 393, 425},
		{"4.txt", 1, 0, 0},
		{"4.txt", 4, 235, 235},
		{"4.txt", 13, 927, 998},
	}
	files := map[string]string{
		"1.txt": text1,
		"2.txt": text2,
		"3.txt": text1,
		"4.txt": text2,
	}

	for i, l := range locs {
		l1 := ctx.Decode(l)
		l2 := goldlocs[i]
		if l1 != l2 {
			t.Errorf("[%d] source locations mismatch: %#v != %#v", i, l1, l2)
		}

		if files[l1.Filename][l1.Offset] != '|' {
			t.Errorf("[%d] source location offset seems incorrect", i)
		}
	}
}
