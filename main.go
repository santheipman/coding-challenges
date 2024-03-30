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

	reader := bufio.NewReader(f)
	buf := make([]byte, 256)

	for {
		_, err := reader.Read(buf)

		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}

		config := &xhex.DumperConfig{
			Columns:      8,
			GroupSize:    4,
			LittleEndian: true,
			Seek:         0,
			Length:       0,
		}

		dump, err := xhex.Dump(buf, config)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", dump)
		return
	}
}
