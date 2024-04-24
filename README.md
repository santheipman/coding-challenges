
# goxxd

## Overview

**Disclaimer: This repository is for educational purpose only.**

This repository aims to solve [Build Your Own Xxd Challenge](https://codingchallenges.fyi/challenges/challenge-xxd/)  by John Crickett.

The idea is based on the standard library `encoding/hex`, and I've modified it to support more features as required by the challenge.

## Features

- Print the hexdump of a binary file. Some supported options are:
  - maximum length
  - little-endian byte order
  - number of hex values per line
  - ... and more. Run `goxxd -h` to see the full list.
- Revert the hexdump back to the binary file.

## Example

Build the tool:
```bash
go build -o goxxd cmd/goxxd/main.go
```

List the run options:
```bash
./goxxd -h
```

Hexdump a file:
```bash
./goxxd ./samples/file1.txt

# input:
# File 1 contents: lorem ipsum dolor sit amet

# output:
# 00000000: 46 69 6c 65 20 31 20 63 6f 6e 74 65 6e 74 73 3a  File 1 contents:
# 00000010: 20 6c 6f 72 65 6d 20 69 70 73 75 6d 20 64 6f 6c   lorem ipsum dol
# 00000020: 6f 72 20 73 69 74 20 61 6d 65 74 0a              or sit amet.
```

## Take-aways

### #1 Reusing Array

When converting the file's content to a hexdump, we convert byte by byte to the hex format. In doing so, we repeatedly call `Write(p []byte) (n int, err error)` to write to the output stream, such as stdout or a file. The `p []byte` can be reused between calls to reduce the workload for memory allocation. Furthermore, because the maximum size needed for `p` is 33 ([see explanation](https://github.com/santheipman/coding-challenges/blob/main/xhex.go#L92-L98)), we use an array instead of a slice and slice it to obtain a slice as needed.

However, reusing arrays/slices makes the code harder to understand and more prone to bugs. Therefore, we should only use it when performance is critical.

### #2 Concatenating strings effectively using `strings.Builder`

String in Go is a value type. When we concatenate two strings together (for example, `a := b + c`), their elements are copied to a new string, resulting in a time complexity of `O(m+n)`, where `m` and `n` represent the lengths of `b` and `c`, respectively.

For large strings or when you need to repeatedly concatenate strings, it is more efficient to use `strings.Builder`. It holds a byte slice, so each concatenation operation is performed by appending to the slice, which has a time complexity of `O(n)`, where `n` is the number of bytes to be appended. Memory allocation only occurs when the capacity of the slice is reached (see more explanation about append [here](https://stackoverflow.com/a/15703422)).

### #3 Using `io.Reader` instead of file names as function parameters for improved testability

At first, I wrote the `Dump` function with the input file name as a parameter. It looks like this:

```go
func Dump(filename string, columns, group, seek, length int, littleEndian bool) {}
```

To test this function, I had to create a separate input file for each test, which bloats up the files, is not self-contained, and is not idiomatic. Therefore, I rewrote it to accept an `io.Reader` instead.

```go
func Dump(textReader io.Reader, columns, group, seek, length int, littleEndian bool)
```
```go
// Standard package:
package io

type Reader interface {
  Read(p []byte) (n int, err error)
}
```

In the "main" code executed by running the CLI, the underlying type of `io.Reader` is `os.File`, which is used to read the content of the file. In tests, I use `strings.Reader`, which is simply created by calling `strings.NewReader(s string)`, to fulfill the `io.Reader` interface. The `strings.Reader` holds a string and a pointer to keep track of the current byte to read from. This behavior is sufficient for my needs.
