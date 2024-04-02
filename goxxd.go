package goxxd

import (
	"bufio"
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

func Dump(textReader io.Reader, columns, group, seek, length int, littleEndian bool) (string, error) {
	config := &DumperConfig{
		Columns:      uint(columns),
		GroupSize:    uint(group),
		LittleEndian: littleEndian,
	}

	reader := bufio.NewReader(textReader)
	_, err := reader.Discard(seek)
	if err != nil {
		return "", err
	}

	buf := make([]byte, 256)

	var out strings.Builder
	var totalBytes int

	for {
		numBytes, err := reader.Read(buf)

		if err != nil {
			if !errors.Is(err, io.EOF) {
				return "", err
			}
			break
		}

		l := numBytes

		if length != -1 && totalBytes+numBytes > length {
			l = length - totalBytes
		}

		config.Offset = totalBytes + seek

		dump, err := dump(buf[:l], config)
		if err != nil {
			return "", err
		}

		_, err = out.WriteString(dump)
		if err != nil {
			return "", err
		}

		totalBytes += l
		if length != -1 && totalBytes >= length {
			break
		}
	}

	return out.String(), nil
}

func RevertDump(dumpReader io.Reader) (string, error) {
	reader := bufio.NewReader(dumpReader)

	var out strings.Builder
	var buf []byte
	var src []byte
	var dst []byte
	for {
		if buf == nil {
			var err error
			buf, err = reader.ReadSlice('\n')
			if err != nil {
				return "", err
			}
			src = make([]byte, len(buf))
		} else {
			// every line has the same length except the last one,
			// so we can reuse `buf`
			_, err := reader.Read(buf)
			if err != nil {
				if err != io.EOF {
					return "", err
				}
				break
			}
		}

		// buf="00000000  46696c65 20322063 6f6e7465 6e74730a  File 2 contents."
		// -> src="46696c65203220636f6e74656e74730a"
		var started bool
		var i, j int
		for {
			if i == len(buf)-1 {
				break
			}

			if buf[i] == ' ' && buf[i+1] == ' ' {
				if started {
					break
				}

				i += 2
				started = true
				continue
			}

			if started && buf[i] != ' ' {
				src[j] = buf[i]
				j++
			}

			i++
		}

		// src="46696c65203220636f6e74656e74730a"
		// -> dst="File 2 contents."
		if dst == nil {
			dst = make([]byte, hex.DecodedLen(len(src[:j])))
		}
		writtenBytes, err := hex.Decode(dst, src[:j])
		if err != nil {
			return "", err
		}

		out.Write(dst[:writtenBytes])
		if err != nil {
			return "", err
		}
	}

	return out.String(), nil
}
