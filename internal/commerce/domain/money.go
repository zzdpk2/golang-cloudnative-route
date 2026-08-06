package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Currency identifies the unit an amount is denominated in. The zero value is
// deliberately not a real currency, so a Money that nobody initialised cannot
// silently pass for AUD.
type Currency int

const (
	_ Currency = iota
	AUD
	USD
	CNY
)

var currencyNames = map[Currency]string{
	AUD: "AUD",
	USD: "USD",
	CNY: "CNY",
}

var currencySymbols = map[Currency]string{
	AUD: "A$",
	USD: "$",
	CNY: "¥",
}

// String returns the ISO code. An unrecognised Currency, including the zero
// value, has to produce something — decide what, and stay consistent with
// Symbol below.
func (c Currency) String() string {
	panic("TODO")
}

// Symbol returns the display symbol.
func (c Currency) Symbol() string {
	panic("TODO")
}

var (
	ErrNegativeAmount   = errors.New("amount cannot be negative")
	ErrCurrencyMismatch = errors.New("currency mismatch")
	ErrDivisionByZero   = errors.New("division by zero")
	ErrInvalidCurrency  = errors.New("invalid currency")
)

// Money is an amount in a currency, stored as an integer number of minor units
// (cents), never as a float.
//
// This is the single most important design decision in the package. Compute
// 0.1 + 0.2 in a Go playground and look at the result: binary floating point
// cannot represent most decimal fractions exactly, so float money accumulates
// error, and money that does not add up is money someone loses. Every real
// payment system stores integers.
//
// Money is immutable: every operation returns a new value, none mutates the
// receiver. That is why the methods below have value receivers.
type Money struct {
	amount   int64
	currency Currency
}

// NewMoney converts a decimal amount into minor units.
//
// It rejects a negative amount and an unknown currency. The conversion from
// float64 to int64 is where precision goes to die: 19.99 does not land exactly
// on 1999 when you simply multiply and truncate. Work out what 19.99*100
// actually is in float64 before you choose between truncation and rounding.
func NewMoney(amount float64, currency Currency) (Money, error) {
	panic("TODO")
}

// MustNewMoney is NewMoney for values known good at compile time, and panics
// otherwise. It exists for tests and package-level variables.
//
// A Must* wrapper is only acceptable when a failure means the program itself is
// wrong. Never reach for it on a value that came from a user.
func MustNewMoney(amount float64, currency Currency) Money {
	panic("TODO")
}

// Zero is the additive identity for a currency.
func Zero(currency Currency) Money { panic("TODO") }

// Amount returns the decimal value, for display and serialisation only.
//
// Note that this hands back the float you spent NewMoney escaping. Never use
// the result for arithmetic; use Cents.
func (m Money) Amount() float64 {
	panic("TODO")
}

func (m Money) Cents() int64           { panic("TODO") }
func (m Money) CurrencyType() Currency { panic("TODO") }

// Add returns the sum, and fails on mismatched currencies.
//
// Refusing to add AUD to USD is the entire reason this type exists. A plain
// int64 would have let it through.
func (m Money) Add(other Money) (Money, error) {
	panic("TODO")
}

// Subtract returns the difference.
//
// Decide what a result below zero means here. Money forbids negative amounts at
// construction — does that rule apply to the outcome of a subtraction too, and
// is a refund a negative amount or a separate concept?
func (m Money) Subtract(other Money) (Money, error) {
	panic("TODO")
}

// Multiply scales the amount by a whole factor. Scaling cannot change the
// currency, so unlike Add it cannot fail.
func (m Money) Multiply(factor int) Money {
	panic("TODO")
}

// Divide splits the amount into equal shares, returning the share and what is
// left over.
//
// The remainder is not an afterthought: A$10.00 / 3 is 333 cents each with 1
// cent left, and that cent has to be accounted for rather than evaporate. A
// system that drops it is a system whose ledger does not balance.
//
// Dividing by zero returns ErrDivisionByZero.
func (m Money) Divide(divisor int) (result Money, remainder Money, err error) {
	panic("TODO")
}

func (m Money) IsZero() bool { panic("TODO") }

// GreaterThan compares two amounts, and fails on mismatched currencies for the
// same reason Add does — there is no true answer to "is A$5 more than US$4".
func (m Money) GreaterThan(other Money) (bool, error) {
	panic("TODO")
}

// Equals reports value equality: same amount and same currency.
//
// Money contains only comparable fields, so Go's == would already do this.
// Write the method anyway and then ask why: what happens to every call site the
// day somebody adds a slice field to this struct?
func (m Money) Equals(other Money) bool {
	panic("TODO")
}

// String formats for humans: symbol, then the decimal amount with exactly two
// digits after the point, e.g. "A$10.50".
//
// Two digits always, including "A$10.00". Trailing-zero trimming is what
// %v would give you, and it is wrong for money.
func (m Money) String() string {
	panic("TODO")
}

func (m Money) GoString() string { panic("TODO") }

// ---- JSON ----

// moneyJSON is the wire shape. Money's own fields are unexported, so the
// default marshaller would emit {} — that is what this exists to fix.
type moneyJSON struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// MarshalJSON emits the wire shape above.
func (m Money) MarshalJSON() ([]byte, error) {
	panic("TODO")
}

// UnmarshalJSON parses the wire shape back, rejecting an unknown currency.
//
// This is the one method on Money with a pointer receiver, because it has to
// modify the value. Convince yourself why a value receiver could not work here,
// and check what the two `var _ json.Marshaler` assertions at the bottom of the
// file are really asserting.
func (m *Money) UnmarshalJSON(data []byte) error {
	panic("TODO")
}

// parseCurrency maps an ISO code back to a Currency, returning
// ErrInvalidCurrency for anything unrecognised.
func parseCurrency(s string) (Currency, error) {
	panic("TODO")
}

// Allocate splits the amount into n parts that sum back to exactly the
// original, distributing any remainder one minor unit at a time from the front.
//
// A$10.00 into 3 gives 334, 333, 333 — not 333, 333, 333 with a cent lost, and
// not 334, 334, 332. This is the classic order-splitting problem from L15.3,
// and the property to hold onto is: the parts must always sum to the whole.
//
// n must be positive.
func (m Money) Allocate(n int) ([]Money, error) {
	panic("TODO")
}

// Max returns the larger of two amounts.
func Max(a, b Money) (Money, error) { panic("TODO") }

// Sum totals any number of amounts.
//
// The empty call is the interesting case: Sum() has no currency to report, yet
// it must succeed and return zero. Look at what the test asserts, then think
// about why an empty sum is well defined at all.
func Sum(amounts ...Money) (Money, error) {
	panic("TODO")
}

// DiscountFunc transforms an amount. Making a discount a function rather than a
// number is what lets the promotion engine in L15.1 compose them.
type DiscountFunc func(Money) Money

// PercentageDiscount returns a discount taking percent off, so
// PercentageDiscount(20) applied to A$100 yields A$80.
//
// The percentage is a float and the amount is an integer, so somewhere in here
// a rounding decision gets made. Whoever is rounding, make it deliberate:
// rounding a discount in the customer's favour versus the merchant's is a
// business decision, not an implementation detail.
func PercentageDiscount(percent float64) DiscountFunc {
	panic("TODO")
}

// FixedDiscount returns a discount subtracting a fixed amount.
//
// DiscountFunc cannot report an error, but Subtract can fail. Decide what this
// closure does with a mismatched currency, and whether that limitation means
// the DiscountFunc signature is wrong.
func FixedDiscount(discount Money) DiscountFunc {
	panic("TODO")
}

func (m Money) ApplyDiscount(fn DiscountFunc) Money { panic("TODO") }

var _ fmt.Stringer = Money{}
var _ fmt.GoStringer = Money{}
var _ json.Marshaler = Money{}
var _ json.Unmarshaler = (*Money)(nil)
