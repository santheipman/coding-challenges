package xhex

import (
	"fmt"
	"strconv"
	"testing"
)

func TestConvertBinToHex(t *testing.T) {
	n := 230
	buf := [4]byte{}

	var b uint8
	b = 255
	fmt.Println(strconv.FormatInt(int64(b), 2))
	b = b * 2
	fmt.Println(strconv.FormatInt(int64(b), 2))

	//fmt.Println(strconv.FormatInt(int64(n), 2))
	//fmt.Println(strconv.FormatInt(int64(n>>24), 2))
	//fmt.Println(strconv.FormatInt(int64(n>>6), 2))
	buf[0] = byte(n >> 24)
	buf[1] = byte(n >> 16)
	buf[2] = byte(n >> 8)
	buf[3] = byte(n)

	fmt.Println(buf, n)
}

func TestDump(t *testing.T) {

}
