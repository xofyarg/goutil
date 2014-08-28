package util

import (
	"bytes"
	"testing"
)

func init() {
	SetLevel(Debug)
}

func TestParseComments(t *testing.T) {
	in := `
#a = -1
##b = 1
 #c = true
`
	e := struct {
		A int
		B uint
		c bool
	}{-1, 1, true}
	cp := e

	r := bytes.NewReader([]byte(in))

	if err := LoadConfig(r, &e); err != nil {
		t.Fatalf("parse error: %s", err)
		return
	}

	if cp != e {
		t.Fatalf("parse error! expect %v, but got %v", e, cp)
	}
}

func TestParseName(t *testing.T) {
	in := `
a = -1
b = 1
`
	e := struct {
		A int
		B int
		b uint
	}{-1, 1, 0}
	cp := e

	r := bytes.NewReader([]byte(in))

	if err := LoadConfig(r, &e); err != nil {
		t.Fatalf("parse error: %s", err)
		return
	}

	if cp != e {
		t.Fatalf("parse error! expect %v, but got %v", e, cp)
	}
}
