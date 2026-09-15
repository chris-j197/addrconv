# addrconv

Most systems store an address as a handful of fields: recipient, street,
city, state, zip. What actually goes on an envelope is different - USPS
wants an uppercase block with punctuation stripped and street types and
directionals abbreviated per Publication 28 ("N MAIN ST", not "North Main
Street."). This converts between the two, in both directions.

## Example

Structured to USPS block:

```
$ echo '{"recipient":"Jane Doe","street":"123 North Main Street","unit":"Apartment 4B","city":"Springfield","state":"IL","zip5":"62704"}' | go run . encode
JANE DOE
123 N MAIN ST APT 4B
SPRINGFIELD IL 62704
```

USPS block back to structured:

```
$ printf 'JANE DOE\n123 N MAIN ST APT 4B\nSPRINGFIELD IL 62704\n' | go run . decode
{
  "recipient": "JANE DOE",
  "street": "123 N MAIN ST",
  "unit": "APT 4B",
  "city": "SPRINGFIELD",
  "state": "IL",
  "zip5": "62704"
}
```

The recipient line is optional - a 2-line block (street, then city/state/
zip) decodes fine without one.

## Batch conversion

Add `-json` to either command to convert a whole array at once instead of
a single address. `encode -json` takes a JSON array of Address objects and
returns a JSON array of line blocks; `decode -json` takes a JSON array of
line blocks and returns a JSON array of Address objects.

```
$ echo '[{"street":"44 East Oak Avenue","city":"Reno","state":"NV","zip5":"89501"}]' | go run . encode -json
[
  [
    "44 E OAK AVE",
    "RENO NV 89501"
  ]
]
```

## What it does not do

It doesn't validate that an address exists, doesn't look up ZIP+4 codes,
and doesn't handle anything outside the US. It's a formatting converter,
not a verification service - see the roadmap in the repo for what's
planned.

## Usage as a library

```go
a := Address{Street: "44 East Oak Avenue", City: "Reno", State: "NV", Zip5: "89501"}
for _, line := range a.Lines() {
    fmt.Println(line)
}
```

## Building

Standard library only, nothing to fetch:

```
go build .
```
