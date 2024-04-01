
# Project Title

## Overview

**Disclaimer: This repository is for educational purpose only.**

The idea is based on the standard library `encoding/hex`, and I've modified it to support more features as the [John Crickett's Coding Challenge](https://codingchallenges.fyi/challenges/challenge-xxd/) suggested.

## Features

- Print the hexdump of a binary file. Some support options are:
  - maximum length
  - little-endian byte order
  - number of hex values per line
  - ... and more. Run `goxxd -h` to see the full list.
- Revert the hexdump back to the binary file.

## Example

Build the tool:
```bash
go build goxxd.go
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
# 00000000  46 69 6c 65 20 31 20 63 6f 6e 74 65 6e 74 73 3a  File 1 contents:
# 00000010  20 6c 6f 72 65 6d 20 69 70 73 75 6d 20 64 6f 6c   lorem ipsum dol
# 00000020  6f 72 20 73 69 74 20 61 6d 65 74 0a              or sit amet.
```

## How does it work?

## Acknowledgements
