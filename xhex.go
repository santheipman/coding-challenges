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
	w          io.Writer
	rightChars [19]byte
	buf        [14]byte
	groupBuf   [33]byte

	used   uint // number of bytes in the current line
	n      uint // number of bytes, total
	closed bool

	groupSize    uint // number of bytes in a group
	littleEndian bool // true: littleEndian, false: bigEndian
	columns      uint // number of bytes per line
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
	// groupSize=1 -> add space after 0,1,2,3,4
	// groupSize=2 -> add space after 1,3,5,...15
	// groupSize=4 -> add space after 3,7,10,...15
	// groupSize=8 -> add space after 7,15
	h.groupBuf[l-1] = ' '

	// Output lines look like:
	// 00000010  2e 2f 30 31 32 33 34 35 36 37 38 39 3a 3b 3c 3d  ./0123456789:;<=
	// ^ offset                                                   ^ ASCII of line.
	for i := range data {
		if h.used == 0 {
			// At the beginning of a line we print the current
			// offset in hex.

			// San: this part the author convert `h.n` to binary digits and store in the first four elements of `buf`,
			// then convert each of them to hexcimal and store in the rest elements of `buf`
			h.buf[0] = byte(h.n >> 24)
			h.buf[1] = byte(h.n >> 16)
			h.buf[2] = byte(h.n >> 8)
			h.buf[3] = byte(h.n)
			hex.Encode(h.buf[4:], h.buf[:4])
			h.buf[12] = ' '
			h.buf[13] = ' '
			_, err = h.w.Write(h.buf[4:])
			if err != nil {
				return
			}
		}
		var j uint
		if h.littleEndian {
			j = 2 * ((h.groupSize - 1) - (h.used % h.groupSize))
		} else {
			j = 2 * (h.used % h.groupSize)
		}
		hex.Encode(h.groupBuf[j:], data[i:i+1])

		n++
		h.used++
		h.rightChars[h.used] = toChar(data[i])
		h.n++

		if i == len(data)-1 { // fill empty space (if needed) for the last group
			if h.littleEndian {
				for k := uint(0); k < j; k++ {
					h.groupBuf[k] = ' '
				}
			} else {
				for k := uint(len(h.groupBuf)) - 1; k > j+1; k-- {
					h.groupBuf[k] = ' '
				}
			}
		}

		if h.used%h.groupSize == 0 || i == len(data)-1 {
			_, err = h.w.Write(h.groupBuf[:l])
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
// `columns` bytes. See the comments in write() for the details of this format.
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
		h.groupBuf[i] = ' '
	}

	for i := uint(0); i < groups; i++ {
		_, err = h.w.Write(h.groupBuf[:l])
		if err != nil {
			return
		}
	}

	h.rightChars[h.used+1] = '\n'
	_, err = h.w.Write(h.rightChars[:h.used+2])
	return
}
