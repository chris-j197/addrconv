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

// suffixAbbrev covers the street suffixes and their common alternate
// spellings from USPS Pub 28 Appendix C1, mapped to the one standard
// abbreviation USPS wants on the delivery line.
var suffixAbbrev = map[string]string{
	"ALLEY": "ALY", "ALLEE": "ALY", "ALLY": "ALY",
	"ANNEX": "ANX", "ANEX": "ANX", "ANNX": "ANX",
	"ARCADE":  "ARC",
	"AVENUE":  "AVE", "AVEN": "AVE", "AVENU": "AVE", "AVNUE": "AVE",
	"BAYOO": "BYU", "BAYOU": "BYU",
	"BEACH": "BCH",
	"BEND":  "BND",
	"BLUFF": "BLF", "BLUFFS": "BLFS",
	"BOTTOM": "BTM", "BOTTM": "BTM",
	"BOULEVARD": "BLVD", "BOULV": "BLVD",
	"BRANCH": "BR", "BRNCH": "BR",
	"BRIDGE": "BRG", "BRDGE": "BRG",
	"BROOK": "BRK", "BROOKS": "BRKS",
	"BURG": "BG", "BURGS": "BGS",
	"BYPASS": "BYP", "BYPAS": "BYP",
	"CAMP": "CP",
	"CANYON": "CYN", "CANYN": "CYN", "CNYN": "CYN",
	"CAPE":     "CPE",
	"CAUSEWAY": "CSWY",
	"CENTER":   "CTR", "CENTRE": "CTR", "CENTR": "CTR", "CENTERS": "CTRS",
	"CIRCLE": "CIR", "CIRCL": "CIR", "CRCL": "CIR", "CIRCLES": "CIRS",
	"CLIFF": "CLF", "CLIFFS": "CLFS",
	"CLUB":    "CLB",
	"COMMON":  "CMN",
	"COMMONS": "CMNS",
	"CORNER":  "COR", "CORNERS": "CORS",
	"COURSE": "CRSE",
	"COURT":  "CT", "COURTS": "CTS",
	"COVE": "CV", "COVES": "CVS",
	"CREEK":    "CRK",
	"CRESCENT": "CRES", "CRSENT": "CRES", "CRSNT": "CRES",
	"CREST":      "CRST",
	"CROSSING":   "XING",
	"CROSSROAD":  "XRD",
	"CROSSROADS": "XRDS",
	"CURVE":      "CURV",
	"DALE":       "DL",
	"DAM":        "DM",
	"DIVIDE":     "DV",
	"DRIVE":      "DR", "DRIV": "DR", "DRIVES": "DRS",
	"ESTATE": "EST", "ESTATES": "ESTS",
	"EXPRESSWAY": "EXPY",
	"EXTENSION":  "EXT", "EXTNSN": "EXT", "EXTENSIONS": "EXTS",
	"FALLS": "FLS",
	"FERRY": "FRY", "FRRY": "FRY",
	"FIELD": "FLD", "FIELDS": "FLDS",
	"FLAT": "FLT", "FLATS": "FLTS",
	"FORD": "FRD", "FORDS": "FRDS",
	"FOREST": "FRST", "FORESTS": "FRST",
	"FORGE": "FRG", "FORG": "FRG", "FORGES": "FRGS",
	"FORK": "FRK", "FORKS": "FRKS",
	"FORT":    "FT",
	"FREEWAY": "FWY", "FRWAY": "FWY",
	"GARDEN": "GDN", "GRDEN": "GDN", "GARDENS": "GDNS",
	"GATEWAY": "GTWY", "GATEWY": "GTWY", "GATWAY": "GTWY",
	"GLEN": "GLN", "GLENS": "GLNS",
	"GREEN": "GRN", "GREENS": "GRNS",
	"GROVE": "GRV", "GROV": "GRV", "GROVES": "GRVS",
	"HARBOR": "HBR", "HARBR": "HBR", "HARBORS": "HBRS",
	"HAVEN":   "HVN",
	"HEIGHTS": "HTS",
	"HIGHWAY": "HWY", "HIGHWY": "HWY", "HIWAY": "HWY", "HIWY": "HWY",
	"HILL": "HL", "HILLS": "HLS",
	"HOLLOW": "HOLW", "HOLLOWS": "HOLW",
	"INLET":  "INLT",
	"ISLAND": "IS", "ISLANDS": "ISS",
	"ISLE":     "ISLE",
	"JUNCTION": "JCT", "JUNCTON": "JCT", "JUNCTIONS": "JCTS",
	"KEY": "KY", "KEYS": "KYS",
	"KNOLL": "KNL", "KNOL": "KNL", "KNOLLS": "KNLS",
	"LAKE":    "LK",
	"LAKES":   "LKS",
	"LANDING": "LNDG",
	"LANE":    "LN",
	"LIGHT":   "LGT", "LIGHTS": "LGTS",
	"LOAF": "LF",
	"LOCK": "LCK", "LOCKS": "LCKS",
	"LODGE": "LDG", "LODG": "LDG",
	"LOOP": "LOOP", "LOOPS": "LOOP",
	"MALL":  "MALL",
	"MANOR": "MNR", "MANORS": "MNRS",
	"MEADOW": "MDW", "MEADOWS": "MDWS",
	"MEWS": "MEWS",
	"MILL": "ML", "MILLS": "MLS",
	"MISSION":  "MSN",
	"MISSN":    "MSN",
	"MOTORWAY": "MTWY",
	"MOUNT":    "MT",
	"MOUNTAIN": "MTN", "MNTAIN": "MTN", "MOUNTAINS": "MTNS",
	"NECK":    "NCK",
	"ORCHARD": "ORCH", "ORCHRD": "ORCH",
	"OVAL":      "OVAL",
	"OVERPASS":  "OPAS",
	"PARK":      "PARK",
	"PARKWAY":   "PKWY",
	"PARKWY":    "PKWY",
	"PARKWAYS":  "PKWY",
	"PASS":      "PASS",
	"PASSAGE":   "PSGE",
	"PATH":      "PATH",
	"PIKE":      "PIKE",
	"PINE":      "PNE",
	"PINES":     "PNES",
	"PLACE":     "PL",
	"PLAIN":     "PLN",
	"PLAINS":    "PLNS",
	"PLAZA":     "PLZ", "PLZA": "PLZ",
	"POINT": "PT", "POINTS": "PTS",
	"PORT": "PRT", "PORTS": "PRTS",
	"PRAIRIE": "PR",
	"RADIAL":  "RADL", "RADIEL": "RADL",
	"RAMP":  "RAMP",
	"RANCH": "RNCH", "RANCHES": "RNCH",
	"RAPID": "RPD", "RAPIDS": "RPDS",
	"REST":  "RST",
	"RIDGE": "RDG", "RDGE": "RDG", "RIDGES": "RDGS",
	"RIVER": "RIV", "RIVR": "RIV",
	"ROAD": "RD", "ROADS": "RDS",
	"ROUTE": "RTE",
	"ROW":   "ROW",
	"RUE":   "RUE",
	"RUN":   "RUN",
	"SHOAL": "SHL", "SHOALS": "SHLS",
	"SHORE": "SHR", "SHOAR": "SHR", "SHORES": "SHRS", "SHOARS": "SHRS",
	"SKYWAY": "SKWY",
	"SPRING": "SPG", "SPNG": "SPG", "SPRINGS": "SPGS", "SPNGS": "SPGS",
	"SPUR":   "SPUR",
	"SQUARE": "SQ", "SQR": "SQ", "SQUARES": "SQS",
	"STATION": "STA", "STATN": "STA",
	"STRAVENUE": "STRA", "STRAVEN": "STRA",
	"STREAM": "STRM", "STREME": "STRM",
	"STREET": "ST", "STRT": "ST", "STREETS": "STS",
	"SUMMIT": "SMT", "SUMIT": "SMT",
	"TERRACE":    "TER",
	"THROUGHWAY": "TRWY",
	"TRACE":      "TRCE",
	"TRACK":      "TRAK", "TRAK": "TRAK",
	"TRAFFICWAY": "TRFY",
	"TRAIL":      "TRL",
	"TRAILER":    "TRLR",
	"TUNNEL":     "TUNL", "TUNEL": "TUNL",
	"TURNPIKE": "TPKE", "TRNPK": "TPKE",
	"UNDERPASS": "UPAS",
	"UNION":     "UN", "UNIONS": "UNS",
	"VALLEY": "VLY", "VALLY": "VLY", "VLLY": "VLY", "VALLEYS": "VLYS",
	"VIADUCT": "VIA", "VIADCT": "VIA",
	"VIEW": "VW", "VIEWS": "VWS",
	"VILLAGE": "VLG", "VILLAG": "VLG", "VILLG": "VLG", "VILLAGES": "VLGS",
	"VILLE": "VL",
	"VISTA": "VIS", "VIST": "VIS", "VSTA": "VIS",
	"WALK": "WALK", "WALKS": "WALK",
	"WALL": "WALL",
	"WAY":  "WAY", "WAYS": "WAYS",
	"WELL": "WL", "WELLS": "WLS",
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
	"BASEMENT":   "BSMT",
	"BUILDING":   "BLDG",
	"DEPARTMENT": "DEPT",
	"FLOOR":      "FL",
	"FRONT":      "FRNT",
	"HANGAR":     "HNGR",
	"LOBBY":      "LBBY",
	"LOWER":      "LOWR",
	"OFFICE":     "OFC",
	"PENTHOUSE":  "PH",
	"PIER":       "PIER",
	"REAR":       "REAR",
	"ROOM":       "RM",
	"SIDE":       "SIDE",
	"SLIP":       "SLIP",
	"SPACE":      "SPC",
	"STOP":       "STOP",
	"SUITE":      "STE",
	"TRAILER":    "TRLR",
	"UNIT":       "UNIT",
	"UPPER":      "UPPR",
}

// validStateCodes is the full set of two-letter codes USPS accepts in
// the state position: the 50 states, DC, the inhabited territories,
// and the military "state" codes used for APO/FPO/DPO addresses.
var validStateCodes = map[string]bool{
	"AL": true, "AK": true, "AZ": true, "AR": true, "CA": true,
	"CO": true, "CT": true, "DE": true, "FL": true, "GA": true,
	"HI": true, "ID": true, "IL": true, "IN": true, "IA": true,
	"KS": true, "KY": true, "LA": true, "ME": true, "MD": true,
	"MA": true, "MI": true, "MN": true, "MS": true, "MO": true,
	"MT": true, "NE": true, "NV": true, "NH": true, "NJ": true,
	"NM": true, "NY": true, "NC": true, "ND": true, "OH": true,
	"OK": true, "OR": true, "PA": true, "RI": true, "SC": true,
	"SD": true, "TN": true, "TX": true, "UT": true, "VT": true,
	"VA": true, "WA": true, "WV": true, "WI": true, "WY": true,
	"DC": true,
	"AS": true, "GU": true, "MP": true, "PR": true, "VI": true,
	"AA": true, "AE": true, "AP": true,
}

// ValidStateCode reports whether code is a two-letter state,
// territory, or military "state" abbreviation that USPS delivers to.
// The check is case-insensitive.
func ValidStateCode(code string) bool {
	return validStateCodes[strings.ToUpper(code)]
}

// Validate checks the fields that have to take one specific form for
// USPS to deliver the piece, rather than just a preferred spelling.
// Right now that's only the state code; Lines will happily abbreviate
// whatever it's given, but a state that isn't on USPS's list means the
// address can't actually be delivered.
func (a Address) Validate() error {
	if !ValidStateCode(a.State) {
		return fmt.Errorf("invalid state code %q", a.State)
	}
	return nil
}

var punctuation = regexp.MustCompile(`[.,]`)
var spaces = regexp.MustCompile(`\s+`)

// multiWordPhrases holds fixed word sequences that have to be matched
// as a unit before the per-word tables run. A two-word directional
// like "NORTH EAST" has no single-word entry in directionalAbbrev, and
// the various ways people write a PO Box line ("Post Office Box",
// "P O Box") all need to collapse to the same "PO BOX" USPS wants,
// which a word-by-word lookup can't do. Longer phrases are listed
// first so a 3-word match is tried before a shorter one could steal
// part of it.
var multiWordPhrases = []struct {
	words []string
	repl  string
}{
	{[]string{"POST", "OFFICE", "BOX"}, "PO BOX"},
	{[]string{"P", "O", "BOX"}, "PO BOX"},
	{[]string{"NORTH", "EAST"}, "NE"},
	{[]string{"NORTH", "WEST"}, "NW"},
	{[]string{"SOUTH", "EAST"}, "SE"},
	{[]string{"SOUTH", "WEST"}, "SW"},
}

// replacePhrases scans words left to right and substitutes any
// multiWordPhrases match it finds, passing through anything else
// unchanged.
func replacePhrases(words []string) []string {
	out := make([]string, 0, len(words))
	for i := 0; i < len(words); {
		matched := false
		for _, p := range multiWordPhrases {
			end := i + len(p.words)
			if end > len(words) {
				continue
			}
			if equalWords(words[i:end], p.words) {
				out = append(out, strings.Fields(p.repl)...)
				i = end
				matched = true
				break
			}
		}
		if !matched {
			out = append(out, words[i])
			i++
		}
	}
	return out
}

func equalWords(a, b []string) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

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
// It also folds multi-word phrases like "NORTH EAST" or "POST OFFICE
// BOX" down to their standard form before the single-word tables run.
func standardizeStreet(s string) string {
	s = punctuation.ReplaceAllString(strings.ToUpper(s), "")
	words := strings.Fields(spaces.ReplaceAllString(s, " "))
	s = strings.Join(replacePhrases(words), " ")
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
	if !ValidStateCode(a.State) {
		return Address{}, fmt.Errorf("last line %q has unrecognized state code %q", clean[1], a.State)
	}

	// The street line may carry a unit designator we recognize; if so,
	// split it off, otherwise treat the whole line as the street.
	words := strings.Fields(strings.ToUpper(strings.TrimSpace(clean[0])))
	a.Street, a.Unit = splitUnit(words)

	return a, nil
}
