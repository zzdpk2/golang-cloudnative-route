package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewQuantity(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		q, err := NewQuantity(5)
		assertNoError(t, err)
		assertEqual(t, q.Value(), 5)
	})
	t.Run("zero invalid", func(t *testing.T) {
		_, err := NewQuantity(0)
		assertErrorIs(t, err, ErrInvalidQuantity)
	})
	t.Run("negative invalid", func(t *testing.T) {
		_, err := NewQuantity(-1)
		assertErrorIs(t, err, ErrInvalidQuantity)
	})
}

func TestQuantity_Add(t *testing.T) {
	a, _ := NewQuantity(3)
	b, _ := NewQuantity(2)
	assertEqual(t, a.Add(b).Value(), 5)
}

func TestQuantity_Subtract(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		a, _ := NewQuantity(5)
		b, _ := NewQuantity(2)
		got, err := a.Subtract(b)
		assertNoError(t, err)
		assertEqual(t, got.Value(), 3)
	})
	t.Run("result zero invalid", func(t *testing.T) {
		a, _ := NewQuantity(3)
		b, _ := NewQuantity(3)
		_, err := a.Subtract(b)
		if err == nil {
			t.Error("quantity 0 should be invalid")
		}
	})
}

func TestNewDateRange(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	t.Run("valid", func(t *testing.T) {
		dr, err := NewDateRange(now, tomorrow)
		assertNoError(t, err)
		assertEqual(t, dr.Start(), now)
	})
	t.Run("end before start", func(t *testing.T) {
		_, err := NewDateRange(tomorrow, now)
		assertErrorIs(t, err, ErrInvalidDateRange)
	})
	t.Run("same time", func(t *testing.T) {
		_, err := NewDateRange(now, now)
		if !errors.Is(err, ErrInvalidDateRange) {
			t.Errorf("same time should be invalid")
		}
	})
}

func TestDateRange_Duration(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC)
	dr, _ := NewDateRange(start, end)
	assertEqual(t, dr.Days(), 7)
	if dr.Duration() != 7*24*time.Hour {
		t.Errorf("Duration = %v", dr.Duration())
	}
}

func TestDateRange_Contains(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	dr, _ := NewDateRange(start, end)

	mid := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	before := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	after := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	if !dr.Contains(mid) {
		t.Error("should contain mid")
	}
	if !dr.Contains(start) {
		t.Error("should contain start")
	}
	if !dr.Contains(end) {
		t.Error("should contain end")
	}
	if dr.Contains(before) {
		t.Error("not before")
	}
	if dr.Contains(after) {
		t.Error("not after")
	}
}

func TestDateRange_Overlaps(t *testing.T) {
	jan := mustDR(t, 2025, 1, 1, 2025, 1, 31)
	feb := mustDR(t, 2025, 2, 1, 2025, 2, 28)
	mid := mustDR(t, 2025, 1, 15, 2025, 2, 15)

	if jan.Overlaps(feb) {
		t.Error("jan/feb no overlap")
	}
	if !jan.Overlaps(mid) {
		t.Error("jan/mid overlap")
	}
	if !mid.Overlaps(feb) {
		t.Error("mid/feb overlap")
	}
}

func TestDateRange_Extend(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	dr, _ := NewDateRange(start, end)
	ext := dr.Extend(7 * 24 * time.Hour)
	wantEnd := time.Date(2025, 2, 7, 0, 0, 0, 0, time.UTC)
	assertEqual(t, ext.End(), wantEnd)
	assertEqual(t, dr.End(), end) // original unchanged
}

func mustDR(t *testing.T, sy, sm, sd, ey, em, ed int) DateRange {
	t.Helper()
	s := time.Date(sy, time.Month(sm), sd, 0, 0, 0, 0, time.UTC)
	e := time.Date(ey, time.Month(em), ed, 0, 0, 0, 0, time.UTC)
	dr, err := NewDateRange(s, e)
	if err != nil {
		t.Fatal(err)
	}
	return dr
}
