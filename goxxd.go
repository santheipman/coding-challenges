package main

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"goxxd/xhex"
	"io"
	"os"
	"strings"
)

var (
	littleEndian = kingpin.Flag("little-endian", "Switch to little-endian hex dump.").Default("false").Short('e').Bool()
	group        = kingpin.Flag("group", "Group hex values.").Default("1").Short('g').Int()
	columns      = kingpin.Flag("columns", "Specify number of hex values per line. `c` must be between 1 and 16").Default("16").Short('c').Int()
	length       = kingpin.Flag("length", "Specify the maximum number of hex values to be printed. Print all values by default.").Default("-1").Short('l').Int()
	seek         = kingpin.Flag("seek", "Print hex values start from `seek`.").Default("1").Short('s').Int()
	revert       = kingpin.Flag("revert", "Revert a hex dump back to a binary file.").Default("false").Short('r').Bool()

	filename = kingpin.Arg("filename", "filename").Required().String()
)

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()

	if *revert {
		text, err := convertBack(*filename)
		if err != nil {
			panic(err)
		}
		fmt.Print(text)
		return
	}

	hexdump, err := dump(*filename, *columns, *group, *seek, *length, *littleEndian)
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Print(hexdump)
}

func dump(filename string, columns, group, seek, length int, littleEndian bool) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	defer f.Close()

	config := &xhex.DumperConfig{
		Columns:      uint(columns),
		GroupSize:    uint(group),
		LittleEndian: littleEndian,
	}

	reader := bufio.NewReader(f)
	_, err = reader.Discard(seek - 1)
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

		config.Offset = totalBytes + (seek - 1)

		dump, err := xhex.Dump(buf[:l], config)
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

func convertBack(hexdumpFilename string) (string, error) {
	f, err := os.Open(hexdumpFilename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	var out strings.Builder
	var buf []byte
	var src []byte
	var dst []byte
	for {
		if buf == nil {
			buf, err = reader.ReadSlice('\n')
			if err != nil {
				return "", err
			}
			src = make([]byte, len(buf))
		} else {
			// every line has the same length except the last one,
			// so we can reuse `buf`
			_, err = reader.Read(buf)
			if err != nil {
				if err != io.EOF {
					return "", err
				}
				break
			}
		}

		// buf="00000000  46696c65 20312063 6f6e7465 6e747373  File 1 contentss"
		// -> src="46696c65203120636f6e74656e747373"
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

		// src="46696c65203120636f6e74656e747373"
		// -> dst="File 1 contentss"
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
