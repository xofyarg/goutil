package opt

import (
	"fmt"
	"testing"
)

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"":    "",
		"abc": "abc",
		"Abc": "abc",
		"ABC": fmt.Sprintf("a%cb%cc", Sep, Sep),
	}

	for _, c := range Breaker {
		k := fmt.Sprintf("A%cB", c)
		v := fmt.Sprintf("a%cb", Sep)
		cases[k] = v

		k = fmt.Sprintf("A%c%cB", c, c)
		v = fmt.Sprintf("a%cb", Sep)
		cases[k] = v
	}

	for in, exp := range cases {
		if out := normalize(in); out != exp {
			t.Errorf("in:[%s], want:[%s], out:[%s]", in, exp, out)
		}
	}
}
