// Package dump writes binary data as a classic hex dump.
package dump

import "fmt"

// NiceFloor rounds n down to a round number of bytes per line: a multiple
// of 16 from 32 up, a multiple of 8 from 8 up, and a power of two below
// that. It returns at least 1.
func NiceFloor(n int) int {
	switch {
	case n >= 32:
		return n / 16 * 16
	case n >= 8:
		return n / 8 * 8
	case n < 1:
		return 1
	}
	p := 1
	for p*2 <= n {
		p *= 2
	}
	return p
}

// OffsetWidth is the number of hex digits used for offsets: 8, or more if
// the last offset needs them. last < 0 means unknown.
func OffsetWidth(last int64) int {
	w := 8
	if last > 0 {
		if n := len(fmt.Sprintf("%x", last)); n > w {
			w = n
		}
	}
	return w
}
