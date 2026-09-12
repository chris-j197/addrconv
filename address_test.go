package main

import (
	"reflect"
	"testing"
)

func TestLines(t *testing.T) {
	cases := []struct {
		name string
		addr Address
		want []string
	}{
		{
			name: "recipient, unit, and zip4",
			addr: Address{
				Recipient: "Jane Doe",
				Street:    "123 North Main Street",
				Unit:      "Apartment 4B",
				City:      "Springfield",
				State:     "IL",
				Zip5:      "62704",
				Zip4:      "1234",
			},
			want: []string{
				"JANE DOE",
				"123 N MAIN ST APT 4B",
				"SPRINGFIELD IL 62704-1234",
			},
		},
		{
			name: "no recipient, no unit, no zip4",
			addr: Address{
				Street: "44 East Oak Avenue",
				City:   "Reno",
				State:  "NV",
				Zip5:   "89501",
			},
			want: []string{
				"44 E OAK AVE",
				"RENO NV 89501",
			},
		},
		{
			name: "punctuation and extra whitespace in street",
			addr: Address{
				Street: "  99   S.  Elm  Rd.  ",
				City:   "Troy",
				State:  "OH",
				Zip5:   "45373",
			},
			want: []string{
				"99 S ELM RD",
				"TROY OH 45373",
			},
		},
		{
			name: "suffix that abbreviates to itself",
			addr: Address{
				Street: "1 Loop",
				City:   "Nowhere",
				State:  "TX",
				Zip5:   "75001",
			},
			want: []string{
				"1 LOOP",
				"NOWHERE TX 75001",
			},
		},
		{
			name: "less common suffix and alternate spelling",
			addr: Address{
				Street: "8 Fifth Crossing",
				Unit:   "Basement 2",
				City:   "Duluth",
				State:  "MN",
				Zip5:   "55801",
			},
			want: []string{
				"8 FIFTH XING BSMT 2",
				"DULUTH MN 55801",
			},
		},
		{
			name: "trailer unit and turnpike suffix",
			addr: Address{
				Street: "17 Old Turnpike",
				Unit:   "Trailer 9",
				City:   "Concord",
				State:  "NH",
				Zip5:   "03301",
			},
			want: []string{
				"17 OLD TPKE TRLR 9",
				"CONCORD NH 03301",
			},
		},
		{
			name: "two-word directional",
			addr: Address{
				Street: "200 North East Main Street",
				City:   "Grand Rapids",
				State:  "MI",
				Zip5:   "49503",
			},
			want: []string{
				"200 NE MAIN ST",
				"GRAND RAPIDS MI 49503",
			},
		},
		{
			name: "post office box spelled out",
			addr: Address{
				Street: "Post Office Box 55",
				City:   "Boise",
				State:  "ID",
				Zip5:   "83702",
			},
			want: []string{
				"PO BOX 55",
				"BOISE ID 83702",
			},
		},
		{
			name: "po box as separated letters",
			addr: Address{
				Street: "P O Box 900",
				City:   "Helena",
				State:  "MT",
				Zip5:   "59601",
			},
			want: []string{
				"PO BOX 900",
				"HELENA MT 59601",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.addr.Lines()
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Lines() = %#v, want %#v", got, c.want)
			}
		})
	}
}

func TestParseLines(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
		want  Address
	}{
		{
			name: "three lines with recipient and unit",
			lines: []string{
				"JANE DOE",
				"123 N MAIN ST APT 4B",
				"SPRINGFIELD IL 62704-1234",
			},
			want: Address{
				Recipient: "JANE DOE",
				Street:    "123 N MAIN ST",
				Unit:      "APT 4B",
				City:      "SPRINGFIELD",
				State:     "IL",
				Zip5:      "62704",
				Zip4:      "1234",
			},
		},
		{
			name: "two lines, no recipient, no unit, no zip4",
			lines: []string{
				"44 E OAK AVE",
				"RENO NV 89501",
			},
			want: Address{
				Street: "44 E OAK AVE",
				City:   "RENO",
				State:  "NV",
				Zip5:   "89501",
			},
		},
		{
			name: "city with multiple words",
			lines: []string{
				"1 MAIN ST",
				"WEST PALM BEACH FL 33401",
			},
			want: Address{
				Street: "1 MAIN ST",
				City:   "WEST PALM BEACH",
				State:  "FL",
				Zip5:   "33401",
			},
		},
		{
			name: "blank lines around content are ignored",
			lines: []string{
				"",
				"JANE DOE",
				"123 N MAIN ST",
				"SPRINGFIELD IL 62704",
				"",
			},
			want: Address{
				Recipient: "JANE DOE",
				Street:    "123 N MAIN ST",
				City:      "SPRINGFIELD",
				State:     "IL",
				Zip5:      "62704",
			},
		},
		{
			name: "lowercase input is upcased",
			lines: []string{
				"123 n main st",
				"springfield il 62704",
			},
			want: Address{
				Street: "123 N MAIN ST",
				City:   "SPRINGFIELD",
				State:  "IL",
				Zip5:   "62704",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseLines(c.lines)
			if err != nil {
				t.Fatalf("ParseLines() error = %v", err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("ParseLines() = %#v, want %#v", got, c.want)
			}
		})
	}
}

func TestParseLinesErrors(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
	}{
		{
			name:  "single line",
			lines: []string{"SPRINGFIELD IL 62704"},
		},
		{
			name:  "four non-blank lines",
			lines: []string{"A", "B", "C", "D"},
		},
		{
			name:  "no lines",
			lines: nil,
		},
		{
			name:  "last line missing state and zip",
			lines: []string{"123 N MAIN ST", "SPRINGFIELD"},
		},
		{
			name:  "last line has three-letter state",
			lines: []string{"123 N MAIN ST", "SPRINGFIELD ILL 62704"},
		},
		{
			name:  "last line has malformed zip",
			lines: []string{"123 N MAIN ST", "SPRINGFIELD IL 627"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := ParseLines(c.lines); err == nil {
				t.Errorf("ParseLines(%v) error = nil, want error", c.lines)
			}
		})
	}
}

func TestLinesParseLinesRoundTrip(t *testing.T) {
	a := Address{
		Recipient: "John Smith",
		Street:    "500 Southwest Parkway",
		Unit:      "Suite 200",
		City:      "Austin",
		State:     "TX",
		Zip5:      "78701",
		Zip4:      "0001",
	}

	got, err := ParseLines(a.Lines())
	if err != nil {
		t.Fatalf("ParseLines() error = %v", err)
	}

	want := Address{
		Recipient: "JOHN SMITH",
		Street:    "500 SW PKWY",
		Unit:      "STE 200",
		City:      "AUSTIN",
		State:     "TX",
		Zip5:      "78701",
		Zip4:      "0001",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("round trip = %#v, want %#v", got, want)
	}
}
