// Data storage for common geoip database, can be used to merge different databases.
//
// Basic usage:
//   1. create a empty table: NewTable
//   2. convert current database into records: NewRecordFrom{CIDR,Range}
//   3. add these records into table: table.Add
//   4. future operation: table.Lookup/table.Dump
package geoip

import (
	"fmt"
	"net"
)

// use cidr internally instead of IPNet for speed
type cidr struct {
	prefix uint32
	size   int
}

func (c *cidr) String() string {
	return fmt.Sprintf("%d.%d.%d.%d/%d",
		uint8(c.prefix>>24),
		uint8(c.prefix>>16),
		uint8(c.prefix>>8),
		uint8(c.prefix),
		c.size)
}

// Payload is an abstract data structer bound to records, used to
// store geo information of a set of IPs.
type Payload interface {
	Equal(Payload) bool
	String() string
}

// Record is used to store a CIDR and its associated payload.
type Record struct {
	i cidr
	v Payload
}

// String will convert internal ip set into a cidr, and then call
// payload's string method.
func (r *Record) String() string {
	if _, ok := r.v.(fmt.Stringer); ok {
		return fmt.Sprintf("%s (%s)", &r.i, r.v)
	}
	return fmt.Sprintf("%s (-)", &r.i)
}

// NewRecordFromCIDR convert an IPNet structer into a Record.
func NewRecordFromCIDR(i *net.IPNet, v Payload) *Record {
	size, _ := i.Mask.Size()
	prefix := ipToNum(i.IP) & sizeToMask(size)

	return &Record{
		i: cidr{
			prefix: prefix,
			size:   size,
		},
		v: v,
	}
}

// NewRecordFromRange parse a range of IP addresses and convert them
// into a slice of Record.
func NewRecordFromRange(a, b net.IP, v Payload) []*Record {
	low := ipToNum(a)
	high := ipToNum(b)

	var rs []*Record
	ns := rangeToSubnet(low, high)
	for i, _ := range ns {
		rs = append(rs, &Record{i: ns[i], v: v})
	}
	return rs
}

func rangeToSubnet(low, high uint32) []cidr {
	if low > high {
		low, high = high, low
	}

	var ns []cidr
	lxh := low ^ high

	// find the LSB that equal
	i := lxh
	j := 32
	for (i & 1) != 0 {
		i >>= 1
		j--
	}

	// already in a subnet
	if i == 0 && (low|lxh) == high {
		ns = append(ns,
			cidr{
				prefix: low,
				size:   j,
			})
	} else {
		// find the MSB that differ
		i = lxh
		j = 0
		for i>>1 != 0 {
			i >>= 1
			j++
		}
		i <<= uint(j)
		i = ^(i - 1) & high
		ns = append(ns, rangeToSubnet(low, i-1)...)
		ns = append(ns, rangeToSubnet(i, high)...)
	}

	return ns
}

func ipToNum(ip net.IP) uint32 {
	if ip != nil {
		ip = ip.To4()
	}

	if ip == nil {
		return 0
	}

	arr := []byte(ip)
	return uint32(arr[0])<<24 | uint32(arr[1])<<16 | uint32(arr[2])<<8 | uint32(arr[3])
}

func sizeToMask(n int) uint32 {
	return ^uint32(0) << uint(32-n)
}

func vequal(a, b Payload) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return a.Equal(b)
}
