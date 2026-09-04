// Package main converts between a structured address (the shape you'd get
// out of a web form or a database row) and the standardized delivery
// address block USPS expects on a mailpiece: uppercase, punctuation
// stripped, street suffixes and directionals abbreviated per USPS
// Publication 28.
package main

import (
	"fmt"
	"regexp"
	"strings"
)

// Address is the structured form: one field per piece of data. This is
// the shape most systems actually store addresses in.
type Address struct {
	Recipient string `json:"recipient,omitempty"`
	Street    string `json:"street"`
	Unit      string `json:"unit,omitempty"`
	City      string `json:"city"`
	State     string `json:"state"`
	Zip5      string `json:"zip5"`
	Zip4      string `json:"zip4,omitempty"`
}

// suffixAbbrev covers the common street suffixes from USPS Pub 28
// Appendix C1. It's not the full 200+ entry table, just the ones that
// show up in practice; extending it is a matter of adding rows.
var suffixAbbrev = map[string]string{
	"STREET":    "ST",
	"AVENUE":    "AVE",
	"BOULEVARD": "BLVD",
	"DRIVE":     "DR",
	"LANE":      "LN",
	"ROAD":      "RD",
	"COURT":     "CT",
	"CIRCLE":    "CIR",
	"PLACE":     "PL",
	"SQUARE":    "SQ",
	"TERRACE":   "TER",
	"TRAIL":     "TRL",
	"PARKWAY":   "PKWY",
	"HIGHWAY":   "HWY",
	"WAY":       "WAY",
	"ALLEY":     "ALY",
	"LOOP":      "LOOP",
	"PIKE":      "PIKE",
	"PLAZA":     "PLZ",
	"POINT":     "PT",
	"RIDGE":     "RDG",
	"ROUTE":     "RTE",
	"STATION":   "STA",
}

// directionalAbbrev covers the compass-direction words that USPS
// abbreviates when they appear as a standalone word in a street name.
var directionalAbbrev = map[string]string{
	"NORTH":     "N",
	"SOUTH":     "S",
	"EAST":      "E",
	"WEST":      "W",
	"NORTHEAST": "NE",
	"NORTHWEST": "NW",
	"SOUTHEAST": "SE",
	"SOUTHWEST": "SW",
}

// unitAbbrev covers the secondary-unit designators from Pub 28
// Appendix C2 that commonly appear in a "Unit" field.
var unitAbbrev = map[string]string{
	"APARTMENT":  "APT",
	"SUITE":      "STE",
	"BUILDING":   "BLDG",
	"FLOOR":      "FL",
	"UNIT":       "UNIT",
	"ROOM":       "RM",
	"DEPARTMENT": "DEPT",
}

var punctuation = regexp.MustCompile(`[.,]`)
var spaces = regexp.MustCompile(`\s+`)

// standardizeWords uppercases s, strips punctuation, and replaces any
// whole word found in the given abbreviation table.
func standardizeWords(s string, table map[string]string) string {
	s = punctuation.ReplaceAllString(strings.ToUpper(s), "")
	words := strings.Fields(spaces.ReplaceAllString(s, " "))
	for i, w := range words {
		if abbr, ok := table[w]; ok {
			words[i] = abbr
		}
	}
	return strings.Join(words, " ")
}

// standardizeStreet abbreviates both directionals and suffixes, since
// a street name can carry both ("NORTH MAIN STREET" -> "N MAIN ST").
func standardizeStreet(s string) string {
	s = standardizeWords(s, directionalAbbrev)
	return standardizeWords(s, suffixAbbrev)
}

// Lines renders the address as the line-by-line block USPS expects on
// an envelope: recipient (if any), street + unit, city/state/zip.
func (a Address) Lines() []string {
	var lines []string
	if a.Recipient != "" {
		lines = append(lines, strings.ToUpper(a.Recipient))
	}

	street := standardizeStreet(a.Street)
	if a.Unit != "" {
		street = street + " " + standardizeWords(a.Unit, unitAbbrev)
	}
	lines = append(lines, street)

	zip := a.Zip5
	if a.Zip4 != "" {
		zip = zip + "-" + a.Zip4
	}
	lines = append(lines, fmt.Sprintf("%s %s %s",
		strings.ToUpper(a.City), strings.ToUpper(a.State), zip))

	return lines
}

// cityStateZip matches the last line of a USPS block: everything up
// to a two-letter state code and a 5 or 5+4 digit zip.
var cityStateZip = regexp.MustCompile(`^(.+?)[\s,]+([A-Z]{2})\s+(\d{5})(?:-(\d{4}))?$`)

// unitTokens holds the abbreviated forms a unit designator can take on
// a street line, e.g. "APT" or "STE", so ParseLines can recognize one
// whether or not it was standardized by Lines first.
var unitTokens = func() map[string]bool {
	set := map[string]bool{}
	for _, abbr := range unitAbbrev {
		set[abbr] = true
	}
	return set
}()

// splitUnit looks for a recognized unit designator among the words of
// a street line and, if found, returns the street with everything
// from that word on split into the unit.
func splitUnit(words []string) (street, unit string) {
	for i, w := range words {
		if unitTokens[w] {
			return strings.Join(words[:i], " "), strings.Join(words[i:], " ")
		}
	}
	return strings.Join(words, " "), ""
}

// ParseLines does the reverse of Lines: given a 2 or 3 line USPS
// block, it recovers the structured fields. The recipient line is
// only present, and only recovered, when there are 3 lines.
func ParseLines(lines []string) (Address, error) {
	var clean []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			clean = append(clean, l)
		}
	}

	var a Address
	switch len(clean) {
	case 3:
		a.Recipient = clean[0]
		clean = clean[1:]
	case 2:
		// no recipient line
	default:
		return Address{}, fmt.Errorf("expected 2 or 3 non-blank lines, got %d", len(clean))
	}

	last := strings.ToUpper(strings.TrimSpace(clean[1]))
	m := cityStateZip.FindStringSubmatch(last)
	if m == nil {
		return Address{}, fmt.Errorf("last line %q is not CITY ST ZIP", clean[1])
	}
	a.City, a.State, a.Zip5, a.Zip4 = strings.TrimSpace(m[1]), m[2], m[3], m[4]

	// The street line may carry a unit designator we recognize; if so,
	// split it off, otherwise treat the whole line as the street.
	words := strings.Fields(strings.ToUpper(strings.TrimSpace(clean[0])))
	a.Street, a.Unit = splitUnit(words)

	return a, nil
}
