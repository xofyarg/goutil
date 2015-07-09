package geoip

import (
	"net"
	"strings"
	"testing"
)

type rcase struct {
	i []string
	o string
}

func doRecordFromRangeTest(t *testing.T, c rcase) {
	rs := NewRecordFromRange(net.ParseIP(c.i[0]), net.ParseIP(c.i[1]), nil)
	var l []string
	for _, r := range rs {
		l = append(l, r.String())
	}
	result := strings.Join(l, "\n")

	if result != c.o {
		t.Errorf("want: [%s]\nget: [%s]", c.o, result)
	}
}

func TestRecordFromRange1(t *testing.T) {
	doRecordFromRangeTest(t,
		rcase{
			[]string{"192.168.0.2", "192.168.0.10"},
			"192.168.0.2/31 (-)\n192.168.0.4/30 (-)\n192.168.0.8/31 (-)\n192.168.0.10/32 (-)",
		},
	)
}

type ps string

func (p ps) Equal(t Payload) bool {
	if s, ok := t.(ps); !ok {
		return false
	} else {
		return p == s
	}
}

func (p ps) String() string {
	return string(p)
}

type input struct {
	cidr      string
	payload   ps
	overwrite bool
}
type acase struct {
	i []input
	o string
}

func doAddTest(t *testing.T, c acase) {
	ta := NewTable()
	for _, in := range c.i {
		_, cidr, _ := net.ParseCIDR(in.cidr)
		r := NewRecordFromCIDR(cidr, in.payload)
		ta.Add(r, in.overwrite)
	}
	result := ta.Dump()
	if result != c.o {
		t.Errorf("want: [%s]\nget: [%s]", c.o, result)
	}
}

func TestAddSep(t *testing.T) {
	doAddTest(t,
		acase{
			[]input{
				input{"1.0.0.0/30", "A", false},
				input{"1.0.0.16/30", "A", false},
			},
			"1.0.0.0/30 (A)\n1.0.0.16/30 (A)",
		},
	)
}

func TestAddFirst(t *testing.T) {
	doAddTest(t,
		acase{
			[]input{
				input{"1.0.0.8/30", "A", false},
				input{"1.0.0.16/30", "A", false},
				input{"1.0.0.0/30", "A", false},
			},
			"1.0.0.0/30 (A)\n1.0.0.8/30 (A)\n1.0.0.16/30 (A)",
		},
	)
}

func TestAddLast(t *testing.T) {
	doAddTest(t,
		acase{
			[]input{
				input{"1.0.0.0/30", "A", false},
				input{"1.0.0.8/30", "A", false},
				input{"1.0.0.16/30", "A", false},
			},
			"1.0.0.0/30 (A)\n1.0.0.8/30 (A)\n1.0.0.16/30 (A)",
		},
	)
}

func TestAddCombineOnce(t *testing.T) {
	doAddTest(t,
		acase{
			[]input{
				input{"1.0.0.0/29", "A", false},
				input{"1.0.0.8/29", "A", false},
			},
			"1.0.0.0/28 (A)",
		},
	)
}

func TestAddCombineTwice(t *testing.T) {
	doAddTest(t,
		acase{
			[]input{
				input{"1.0.0.0/29", "A", false},
				input{"1.0.0.8/29", "A", false},
				input{"1.0.0.16/28", "A", false},
			},
			"1.0.0.0/27 (A)",
		},
	)
}

func TestAddMergeEqual(t *testing.T) {
	doAddTest(t,
		acase{
			[]input{
				input{"1.0.0.0/29", "A", false},
				input{"1.0.0.8/29", "A", false},
				input{"1.0.0.0/29", "A", false},
			},
			"1.0.0.0/28 (A)",
		},
	)
}

func TestAddInsert(t *testing.T) {
	doAddTest(t,
		acase{
			[]input{
				input{"1.0.0.0/29", "A", false},
				input{"1.0.0.8/29", "A", false},
				input{"1.0.0.4/31", "A", false},
			},
			"1.0.0.0/28 (A)",
		},
	)
}

//                o
//        o               o
//    .       o       o       .
//  .   .   .   .   .   .   .   .
// . . . . . . . . . . . . . . . .
func TestAddCover(t *testing.T) {
	doAddTest(t,
		acase{
			[]input{
				input{"1.0.0.4/30", "A", false},
				input{"1.0.0.8/30", "A", false},
				input{"1.0.0.0/28", "A", false},
			},
			"1.0.0.0/28 (A)",
		},
	)
}

func TestAddMergeInto(t *testing.T) {
	doAddTest(t,
		acase{
			[]input{
				input{"1.0.0.4/30", "A", false},
				input{"1.0.0.8/30", "A", false},
				input{"1.0.0.4/32", "B", true},
			},
			"1.0.0.4/32 (B)\n1.0.0.5/32 (A)\n1.0.0.6/31 (A)\n1.0.0.8/30 (A)",
		},
	)
}

func TestAddMergeFrom(t *testing.T) {
	doAddTest(t,
		acase{
			[]input{
				input{"1.0.0.4/30", "A", false},
				input{"1.0.0.8/30", "A", false},
				input{"1.0.0.0/29", "B", true},
			},
			"1.0.0.0/29 (B)\n1.0.0.8/30 (A)",
		},
	)
}
