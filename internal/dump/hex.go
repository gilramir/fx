package dump

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

// HexStandardCols is the bytes per line when not fitting a terminal.
const HexStandardCols = 16

// HexOptions controls a classic hex dump, laid out like xxd:
//
//	00000000: 4865 6c6c 6f2c 2077 6f72 6c64 210a 5468  Hello, world!.Th
type HexOptions struct {
	Cols        int   // bytes per line
	Group       int   // bytes per space-separated group
	Upper       bool  // upper-case hex digits
	Start       int64 // offset of the first byte, for the offset column
	OffsetWidth int   // hex digits in the offset column
}

// HexLineWidth is the width in columns of a full line.
func HexLineWidth(cols, group, offW int) int {
	groups := (cols + group - 1) / group
	return offW + 2 + 2*cols + groups - 1 + 2 + cols
}

// FitHexCols returns the largest round number of bytes per line whose
// lines fit in width columns. Lines are kept to a whole number of groups
// when there's room for at least one.
func FitHexCols(width, group, offW int) int {
	n := 1
	for HexLineWidth(n+1, group, offW) <= width {
		n++
	}
	n = NiceFloor(n)
	if n >= group {
		n = n / group * group
	}
	return n
}

func (o HexOptions) validate() error {
	if o.Cols < 1 {
		return fmt.Errorf("bytes per line must be at least 1")
	}
	if o.Group < 1 {
		return fmt.Errorf("group size must be at least 1")
	}
	return nil
}

// Hex writes a hex dump of r to w.
func Hex(w io.Writer, r io.Reader, o HexOptions) error {
	if err := o.validate(); err != nil {
		return err
	}
	digits := "0123456789abcdef"
	if o.Upper {
		digits = "0123456789ABCDEF"
	}
	bw := bufio.NewWriter(w)
	buf := make([]byte, o.Cols)
	line := make([]byte, 0, HexLineWidth(o.Cols, o.Group, o.OffsetWidth)+1)
	offset := o.Start
	for {
		n, err := io.ReadFull(r, buf)
		if n > 0 {
			line = line[:0]
			line = fmt.Appendf(line, "%0*x: ", o.OffsetWidth, offset)
			for i := 0; i < o.Cols; i++ {
				if i > 0 && i%o.Group == 0 {
					line = append(line, ' ')
				}
				if i < n {
					line = append(line, digits[buf[i]>>4], digits[buf[i]&15])
				} else {
					line = append(line, ' ', ' ')
				}
			}
			line = append(line, ' ', ' ')
			for _, b := range buf[:n] {
				if b < 0x20 || b > 0x7e {
					b = '.'
				}
				line = append(line, b)
			}
			line = append(line, '\n')
			if _, werr := bw.Write(line); werr != nil {
				return werr
			}
			offset += int64(n)
		}
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			break
		}
		if err != nil {
			return err
		}
	}
	return bw.Flush()
}
