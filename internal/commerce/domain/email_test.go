package domain

import (
	"errors"
	"testing"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid", "Rex@Gmail.com", "rex@gmail.com", false},
		{"spaces", "  user@test.com  ", "user@test.com", false},
		{"no @", "invalid", "", true},
		{"@ at start", "@domain.com", "", true},
		{"@ at end", "user@", "", true},
		{"no dot after @", "user@localhost", "", true},
		{"subdomain", "user@mail.test.com", "user@mail.test.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if !errors.Is(err, ErrInvalidEmail) {
					t.Errorf("should wrap ErrInvalidEmail")
				}
				return
			}
			assertNoError(t, err)
			assertEqual(t, email.String(), tt.want)
		})
	}
}

func TestEmail_Domain(t *testing.T) {
	e, _ := NewEmail("rex@gmail.com")
	assertEqual(t, e.Domain(), "gmail.com")
}

func TestEmail_LocalPart(t *testing.T) {
	e, _ := NewEmail("rex@gmail.com")
	assertEqual(t, e.LocalPart(), "rex")
}

func TestEmail_Equals(t *testing.T) {
	a, _ := NewEmail("Rex@Gmail.com")
	b, _ := NewEmail("rex@gmail.com")
	if !a.Equals(b) {
		t.Error("same email should be equal")
	}
}

func TestEmail_MaskedString(t *testing.T) {
	tests := []struct {
		name, email, want string
	}{
		{"ascii", "rex@gmail.com", "r***@gmail.com"},
		{"long local part", "jordan@qq.com", "j***@qq.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, _ := NewEmail(tt.email)
			assertEqual(t, e.MaskedString(), tt.want)
		})
	}
}
