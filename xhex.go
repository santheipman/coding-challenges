package goxxd

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

type DumperConfig struct {
	GroupSize    uint
	LittleEndian bool
	Columns      uint
	Offset       int
}

// dump returns a string that contains a hexdump of the given data.
func dump(data []byte, config *DumperConfig) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	var buf strings.Builder
	// Dumper will write 76 bytes at most per complete 16 byte chunk
	buf.Grow((1 + ((len(data) - 1) / 16)) * 76)

	dumper, err := NewDumper(&buf, config)
	if err != nil {
		return "", err
	}
	_, err = dumper.Write(data)
	if err != nil {
		return "", err
	}
	err = dumper.Close()
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

const (
	minColumns = 1
	maxColumns = 16
)

// NewDumper returns a Dumper that writes a hexdump of all written data to w
func NewDumper(w io.Writer, config *DumperConfig) (*Dumper, error) {
	if config == nil {
		return &Dumper{
			w:            w,
			groupSize:    1,
			littleEndian: false,
			columns:      maxColumns,
		}, nil
	}

	tmp := config.GroupSize
	for {
		if tmp < 0 || tmp%2 != 0 {
			break
		}
		tmp = tmp / 2
	}
	if config.GroupSize != 1 && tmp != 1 {
		return nil, errors.New("groupSize must be a power of 2")
	}

	if config.Columns > maxColumns || config.Columns < minColumns {
		return nil, fmt.Errorf("columns must be between %d and %d", minColumns, maxColumns)
	}

	if config.GroupSize > config.Columns {
		return nil, errors.New("columns must be larger than or equal to groupSize")
	}

	return &Dumper{
		w:            w,
		n:            uint(config.Offset),
		groupSize:    config.GroupSize,
		littleEndian: config.LittleEndian,
		columns:      config.Columns,
	}, nil
}

type Dumper struct {
	w io.Writer
	// rightChars has this format <space><16_characters><new_line>
	rightChars [18]byte
	// buf is used for write offset and group data.
	// For offset, we utilize first 4 bytes in buf for hex conversion and store the result in the buf from 4th->14th byte.
	// Therefore, the maximum bytes needed is 14.
	// For group data, we need at most 32 bytes to store the largest group size allowed (16) plus one space at the end. So the
	// maximum bytes in this case is 33.
	// Overall, we choose 33 is the size of the byte buffer.
	buf [33]byte

	// used is number of bytes in the current line
	used uint
	// n is number of bytes, total
	n      uint
	closed bool

	// groupSize is number of bytes in a group
	groupSize uint
	// littleEndian true: littleEndian, false: bigEndian
	littleEndian bool
	// columns is number of bytes per line
	columns uint
}

func toChar(b byte) byte {
	if b < 32 || b > 126 {
		return '.'
	}
	return b
}

func (h *Dumper) Write(data []byte) (n int, err error) {
	if h.closed {
		return 0, errors.New("encoding/hex: Dumper closed")
	}

	h.rightChars[0] = ' '

	l := 2*h.groupSize + 1

	// Output lines look like:
	// 00000010: 2e 2f 30 31 32 33 34 35 36 37 38 39 3a 3b 3c 3d  ./0123456789:;<=
	// ^ offset  ^ a group with size=1                            ^ ASCII of line.
	var j uint
	for i := range data {
		if h.used == 0 {
			// At the beginning of a line we print the current
			// offset in hex.
			h.buf[0] = byte(h.n >> 24)
			h.buf[1] = byte(h.n >> 16)
			h.buf[2] = byte(h.n >> 8)
			h.buf[3] = byte(h.n)
			hex.Encode(h.buf[4:], h.buf[:4])
			h.buf[12] = ':'
			h.buf[13] = ' '
			_, err = h.w.Write(h.buf[4:14])
			if err != nil {
				return
			}
			h.buf[l-1] = ' '
		}
		if h.littleEndian {
			j = 2 * ((h.groupSize - 1) - (h.used % h.groupSize))
		} else {
			j = 2 * (h.used % h.groupSize)
		}
		hex.Encode(h.buf[j:], data[i:i+1])

		n++
		h.used++
		h.rightChars[h.used] = toChar(data[i])
		h.n++

		if i == len(data)-1 { // fill empty space (if needed) for the last group
			if h.littleEndian {
				for k := uint(0); k < j; k++ {
					h.buf[k] = ' '
				}
			} else {
				for k := uint(len(h.buf)) - 1; k > j+1; k-- {
					h.buf[k] = ' '
				}
			}
		}

		if h.used%h.groupSize == 0 || i == len(data)-1 {
			_, err = h.w.Write(h.buf[:l])
			if err != nil {
				return
			}
		}

		if h.used == h.columns {
			h.rightChars[h.columns+1] = '\n'
			_, err = h.w.Write(h.rightChars[:h.columns+2])
			if err != nil {
				return
			}
			h.used = 0
		}

	}
	return
}

// Close pads empty space for the last line in case it has less than
// `columns` bytes. See the comments in Write() for the details of this format.
func (h *Dumper) Close() (err error) {
	if h.closed {
		return
	}
	h.closed = true
	if h.used == 0 {
		return
	}

	pads := h.columns - h.used
	groups := pads / h.groupSize
	l := 2*h.groupSize + 1

	for i := uint(0); i < l; i++ {
		h.buf[i] = ' '
	}

	for i := uint(0); i < groups; i++ {
		_, err = h.w.Write(h.buf[:l])
		if err != nil {
			return
		}
	}

	h.rightChars[h.used+1] = '\n'
	_, err = h.w.Write(h.rightChars[:h.used+2])
	return
}
