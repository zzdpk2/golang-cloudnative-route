package language

import (
	"errors"
	"fmt"
	"testing"
)

func TestDomainError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *DomainError
		want string
	}{
		{"basic", &DomainError{Code: CodeNotFound, Message: "user not found"},
			"[NOT_FOUND] user not found"},
		{"with field", &DomainError{Code: CodeValidation, Field: "email", Message: "invalid format"},
			"[VALIDATION] email: invalid format"},
		{"with cause", &DomainError{Code: CodeInternal, Message: "failed", Cause: fmt.Errorf("db down")},
			"[INTERNAL] failed: db down"},
		{"with field and cause", &DomainError{Code: CodeValidation, Field: "email", Message: "invalid format", Cause: fmt.Errorf("missing @")},
			"[VALIDATION] email: invalid format: missing @"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDomainError_Is(t *testing.T) {
	err := NewNotFound("User", "123")
	sentinel := &DomainError{Code: CodeNotFound}

	if !errors.Is(err, sentinel) {
		t.Error("should match by code")
	}

	other := &DomainError{Code: CodeConflict}
	if errors.Is(err, other) {
		t.Error("different code should not match")
	}
}

func TestDomainError_Unwrap(t *testing.T) {
	inner := fmt.Errorf("connection refused")
	err := NewInternal(inner)

	if !errors.Is(err, inner) {
		t.Error("errors.Is should find inner through Unwrap")
	}
}

func TestDomainError_As(t *testing.T) {
	err := NewValidation("email", "invalid")
	wrapped := fmt.Errorf("create user: %w", err)

	var de *DomainError
	if !errors.As(wrapped, &de) {
		t.Fatal("errors.As should find DomainError through wrap")
	}
	if de.Code != CodeValidation {
		t.Errorf("Code = %s", de.Code)
	}
	if de.Field != "email" {
		t.Errorf("Field = %s", de.Field)
	}
}

func TestDomainError_DeepWrapping(t *testing.T) {
	base := NewNotFound("Order", "99")
	layer1 := fmt.Errorf("repository: %w", base)
	layer2 := fmt.Errorf("service: %w", layer1)
	layer3 := fmt.Errorf("handler: %w", layer2)

	if !errors.Is(layer3, &DomainError{Code: CodeNotFound}) {
		t.Error("should find through 4 layers")
	}

	var de *DomainError
	if !errors.As(layer3, &de) {
		t.Fatal("should extract DomainError")
	}
	if de.Code != CodeNotFound {
		t.Errorf("Code = %s", de.Code)
	}
}

// ---- MultiError Tests ----

func TestMultiError(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		m := NewMultiError()
		if m.HasErrors() {
			t.Error("should have no errors")
		}
		if m.ToError() != nil {
			t.Error("ToError should return nil for empty")
		}
	})

	t.Run("single error", func(t *testing.T) {
		m := NewMultiError()
		m.Add(fmt.Errorf("first"))
		if !m.HasErrors() {
			t.Error("should have errors")
		}
		if m.ToError() == nil {
			t.Error("should return error")
		}
	})

	t.Run("nil error ignored", func(t *testing.T) {
		m := NewMultiError()
		m.Add(nil) // should be ignored
		if m.HasErrors() {
			t.Error("nil should not be counted")
		}
	})

	t.Run("error message format", func(t *testing.T) {
		m := NewMultiError()
		m.Add(fmt.Errorf("a"))
		m.Add(fmt.Errorf("b"))
		m.Add(fmt.Errorf("c"))
		got := m.Error()
		want := "3 errors: a; b; c"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestMultiError_UnwrapMulti(t *testing.T) {
	notFound := NewNotFound("Order", "1")
	conflict := NewConflict("duplicate")

	m := NewMultiError()
	m.Add(notFound)
	m.Add(conflict)

	err := m.ToError()

	if !errors.Is(err, &DomainError{Code: CodeNotFound}) {
		t.Error("should find NotFound in multi error")
	}
	if !errors.Is(err, &DomainError{Code: CodeConflict}) {
		t.Error("should find Conflict in multi error")
	}
	if errors.Is(err, &DomainError{Code: CodeTimeout}) {
		t.Error("should not find Timeout")
	}
}

func TestMultiError_ErrorsReturnsCopy(t *testing.T) {
	m := NewMultiError()
	m.Add(fmt.Errorf("first"))
	m.Add(fmt.Errorf("second"))

	got := m.Errors()
	if len(got) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(got))
	}

	got[0] = fmt.Errorf("mutated")

	again := m.Errors()
	if again[0].Error() != "first" {
		t.Fatalf("Errors should return a defensive copy, got %q", again[0].Error())
	}
}

func TestMultiError_ConcurrentAdd(t *testing.T) {
	m := NewMultiError()
	const n = 100

	done := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			m.Add(fmt.Errorf("err %d", i))
			done <- struct{}{}
		}()
	}

	for i := 0; i < n; i++ {
		<-done
	}

	if len(m.Errors()) != n {
		t.Fatalf("expected %d errors, got %d", n, len(m.Errors()))
	}
}

// ---- ErrorGroup Tests ----

func TestErrorGroup(t *testing.T) {
	t.Run("all succeed", func(t *testing.T) {
		g := NewErrorGroup()
		for i := 0; i < 5; i++ {
			g.Go(func() error { return nil })
		}
		if err := g.Wait(); err != nil {
			t.Errorf("all should succeed: %v", err)
		}
	})

	t.Run("some fail", func(t *testing.T) {
		g := NewErrorGroup()
		g.Go(func() error { return nil })
		g.Go(func() error { return fmt.Errorf("task 2 failed") })
		g.Go(func() error { return nil })
		g.Go(func() error { return fmt.Errorf("task 4 failed") })

		err := g.Wait()
		if err == nil {
			t.Fatal("should have errors")
		}

		var me *MultiError
		if !errors.As(err, &me) {
			t.Fatal("should be MultiError")
		}
		if len(me.Errors()) != 2 {
			t.Errorf("expected 2 errors, got %d", len(me.Errors()))
		}
	})

	t.Run("all fail", func(t *testing.T) {
		g := NewErrorGroup()
		for i := 0; i < 10; i++ {
			i := i
			g.Go(func() error { return fmt.Errorf("fail %d", i) })
		}
		err := g.Wait()
		var me *MultiError
		errors.As(err, &me)
		if len(me.Errors()) != 10 {
			t.Errorf("expected 10 errors, got %d", len(me.Errors()))
		}
	})
}

// ---- Wrap Helpers ----

func TestWrap(t *testing.T) {
	t.Run("wraps error", func(t *testing.T) {
		inner := fmt.Errorf("connection refused")
		got := Wrap(inner, "save order")
		if got.Error() != "save order: connection refused" {
			t.Errorf("got %q", got.Error())
		}
		if !errors.Is(got, inner) {
			t.Error("should unwrap to inner")
		}
	})

	t.Run("nil passes through", func(t *testing.T) {
		got := Wrap(nil, "something")
		if got != nil {
			t.Error("wrapping nil should return nil")
		}
	})
}

func TestWrapIf(t *testing.T) {
	err := fmt.Errorf("base")
	got := WrapIf(err, true, "wrapped")
	if got.Error() != "wrapped: base" {
		t.Errorf("got %q", got.Error())
	}

	got2 := WrapIf(err, false, "wrapped")
	if got2.Error() != "base" {
		t.Errorf("got %q", got2.Error())
	}

	if got := WrapIf(nil, true, "wrapped"); got != nil {
		t.Errorf("nil should pass through, got %v", got)
	}
}

func TestRecover(t *testing.T) {
	err := fmt.Errorf("outer: %w", NewNotFound("Order", "1"))
	de, ok := Recover[*DomainError](err)
	if !ok {
		t.Fatal("should recover DomainError")
	}
	if de.Code != CodeNotFound {
		t.Errorf("Code = %s", de.Code)
	}

	_, ok = Recover[*MultiError](err)
	if ok {
		t.Error("should not find MultiError")
	}
}

// ---- ValidationBuilder Tests ----

func TestValidationBuilder(t *testing.T) {
	t.Run("all valid", func(t *testing.T) {
		err := NewValidationBuilder("User").
			CheckNotEmpty("Rex", "name").
			CheckNotEmpty("rex@test.com", "email").
			CheckRange(25, 0, 150, "age").
			Build()
		if err != nil {
			t.Errorf("should be valid: %v", err)
		}
	})

	t.Run("multiple failures", func(t *testing.T) {
		err := NewValidationBuilder("User").
			CheckNotEmpty("", "name").
			CheckNotEmpty("", "email").
			CheckRange(-1, 0, 150, "age").
			Build()

		if err == nil {
			t.Fatal("should have errors")
		}

		var me *MultiError
		if !errors.As(err, &me) {
			t.Fatal("should be MultiError")
		}
		if len(me.Errors()) != 3 {
			t.Errorf("expected 3 validation errors, got %d", len(me.Errors()))
		}

		for _, e := range me.Errors() {
			if !errors.Is(e, &DomainError{Code: CodeValidation}) {
				t.Errorf("expected validation error, got: %v", e)
			}
		}
	})

	t.Run("chaining", func(t *testing.T) {
		vb := NewValidationBuilder("Product")
		result := vb.CheckNotEmpty("Widget", "name").CheckRange(100, 1, 1000, "price")
		if result != vb {
			t.Error("Check should return same builder for chaining")
		}
	})
}
