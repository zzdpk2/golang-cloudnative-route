package domain

import (
	"encoding/json"
	"testing"
)

func TestNewAddress(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		a, err := NewAddress("123 George St", "Sydney", "NSW", "2000", "AU")
		assertNoError(t, err)
		assertEqual(t, a.City, "Sydney")
	})
	t.Run("empty street", func(t *testing.T) {
		_, err := NewAddress("", "Sydney", "NSW", "2000", "AU")
		if err == nil {
			t.Error("expected error")
		}
	})
	t.Run("empty state ok", func(t *testing.T) {
		_, err := NewAddress("123 St", "Sydney", "", "2000", "AU")
		assertNoError(t, err)
	})
}

func TestAddress_WithGeo(t *testing.T) {
	a, _ := NewAddress("123 St", "Sydney", "NSW", "2000", "AU")
	b := a.WithGeo(-33.8688, 151.2093)
	if a.Latitude != 0 {
		t.Error("original should not change")
	}
	if b.Latitude == 0 {
		t.Error("new should have geo")
	}
	if !a.Equals(b) {
		t.Error("text parts should be equal")
	}
}

func TestAddress_SingleLine(t *testing.T) {
	a, _ := NewAddress("123 George St", "Sydney", "NSW", "2000", "AU")
	assertEqual(t, a.SingleLine(), "123 George St, Sydney, NSW 2000, AU")
}

func TestAddress_JSON_Embedded(t *testing.T) {
	a, _ := NewAddress("123 St", "Sydney", "NSW", "2000", "AU")
	a = a.WithGeo(-33.8688, 151.2093)
	data, _ := json.Marshal(a)
	var restored Address
	json.Unmarshal(data, &restored)
	assertEqual(t, restored.Street, "123 St")
	if restored.Latitude == 0 {
		t.Error("geo should survive JSON round trip")
	}
}

func TestAddress_ValueCopy(t *testing.T) {
	a, _ := NewAddress("1 Main St", "Sydney", "NSW", "2000", "AU")
	b := a
	b.Street = "999 Other St"
	assertEqual(t, a.Street, "1 Main St")
	assertEqual(t, b.Street, "999 Other St")
}
