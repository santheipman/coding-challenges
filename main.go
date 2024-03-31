package main

import (
	"bufio"
	"fmt"
	"goxxd/xhex"
	"io"
	"log"
	"os"
)

func main() {
	read("./samples/file1.txt")
}

func read(filename string) {
	f, err := os.Open(filename)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	config := &xhex.DumperConfig{
		Columns:      16,
		GroupSize:    4,
		LittleEndian: false,
	}
	seek := 1
	length := 19
	length = -1

	reader := bufio.NewReader(f)
	_, err = reader.Discard(seek - 1)
	if err != nil {
		panic(err)
	}
	buf := make([]byte, 256)

	var totalBytes int

	for {
		numBytes, err := reader.Read(buf)

		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
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
			panic(err)
		}
		fmt.Printf("%s", dump)

		totalBytes += l
		if length != -1 && totalBytes >= length {
			break
		}
	}
}
