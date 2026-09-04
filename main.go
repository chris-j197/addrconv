package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintln(os.Stderr, `addrconv - convert between structured and USPS address formats

usage:
  addrconv encode   read a JSON Address object from stdin, write a USPS
                    delivery address block (2-3 lines) to stdout
  addrconv decode   read a USPS delivery address block (2-3 lines) from
                    stdin, write a JSON Address object to stdout

JSON Address shape:
  {"recipient": "...", "street": "...", "unit": "...",
   "city": "...", "state": "...", "zip5": "...", "zip4": "..."}`)
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "encode":
		err = runEncode(os.Stdin, os.Stdout)
	case "decode":
		err = runDecode(os.Stdin, os.Stdout)
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
	for _, line := range a.Lines() {
		fmt.Fprintln(out, line)
	}
	return nil
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
