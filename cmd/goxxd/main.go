package main

import (
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"goxxd"
	"os"
)

var (
	littleEndian = kingpin.Flag("little-endian", "Switch to little-endian hex dump.").Default("false").Short('e').Bool()
	group        = kingpin.Flag("group", "Group hex values.").Default("1").Short('g').Int()
	columns      = kingpin.Flag("columns", "Specify number of hex values per line.").Default("16").Short('c').Int()
	length       = kingpin.Flag("length", "Specify the maximum number of hex values to be printed. Print all values by default.").Default("-1").Short('l').Int()
	seek         = kingpin.Flag("seek", "Print hex values start from `seek`-th byte (0-based).").Default("0").Short('s').Int()
	revert       = kingpin.Flag("revert", "Revert a hex dump back to a binary file.").Default("false").Short('r').Bool()

	filename = kingpin.Arg("filename", "filename").Required().String()
)

func main() {
	kingpin.Parse()

	ioReader, err := os.Open(*filename)
	if err != nil {
		fmt.Print(err)
		return
	}

	if *revert {
		text, err := goxxd.RevertDump(ioReader)
		if err != nil {
			panic(err)
		}
		fmt.Print(text)
		return
	}

	hexdump, err := goxxd.Dump(ioReader, *columns, *group, *seek, *length, *littleEndian)
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Print(hexdump)
}
