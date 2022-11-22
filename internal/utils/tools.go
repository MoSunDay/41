package utils

import (
	"encoding/binary"
	"net"
	"unsafe"
)

func IP2int(ip net.IP) uint32 {
	if len(ip) == 0 {
		return 0
	}

	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func SliceToString(buf []byte) string {
	return *(*string)(unsafe.Pointer(&buf))
}

func AtoI(s []byte, base int) (num int, ok bool) {
	var v int
	ok = true
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			ok = false
			break
		}
		v = int(HexTable[s[i]])
		if v >= base || (v == 0 && s[i] != '0') {
			ok = false
			break
		}
		num = (num * base) + v
	}
	return
}

var HexTable = [128]byte{
	'0': 0,
	'1': 1,
	'2': 2,
	'3': 3,
	'4': 4,
	'5': 5,
	'6': 6,
	'7': 7,
	'8': 8,
	'9': 9,
	'A': 10,
	'a': 10,
	'B': 11,
	'b': 11,
	'C': 12,
	'c': 12,
	'D': 13,
	'd': 13,
	'E': 14,
	'e': 14,
	'F': 15,
	'f': 15,
}

// Cut elements from slice for a given range
func Cut(a []byte, from, to int) []byte {
	copy(a[from:], a[to:])
	a = a[:len(a)-to+from]

	return a
}

// Insert new slice at specified position
func Insert(a []byte, i int, b []byte) []byte {
	a = append(a, make([]byte, len(b))...)
	copy(a[i+len(b):], a[i:])
	copy(a[i:i+len(b)], b)

	return a
}

// Replace function unlike bytes.Replace allows you to specify range
func Replace(a []byte, from, to int, new []byte) []byte {
	lenDiff := len(new) - (to - from)

	if lenDiff > 0 {
		// Extend if new segment bigger
		a = append(a, make([]byte, lenDiff)...)
		copy(a[to+lenDiff:], a[to:])
		copy(a[from:from+len(new)], new)

		return a
	}

	if lenDiff < 0 {
		copy(a[from:], new)
		copy(a[from+len(new):], a[to:])
		return a[:len(a)+lenDiff]
	}

	// same size
	copy(a[from:], new)
	return a
}
