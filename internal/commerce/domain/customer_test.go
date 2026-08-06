package domain

import (
	"encoding/json"
	"errors"
	"testing"
)

func makeTestEmail(t *testing.T) Email {
	t.Helper()
	e, err := NewEmail("rex@test.com")
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func makeTestCustomer(t *testing.T) *Customer {
	t.Helper()
	c, err := NewCustomer("cust-001", "Rex", makeTestEmail(t))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewCustomer(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		c := makeTestCustomer(t)
		if c.Name() != "Rex" {
			t.Errorf("Name = %q", c.Name())
		}
	})

	t.Run("empty name returns ValidationError", func(t *testing.T) {
		_, err := NewCustomer("id", "", makeTestEmail(t))
		if err == nil {
			t.Fatal("expected error")
		}
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("expected ValidationError, got %T: %v", err, err)
		}
		if ve.Field != "name" {
			t.Errorf("Field = %q, want 'name'", ve.Field)
		}
	})
}

func TestCustomer_NilVsEmptySlice(t *testing.T) {
	c := makeTestCustomer(t)
	addrs := c.Addresses()

	if addrs == nil {
		t.Error("addresses should be empty slice, not nil")
	}
	if len(addrs) != 0 {
		t.Error("addresses should have length 0")
	}

	nilSlice := []Address(nil)
	emptySlice := []Address{}

	nilJSON, _ := json.Marshal(nilSlice)
	emptyJSON, _ := json.Marshal(emptySlice)

	if string(nilJSON) != "null" {
		t.Errorf("nil slice JSON = %s, want 'null'", nilJSON)
	}
	if string(emptyJSON) != "[]" {
		t.Errorf("empty slice JSON = %s, want '[]'", emptyJSON)
	}
}

func TestCustomer_DefensiveCopy(t *testing.T) {
	c := makeTestCustomer(t)
	addr, _ := NewAddress("1 Main St", "Sydney", "NSW", "2000", "AU")
	c.AddAddress(addr)

	addrs := c.Addresses()
	if len(addrs) != 1 {
		t.Fatal("should have 1 address")
	}

	addrs[0] = Address{} // See the corresponding tests for the intended behavior.
	original := c.Addresses()
	if original[0].Street != "1 Main St" {
		t.Error("defensive copy failed: modifying returned slice affected internal state")
	}
}

func TestCustomer_AddRemoveAddress(t *testing.T) {
	c := makeTestCustomer(t)
	a1, _ := NewAddress("1 Main St", "Sydney", "NSW", "2000", "AU")
	a2, _ := NewAddress("2 King St", "Melbourne", "VIC", "3000", "AU")
	a3, _ := NewAddress("3 Queen St", "Brisbane", "QLD", "4000", "AU")

	c.AddAddress(a1)
	c.AddAddress(a2)
	c.AddAddress(a3)

	if len(c.Addresses()) != 3 {
		t.Fatalf("should have 3 addresses, got %d", len(c.Addresses()))
	}

	err := c.RemoveAddress(1)
	if err != nil {
		t.Fatal(err)
	}

	addrs := c.Addresses()
	if len(addrs) != 2 {
		t.Fatalf("should have 2 addresses, got %d", len(addrs))
	}
	if addrs[0].Street != "1 Main St" {
		t.Errorf("first address = %q", addrs[0].Street)
	}
	if addrs[1].Street != "3 Queen St" {
		t.Errorf("second address = %q", addrs[1].Street)
	}

	err = c.RemoveAddress(99)
	if err == nil {
		t.Error("should error on out-of-bounds index")
	}
}

func TestCustomer_UpdateEmail_ErrorWrapping(t *testing.T) {
	c := makeTestCustomer(t)
	c.Deactivate()

	newEmail, _ := NewEmail("new@test.com")
	err := c.UpdateEmail(newEmail)

	if !errors.Is(err, ErrCustomerDeactivated) {
		t.Errorf("errors.Is should match ErrCustomerDeactivated, got: %v", err)
	}

	if err.Error() == ErrCustomerDeactivated.Error() {
		t.Error("error should have more context than just the sentinel")
	}
}

func TestCustomer_PlaceOrder_MultiLayerWrapping(t *testing.T) {
	c := makeTestCustomer(t)
	c.Deactivate()

	err := c.PlaceOrder()
	if !errors.Is(err, ErrCustomerDeactivated) {
		t.Errorf("errors.Is should match through multi-layer wrapping, got: %v", err)
	}

	var ve *ValidationError
	if errors.As(err, &ve) {
		t.Error("should not be a ValidationError")
	}
}

func TestCustomer_PlaceOrder_NoAddress(t *testing.T) {
	c := makeTestCustomer(t)

	err := c.PlaceOrder()
	if err == nil {
		t.Fatal("should error without address")
	}

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
	if ve.Field != "addresses" {
		t.Errorf("Field = %q", ve.Field)
	}
}

func TestCustomer_PlaceOrder_Success(t *testing.T) {
	c := makeTestCustomer(t)
	addr, _ := NewAddress("1 Main St", "Sydney", "NSW", "2000", "AU")
	c.AddAddress(addr)

	err := c.PlaceOrder()
	if err != nil {
		t.Fatalf("should succeed: %v", err)
	}
}

func TestBehavioralError(t *testing.T) {
	inner := errors.New("connection timeout")
	retryErr := &RetryableError{Cause: inner, RetryAfter: 5}

	if !errors.Is(retryErr, inner) {
		t.Error("errors.Is should find inner error through Unwrap")
	}

	var tempErr TemporaryError
	if !errors.As(retryErr, &tempErr) {
		t.Fatal("should match TemporaryError interface")
	}
	if !tempErr.IsTemporary() {
		t.Error("should be temporary")
	}
}

func TestCustomer_AuditLog_Defer(t *testing.T) {
	c := makeTestCustomer(t)

	t.Run("success", func(t *testing.T) {
		result, err := c.AuditLog("update")
		if err != nil {
			t.Fatal(err)
		}
		want := "AUDIT [cust-001] update: SUCCESS"
		if result != want {
			t.Errorf("result = %q, want %q", result, want)
		}
	})

	t.Run("failure", func(t *testing.T) {
		result, err := c.AuditLog("delete")
		if err == nil {
			t.Fatal("delete should fail")
		}
		if result == "" {
			t.Error("result should be set by defer even on error")
		}
		if len(result) < 5 {
			t.Error("result should contain audit info")
		}
	})
}
