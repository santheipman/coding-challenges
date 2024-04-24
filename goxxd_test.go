package goxxd

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
)

func TestDump(t *testing.T) {
	reader := strings.NewReader("File 1 contents")
	columns := 16
	group := 1
	seek := 0
	length := -1
	littleEndian := false

	expected := "00000000: 46 69 6c 65 20 31 20 63 6f 6e 74 65 6e 74 73     File 1 contents\n"

	result, err := Dump(reader, columns, group, seek, length, littleEndian)
	if err != nil {
		t.Errorf("Dump error: %v", err)
		return
	}

	if result != expected {
		t.Errorf("expected dump [%s], got [%s]", expected, result)
	}
}

func TestDumpColumn(t *testing.T) {
	reader := strings.NewReader(`File 1 contents: lorem ipsum dolor sit amet
`)
	columns := 6
	group := 1
	seek := 0
	length := -1
	littleEndian := false

	expected := `00000000: 46 69 6c 65 20 31  File 1
00000006: 20 63 6f 6e 74 65   conte
0000000c: 6e 74 73 3a 20 6c  nts: l
00000012: 6f 72 65 6d 20 69  orem i
00000018: 70 73 75 6d 20 64  psum d
0000001e: 6f 6c 6f 72 20 73  olor s
00000024: 69 74 20 61 6d 65  it ame
0000002a: 74 0a              t.
`

	result, err := Dump(reader, columns, group, seek, length, littleEndian)
	if err != nil {
		t.Errorf("Dump error: %v", err)
		return
	}

	if result != expected {
		t.Errorf("expected dump [%s], got [%s]", expected, result)
	}
}

func TestDumpLength(t *testing.T) {
	reader := strings.NewReader(`File 1 contents: lorem ipsum dolor sit amet
`)
	columns := 16
	group := 1
	seek := 0
	length := 18
	littleEndian := false

	expected := `00000000: 46 69 6c 65 20 31 20 63 6f 6e 74 65 6e 74 73 3a  File 1 contents:
00000010: 20 6c                                             l
`

	result, err := Dump(reader, columns, group, seek, length, littleEndian)
	if err != nil {
		t.Errorf("Dump error: %v", err)
		return
	}

	if result != expected {
		t.Errorf("expected dump [%s], got [%s]", expected, result)
	}
}

func TestDumpGroup(t *testing.T) {
	reader := strings.NewReader(`File 1 contents: lorem ipsum dolor sit amet
`)
	columns := 16
	group := 4
	seek := 0
	length := -1
	littleEndian := false

	expected := `00000000: 46696c65 20312063 6f6e7465 6e74733a  File 1 contents:
00000010: 206c6f72 656d2069 7073756d 20646f6c   lorem ipsum dol
00000020: 6f722073 69742061 6d65740a           or sit amet.
`

	result, err := Dump(reader, columns, group, seek, length, littleEndian)
	if err != nil {
		t.Errorf("Dump error: %v", err)
		return
	}

	if result != expected {
		t.Errorf("expected dump [%s], got [%s]", expected, result)
	}
}

func TestDumpLittleEndian(t *testing.T) {
	reader := strings.NewReader(`File 1 contents: lorem ipsum dolor sit amet
`)
	columns := 16
	group := 2
	seek := 0
	length := -1
	littleEndian := true

	expected := `00000000: 6946 656c 3120 6320 6e6f 6574 746e 3a73  File 1 contents:
00000010: 6c20 726f 6d65 6920 7370 6d75 6420 6c6f   lorem ipsum dol
00000020: 726f 7320 7469 6120 656d 0a74            or sit amet.
`

	result, err := Dump(reader, columns, group, seek, length, littleEndian)
	if err != nil {
		t.Errorf("Dump error: %v", err)
		return
	}

	if result != expected {
		t.Errorf("expected dump [%s], got [%s]", expected, result)
	}
}

func TestDumpSeek(t *testing.T) {
	reader := strings.NewReader(`File 1 contents: lorem ipsum dolor sit amet
`)
	columns := 16
	group := 1
	seek := 12
	length := -1
	littleEndian := true

	expected := `0000000c: 6e 74 73 3a 20 6c 6f 72 65 6d 20 69 70 73 75 6d  nts: lorem ipsum
0000001c: 20 64 6f 6c 6f 72 20 73 69 74 20 61 6d 65 74 0a   dolor sit amet.
`

	result, err := Dump(reader, columns, group, seek, length, littleEndian)
	if err != nil {
		t.Errorf("Dump error: %v", err)
		return
	}

	if result != expected {
		t.Errorf("expected dump [%s], got [%s]", expected, result)
	}
}

func TestRevertDump(t *testing.T) {
	reader := strings.NewReader(`00000000: 46 69 6c 65 20 31 20 63 6f 6e 74 65 6e 74 73 3a  File 1 contents:
00000010: 20 6c 6f 72 65 6d 20 69 70 73 75 6d 20 64 6f 6c   lorem ipsum dol
00000020: 6f 72 20 73 69 74 20 61 6d 65 74 0a              or sit amet.
`)
	expected := `File 1 contents: lorem ipsum dolor sit amet
`

	result, err := RevertDump(reader)
	if err != nil {
		t.Errorf("Dump error: %v", err)
		return
	}

	if result != expected {
		t.Errorf("expected [%s], got [%s]", expected, result)
	}
}

func BenchmarkDumpStandard(b *testing.B) {
	for _, size := range []int{256, 1024, 4096, 16384} {
		src := bytes.Repeat([]byte{2, 3, 5, 7, 9, 11, 13, 17}, size/8)

		b.Run(fmt.Sprintf("%v", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			for i := 0; i < b.N; i++ {
				hex.Dump(src)
			}
		})
	}
}

func BenchmarkDump(b *testing.B) {
	config := &DumperConfig{
		Columns:      16,
		GroupSize:    1,
		LittleEndian: false,
	}

	for _, size := range []int{256, 1024, 4096, 16384} {
		bts := bytes.Repeat([]byte{2, 3, 5, 7, 9, 11, 13, 17}, size/8)

		b.Run(fmt.Sprintf("%v", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			for i := 0; i < b.N; i++ {
				_, err := dump(bts, config)
				if err != nil {
					b.Error(err)
					return
				}
			}
		})
	}
}
