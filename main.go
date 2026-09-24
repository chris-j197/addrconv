package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintln(os.Stderr, `addrconv - convert between structured and USPS address formats

usage:
  addrconv encode [-json]   read a JSON Address object from stdin, write a
                            USPS delivery address block (2-3 lines) to stdout
  addrconv decode [-json]   read a USPS delivery address block (2-3 lines)
                            from stdin, write a JSON Address object to stdout

  -json switches either command to batch mode: encode reads a JSON array
  of Address objects and writes a JSON array of line blocks; decode reads
  a JSON array of line blocks and writes a JSON array of Address objects.

JSON Address shape:
  {"recipient": "...", "street": "...", "unit": "...",
   "city": "...", "state": "...", "zip5": "...", "zip4": "..."}`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	batch := fs.Bool("json", false, "batch mode: convert a JSON array instead of a single address")
	fs.Parse(os.Args[2:])

	var err error
	switch cmd {
	case "encode":
		if *batch {
			err = runEncodeBatch(os.Stdin, os.Stdout)
		} else {
			err = runEncode(os.Stdin, os.Stdout)
		}
	case "decode":
		if *batch {
			err = runDecodeBatch(os.Stdin, os.Stdout)
		} else {
			err = runDecode(os.Stdin, os.Stdout)
		}
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "addrconv:", err)
		os.Exit(1)
	}
}

func runEncode(in *os.File, out *os.File) error {
	var a Address
	if err := json.NewDecoder(in).Decode(&a); err != nil {
		return fmt.Errorf("reading JSON address: %w", err)
	}
	if err := a.Validate(); err != nil {
		return err
	}
	for _, line := range a.Lines() {
		fmt.Fprintln(out, line)
	}
	return nil
}

func runEncodeBatch(in *os.File, out *os.File) error {
	var addrs []Address
	if err := json.NewDecoder(in).Decode(&addrs); err != nil {
		return fmt.Errorf("reading JSON address array: %w", err)
	}
	blocks := make([][]string, len(addrs))
	for i, a := range addrs {
		if err := a.Validate(); err != nil {
			return fmt.Errorf("address %d: %w", i, err)
		}
		blocks[i] = a.Lines()
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(blocks)
}

func runDecode(in *os.File, out *os.File) error {
	var lines []string
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading address lines: %w", err)
	}

	a, err := ParseLines(lines)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(a)
}

func runDecodeBatch(in *os.File, out *os.File) error {
	var blocks [][]string
	if err := json.NewDecoder(in).Decode(&blocks); err != nil {
		return fmt.Errorf("reading JSON line-block array: %w", err)
	}
	addrs := make([]Address, len(blocks))
	for i, lines := range blocks {
		a, err := ParseLines(lines)
		if err != nil {
			return fmt.Errorf("address %d: %w", i, err)
		}
		addrs[i] = a
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(addrs)
}
