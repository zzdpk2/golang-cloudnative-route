package domain

import "fmt"

// GeoLocation is optional coordinates for an address.
type GeoLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Address is a postal address, optionally geocoded.
//
// Unlike Money and Email, this type's fields are exported, because the JSON
// round-trip test marshals it directly. That is a real trade-off and you should
// notice it: exported fields mean nothing stops a caller from building an
// Address that NewAddress would have rejected. Ask yourself what the type gains
// from the constructor once that door is open, and what it would cost to close
// it.
//
// GeoLocation is *embedded*, not a named field. That is why a.Latitude works
// even though Address has no Latitude field, and why the JSON tag on the
// embedded field behaves differently from a tag on a normal one. Marshal an
// Address and look at the shape before you assume.
type Address struct {
	Street      string `json:"street"`
	City        string `json:"city"`
	State       string `json:"state"`
	Postcode    string `json:"postcode"`
	Country     string `json:"country"`
	GeoLocation `json:"geo,omitempty"`
}

// NewAddress validates the postal parts. Decide which of the five are genuinely
// required — an address with no state is normal in much of the world, and a
// platform doing cross-border business will meet those customers.
func NewAddress(street, city, state, postcode, country string) (Address, error) {
	panic("TODO")
}

// WithGeo returns a copy carrying coordinates, leaving the receiver untouched.
//
// The test asserts the original still has a zero Latitude afterwards. With a
// value receiver you get that for free — but only because every field of
// Address happens to be a value type. Work out what would break here the day
// somebody adds a []string of delivery notes.
func (a Address) WithGeo(lat, lng float64) Address {
	panic("TODO")
}

// SingleLine renders the address on one line:
//
//	"123 George St, Sydney, NSW 2000, AU"
//
// Look closely at the separators — comma between most parts, but a single space
// between state and postcode.
func (a Address) SingleLine() string {
	panic("TODO")
}

// Equals compares the postal parts only.
//
// The test builds an address, geocodes a copy, and asserts the two are still
// equal. So coordinates are *not* part of identity here: two addresses that
// name the same place are the same address whether or not anyone has looked up
// its coordinates.
//
// That also means Go's == would give the wrong answer, which is exactly why
// this method exists — the opposite conclusion from Money.Equals. Make sure you
// can explain why the two types differ.
func (a Address) Equals(other Address) bool {
	panic("TODO")
}

func (a Address) String() string { panic("TODO") }

var _ fmt.Stringer = Address{}
