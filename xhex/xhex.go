package xhex

import (
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
	Seek         int
	Length       int
	Columns      uint
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
	buf.Grow((1 + ((len(data) - 1) / 16)) * 78)

	dumper, err := Dumper(&buf, config)
	if err != nil {
		return "", err
	}
	_, _ = dumper.Write(data)
	_ = dumper.Close()
	return buf.String(), nil
}

// Dumper returns a WriteCloser that writes a hex dump of all written data to w
func Dumper(w io.Writer, config *DumperConfig) (io.WriteCloser, error) {
	if config != nil {
		// TODO validate config:
		// - groupSize must be a power of 2

		if config.Columns > 16 || config.Columns < 1 {
			return nil, errors.New("columns must be between 1 and 16")
		}

		if config.GroupSize > config.Columns {
			return nil, errors.New("columns must be larger than or equal to groupSize")
		}

		return &dumper{
			w:            w,
			groupSize:    config.GroupSize,
			littleEndian: config.LittleEndian,
			seek:         config.Seek,
			length:       config.Length,
			columns:      config.Columns,
		}, nil
	} else {
		return &dumper{
			w:            w,
			groupSize:    1,
			littleEndian: false,
			seek:         -1,
			length:       -1,
			columns:      16,
		}, nil
	}
}

type dumper struct {
	w          io.Writer
	rightChars [19]byte
	buf        [14]byte
	groupBuf   [33]byte

	used   uint // number of bytes in the current line
	n      uint // number of bytes, total
	closed bool

	groupSize    uint
	littleEndian bool // true: littleEndian, false: bigEndian
	seek         int  // -> should handle at bufio?
	length       int  // -> should handle at bufio?
	columns      uint
}

func toChar(b byte) byte {
	if b < 32 || b > 126 {
		return '.'
	}
	return b
}

func (h *dumper) Write(data []byte) (n int, err error) {
	if h.closed {
		return 0, errors.New("encoding/hex: dumper closed")
	}

	h.rightChars[0] = ' '

	l := 2 * h.groupSize
	// groupSize=1 -> add space after 0,1,2,3,4
	// groupSize=2 -> add space after 1,3,5,...15
	// groupSize=4 -> add space after 3,7,10,...15
	// groupSize=8 -> add space after 7,15
	h.groupBuf[l] = ' '

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
			Encode(h.buf[4:], h.buf[:4])
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
		Encode(h.groupBuf[j:], data[i:i+1])

		n++
		h.rightChars[h.used+1] = toChar(data[i])
		h.used++
		h.n++

		if h.used%h.groupSize == 0 {
			_, err = h.w.Write(h.groupBuf[:l+1])
			if err != nil {
				return
			}
		}

		if h.used == h.columns {
			h.rightChars[h.columns+1] = ' '
			h.rightChars[h.columns+2] = '\n'
			_, err = h.w.Write(h.rightChars[:])
			if err != nil {
				return
			}
			h.used = 0
		}

	}
	return
}

func (h *dumper) Close() (err error) {
	//// See the comments in Write() for the details of this format.
	//if h.closed {
	//	return
	//}
	//h.closed = true
	//if h.used == 0 {
	//	return
	//}
	//h.buf[0] = ' '
	//h.buf[1] = ' '
	//h.buf[2] = ' '
	//h.buf[3] = ' '
	//h.buf[4] = '|'
	//nBytes := h.used
	//for h.used < 16 {
	//	l := 3
	//	if h.used == 15 {
	//		l = 4
	//	}
	//	_, err = h.w.Write(h.buf[:l])
	//	if err != nil {
	//		return
	//	}
	//	h.used++
	//}
	//h.rightChars[nBytes] = '|'
	//h.rightChars[nBytes+1] = '\n'
	//_, err = h.w.Write(h.rightChars[:nBytes+2])
	return
}

// Encode encodes src into EncodedLen(len(src))
// bytes of dst. As a convenience, it returns the number
// of bytes written to dst, but this value is always EncodedLen(len(src)).
// Encode implements hexadecimal encoding.
func Encode(dst, src []byte) int {
	j := 0
	for _, v := range src {
		dst[j] = hextable[v>>4]
		dst[j+1] = hextable[v&0x0f]
		j += 2
	}
	return len(src) * 2
}
