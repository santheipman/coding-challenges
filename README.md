
# Project Title

## Overview

**Disclaimer: This repository is for educational purpose.**

The idea is based on the standard library `encoding/hex`, and I've modified it to support more features as the challenge suggested. See `dumper.Write` function in `xhex.go`.

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
echo ./samples/file1.txt
echo "hexdump:"
./goxxd ./samples/file1.txt
```

## How does it work?

## Acknowledgements
