package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"testing"
)

func TestCurrency_String(t *testing.T) {
	tests := []struct {
		name     string
		currency Currency
		want     string
	}{
		{"AUD", AUD, "AUD"},
		{"USD", USD, "USD"},
		{"CNY", CNY, "CNY"},
		{"Unknown", Currency(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.currency.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewMoney(t *testing.T) {
	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		m, err := NewMoney(10.50, AUD)
		assertNoError(t, err)
		assertEqual(t, m.Cents(), int64(1050))
		assertEqual(t, m.CurrencyType(), AUD)
	})
	t.Run("zero", func(t *testing.T) {
		t.Parallel()
		m, err := NewMoney(0, USD)
		assertNoError(t, err)
		if !m.IsZero() {
			t.Error("expected zero")
		}
	})
	t.Run("negative", func(t *testing.T) {
		t.Parallel()
		_, err := NewMoney(-5, AUD)
		assertErrorIs(t, err, ErrNegativeAmount)
	})
	t.Run("invalid currency", func(t *testing.T) {
		t.Parallel()
		_, err := NewMoney(10, Currency(99))
		assertErrorIs(t, err, ErrInvalidCurrency)
	})
	t.Run("float precision 19.99", func(t *testing.T) {
		t.Parallel()
		m, _ := NewMoney(19.99, AUD)
		assertEqual(t, m.Cents(), int64(1999))
	})
	t.Run("float precision 0.30", func(t *testing.T) {
		t.Parallel()
		m, _ := NewMoney(0.30, AUD)
		assertEqual(t, m.Cents(), int64(30))
	})
}

func TestMustNewMoney(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		m := MustNewMoney(10, AUD)
		assertEqual(t, m.Cents(), int64(1000))
	})
	t.Run("panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		MustNewMoney(-1, AUD)
	})
}

func TestMoney_Amount(t *testing.T) {
	m := MustNewMoney(10.50, AUD)
	if math.Abs(m.Amount()-10.50) > 0.001 {
		t.Errorf("Amount() = %f", m.Amount())
	}
}

func TestMoney_Add(t *testing.T) {
	tests := []struct {
		name    string
		a, b    Money
		want    int64
		wantErr error
	}{
		{"same currency", MustNewMoney(10, AUD), MustNewMoney(5.50, AUD), 1550, nil},
		{"add zero", MustNewMoney(10, AUD), Zero(AUD), 1000, nil},
		{"diff currency", MustNewMoney(10, AUD), MustNewMoney(5, USD), 0, ErrCurrencyMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.Add(tt.b)
			if tt.wantErr != nil {
				assertErrorIs(t, err, tt.wantErr)
				return
			}
			assertNoError(t, err)
			assertEqual(t, got.Cents(), tt.want)
		})
	}
}

func TestMoney_Subtract(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		got, err := MustNewMoney(10, AUD).Subtract(MustNewMoney(3.50, AUD))
		assertNoError(t, err)
		assertEqual(t, got.Cents(), int64(650))
	})
	t.Run("negative result", func(t *testing.T) {
		_, err := MustNewMoney(3, AUD).Subtract(MustNewMoney(5, AUD))
		assertErrorIs(t, err, ErrNegativeAmount)
	})
	t.Run("currency mismatch", func(t *testing.T) {
		_, err := MustNewMoney(10, AUD).Subtract(MustNewMoney(5, CNY))
		assertErrorIs(t, err, ErrCurrencyMismatch)
	})
}

func TestMoney_Multiply(t *testing.T) {
	assertEqual(t, MustNewMoney(9.99, AUD).Multiply(3).Cents(), int64(2997))
}

func TestMoney_Divide(t *testing.T) {
	t.Run("even", func(t *testing.T) {
		r, rem, err := MustNewMoney(10, AUD).Divide(2)
		assertNoError(t, err)
		assertEqual(t, r.Cents(), int64(500))
		assertEqual(t, rem.Cents(), int64(0))
	})
	t.Run("remainder", func(t *testing.T) {
		r, rem, err := MustNewMoney(10, AUD).Divide(3)
		assertNoError(t, err)
		assertEqual(t, r.Cents(), int64(333))
		assertEqual(t, rem.Cents(), int64(1))
	})
	t.Run("by zero", func(t *testing.T) {
		_, _, err := MustNewMoney(10, AUD).Divide(0)
		assertErrorIs(t, err, ErrDivisionByZero)
	})
}

func TestMoney_GreaterThan(t *testing.T) {
	a, b := MustNewMoney(10, AUD), MustNewMoney(5, AUD)
	gt, _ := a.GreaterThan(b)
	if !gt {
		t.Error("10 > 5")
	}
	gt, _ = b.GreaterThan(a)
	if gt {
		t.Error("5 not > 10")
	}
}

func TestMoney_Equals(t *testing.T) {
	a := MustNewMoney(10.50, AUD)
	b := MustNewMoney(10.50, AUD)
	c := MustNewMoney(10.50, USD)
	if !a.Equals(b) {
		t.Error("same")
	}
	if a.Equals(c) {
		t.Error("diff currency")
	}
}

func TestMoney_String(t *testing.T) {
	assertEqual(t, MustNewMoney(10.50, AUD).String(), "A$10.50")
	s := fmt.Sprintf("Price: %s", MustNewMoney(10.50, AUD))
	assertEqual(t, s, "Price: A$10.50")
}

func TestMoney_GoString(t *testing.T) {
	got := fmt.Sprintf("%#v", MustNewMoney(10.50, AUD))
	assertEqual(t, got, "Money{amount: 1050, currency: AUD}")
}

func TestMoney_JSON(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		data, err := json.Marshal(MustNewMoney(99.99, AUD))
		assertNoError(t, err)
		var raw map[string]interface{}
		json.Unmarshal(data, &raw)
		assertEqual(t, raw["currency"].(string), "AUD")
		if math.Abs(raw["amount"].(float64)-99.99) > 0.001 {
			t.Errorf("amount = %v", raw["amount"])
		}
	})
	t.Run("unmarshal", func(t *testing.T) {
		var m Money
		err := json.Unmarshal([]byte(`{"amount":42.50,"currency":"USD"}`), &m)
		assertNoError(t, err)
		assertEqual(t, m.Cents(), int64(4250))
		assertEqual(t, m.CurrencyType(), USD)
	})
	t.Run("unmarshal invalid currency", func(t *testing.T) {
		var m Money
		err := json.Unmarshal([]byte(`{"amount":10,"currency":"BTC"}`), &m)
		if err == nil {
			t.Error("expected error")
		}
	})
	t.Run("round trip", func(t *testing.T) {
		orig := MustNewMoney(123.45, CNY)
		data, _ := json.Marshal(orig)
		var restored Money
		json.Unmarshal(data, &restored)
		if !orig.Equals(restored) {
			t.Error("round trip failed")
		}
	})
}

func TestMoney_Allocate(t *testing.T) {
	t.Run("even", func(t *testing.T) {
		parts, err := MustNewMoney(10, AUD).Allocate(4)
		assertNoError(t, err)
		assertEqual(t, len(parts), 4)
		for _, p := range parts {
			assertEqual(t, p.Cents(), int64(250))
		}
	})
	t.Run("uneven", func(t *testing.T) {
		parts, err := MustNewMoney(10, AUD).Allocate(3)
		assertNoError(t, err)
		assertEqual(t, parts[0].Cents(), int64(334))
		assertEqual(t, parts[1].Cents(), int64(333))
		assertEqual(t, parts[2].Cents(), int64(333))
		var total int64
		for _, p := range parts {
			total += p.Cents()
		}
		assertEqual(t, total, int64(1000))
	})
	t.Run("zero divisor", func(t *testing.T) {
		_, err := MustNewMoney(10, AUD).Allocate(0)
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestSum_Variadic(t *testing.T) {
	t.Run("multiple", func(t *testing.T) {
		got, err := Sum(MustNewMoney(10, AUD), MustNewMoney(20, AUD), MustNewMoney(30, AUD))
		assertNoError(t, err)
		assertEqual(t, got.Cents(), int64(6000))
	})
	t.Run("empty", func(t *testing.T) {
		got, err := Sum()
		assertNoError(t, err)
		if !got.IsZero() {
			t.Error("expected zero")
		}
	})
	t.Run("spread slice", func(t *testing.T) {
		amounts := []Money{MustNewMoney(1, AUD), MustNewMoney(2, AUD), MustNewMoney(3, AUD)}
		got, _ := Sum(amounts...)
		assertEqual(t, got.Cents(), int64(600))
	})
	t.Run("mixed currencies", func(t *testing.T) {
		_, err := Sum(MustNewMoney(10, AUD), MustNewMoney(20, USD))
		assertErrorIs(t, err, ErrCurrencyMismatch)
	})
}

func TestPercentageDiscount(t *testing.T) {
	got := MustNewMoney(100, AUD).ApplyDiscount(PercentageDiscount(20))
	assertEqual(t, got.Cents(), int64(8000))
}

func TestFixedDiscount(t *testing.T) {
	got := MustNewMoney(100, AUD).ApplyDiscount(FixedDiscount(MustNewMoney(30, AUD)))
	assertEqual(t, got.Cents(), int64(7000))
	got = MustNewMoney(100, AUD).ApplyDiscount(FixedDiscount(MustNewMoney(200, AUD)))
	assertEqual(t, got.Cents(), int64(0))
}

// ---- helpers ----
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertErrorIs(t *testing.T, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("error = %v, want errors.Is(%v)", err, target)
	}
}

func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
