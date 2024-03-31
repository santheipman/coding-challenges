package xhex

import (
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

const (
	hextable = "0123456789abcdef"
	//reverseHexTable = "" +
	//	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	//	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	//	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	//	"\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\xff\xff\xff\xff\xff\xff" +
	//	"\xff\x0a\x0b\x0c\x0d\x0e\x0f\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	//	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	//	"\xff\x0a\x0b\x0c\x0d\x0e\x0f\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	//	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	//	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	//	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	//	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	//	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	//	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	//	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	//	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
	//	"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff"
)

type DumperConfig struct {
	GroupSize    uint
	LittleEndian bool
	Columns      uint
	Offset       int
}

// Dump returns a string that contains a hex dump of the given data. The format
// of the hex dump matches the output of `hexdump -C` on the command line.
func Dump(data []byte, config *DumperConfig) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	var buf strings.Builder
	// Dumper will write 78 bytes per complete 16 byte chunk, and at least
	// 64 bytes for whatever remains. Round the allocation up, since only a
	// maximum of 15 bytes will be wasted.
	//buf.Grow((1 + ((len(data) - 1) / 16)) * 78)
	buf.Grow(2000) // TODO @San

	dumper, err := NewDumper(&buf, config)
	if err != nil {
		return "", err
	}
	_, _ = dumper.Write(data)
	_ = dumper.Close()
	return buf.String(), nil
}

// NewDumper returns a Dumper that writes a hex dump of all written data to w
func NewDumper(w io.Writer, config *DumperConfig) (*Dumper, error) {
	if config != nil {
		// TODO validate config:
		// - groupSize must be a power of 2

		if config.Columns > 16 || config.Columns < 1 {
			return nil, errors.New("columns must be between 1 and 16")
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
	} else {
		return &Dumper{
			w:            w,
			groupSize:    1,
			littleEndian: false,
			columns:      16,
		}, nil
	}
}

type Dumper struct {
	w          io.Writer
	rightChars [19]byte
	buf        [14]byte
	groupBuf   [33]byte

	used   uint // number of bytes in the current line
	n      uint // number of bytes, total
	closed bool

	groupSize    uint
	littleEndian bool // true: littleEndian, false: bigEndian
	columns      uint
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
			_, err = h.w.Write(h.rightChars[:])
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
