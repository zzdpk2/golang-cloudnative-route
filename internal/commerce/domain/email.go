package domain

import (
	"fmt"
)

var ErrInvalidEmail = fmt.Errorf("invalid email address")

// Email is a validated, normalised address.
//
// Normalisation is the point as much as validation: "Rex@Gmail.com " and
// "rex@gmail.com" are the same account, and if the type does not decide that,
// every comparison and every uniqueness check downstream has to remember to.
type Email struct {
	value string
}

// NewEmail normalises and validates a raw address.
//
// Normalise first, then validate the normalised form — otherwise a trailing
// space can fail a check that the stored value would have passed.
//
// Full RFC 5322 validation is a famous rabbit hole; do not go down it. Decide
// on the few cheap structural rules you actually want (exactly one @, non-empty
// on both sides, a dot in the domain) and be explicit that you chose a
// pragmatic subset. Real systems validate an address by sending mail to it.
func NewEmail(raw string) (Email, error) {
	panic("TODO")
}

func (e Email) String() string { panic("TODO") }

// Domain returns the part after the @.
func (e Email) Domain() string {
	panic("TODO")
}

// LocalPart returns the part before the @.
//
// Both this and Domain are called on an Email that has already been validated,
// which is what lets them be this simple. Notice how much a validated
// constructor buys the rest of the type.
func (e Email) LocalPart() string {
	panic("TODO")
}

func (e Email) Equals(other Email) bool { panic("TODO") }

// MaskedString is the form safe to put in a log or on a support screen:
// "rex@gmail.com" becomes "r***@gmail.com".
//
// The mask is a fixed three asterisks, not one per hidden character — length is
// information too. The tests cover a 3-character local part and a 6-character
// one and expect the same mask for both, which tells you the rule.
//
// Worth pausing on: this method exists because logging a customer's email in
// full is a privacy incident waiting for an audit. Ask which other types in
// this package hold data that should never reach a log in raw form.
func (e Email) MaskedString() string {
	panic("TODO")
}
