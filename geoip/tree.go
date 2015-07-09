package geoip

import (
	"net"
	"strings"
)

type node struct {
	p, l, r *node
	v       Payload
	leaf    bool
}

// Tree is a radix tree links all the records together.
type Tree struct {
	root node
}

// NewTable creates a empty radix tree.
func NewTable() *Tree {
	return &Tree{}
}

// Add append a record to the tree, overwrite controls the behaivor
// when join overlapped IP sets with different Payload. If true, the
// latter wins.
func (t *Tree) Add(r *Record, overwrite bool) {
	prefix := r.i.prefix
	size := r.i.size
	mask := uint32(1 << 31)
	n := &t.root

	// if mod == true, we need try to combine adjacent nodes
	mod := false

	for depth := 1; depth <= size; depth++ {
		msb := (prefix & mask) >> 31
		prefix <<= 1

		var tbranch, obranch **node
		if msb == 0 {
			tbranch = &n.l
			obranch = &n.r
		} else {
			tbranch = &n.r
			obranch = &n.l
		}

		if n.leaf {
			// reach a leaf without the need to go deeper
			if !overwrite || vequal(n.v, r.v) {
				break
			}
			n.leaf = false
			*tbranch = &node{p: n}
			(*tbranch).leaf = true
			if depth == size {
				(*tbranch).v = r.v
			} else {
				// always add new node as a leaf to keep original info
				(*tbranch).v = n.v
			}
			*obranch = &node{p: n, v: n.v, leaf: true}
			n.v = nil
			mod = true
		} else {
			if depth == size {
				// unfinished node, not leaf, but no child
				if overwrite || (*tbranch == nil) {
					*tbranch = &node{p: n, v: r.v, leaf: true}
					mod = true
				} else {
					fill(*tbranch, r.v)
					// deal compression inside fill
					mod = false
				}
			} else {
				if *tbranch == nil {
					*tbranch = &node{p: n}
					mod = true
				}
			}
		}
		n = *tbranch
	}

	if mod {
		compress(n)
	}
}

// Dump prints all the records inside the tree one by one.
func (t *Tree) Dump() string {
	cb := func(r *Record, ud interface{}) {
		arr := ud.(*[]string)
		*arr = append(*arr, r.String())
	}

	var s []string
	t.walk(cb, &s)

	return strings.Join(s, "\n")
}

// Lookup find the associated payload of an IP from the tree. It
// returns the payload and true on success, otherwise, returns false,
// and the payload returned is undefined.
func (t *Tree) Lookup(ip net.IP) (Payload, bool) {
	if t == nil {
		return nil, false
	}

	prefix := ipToNum(ip)
	mask := uint32(1 << 31)
	n := &t.root
	for depth := 1; depth <= 32; depth++ {
		msb := (prefix & mask) >> 31
		prefix <<= 1

		if msb == 0 {
			n = n.l
		} else {
			n = n.r
		}

		if n == nil {
			return nil, false
		}

		if n.leaf {
			return n.v, true
		}
	}
	panic("should not reach here")
}

func (t *Tree) walk(cb func(r *Record, ud interface{}), ud interface{}) {
	var f func(n *node, prefix uint32, depth int)
	f = func(n *node, prefix uint32, depth int) {
		if n.leaf {
			r := &Record{
				i: cidr{
					prefix: prefix << uint(32-depth),
					size:   depth,
				},
				v: n.v,
			}
			cb(r, ud)
		} else {
			prefix <<= 1
			depth++
			if n.l != nil {
				f(n.l, prefix, depth)
			}
			prefix |= 1
			if n.r != nil {
				f(n.r, prefix, depth)
			}
		}
	}

	f(&t.root, 0, 0)
}

func compress(n *node) int64 {
	if !n.leaf {
		panic("must be a leaf")
	}

	var delta int64

	for {
		p := n.p
		if p == nil {
			break
		}
		if p.l != nil && p.r != nil &&
			p.l.leaf && p.r.leaf &&
			vequal(p.l.v, p.r.v) {
			p.leaf = true
			p.v = p.l.v
			p.l = nil
			p.r = nil
			n = p
			delta--
		} else {
			break
		}
	}
	return delta
}

func fill(n *node, v Payload) (int64, bool) {
	f := func(n *node, v Payload) (int64, bool) {
		if n == nil {
			return 0, true
		} else if n.leaf {
			// compare, return merge result
			if n.v == v {
				return -1, true
			} else {
				return 0, false
			}
		} else {
			return fill(n, v)
		}
	}

	dl, cl := f(n.l, v)
	dr, cr := f(n.r, v)

	// n.l == nil && n.r == nil
	// should be handled in previous iteration

	if cl && cr {
		n.v = v
		n.leaf = true
		n.l = nil
		n.r = nil
		return dl + dr, true
	}

	if n.l == nil {
		n.l = &node{p: n, v: v, leaf: true}
		return 1, false
	}

	if n.r == nil {
		n.r = &node{p: n, v: v, leaf: true}
		return 1, false
	}

	return dl + dr, false
}
