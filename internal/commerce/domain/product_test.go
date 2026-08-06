package domain

import (
	"encoding/json"
	"testing"
)

func makeTestProduct(t *testing.T) *Product {
	t.Helper()
	p, err := NewProduct("prod-001", "Go Book", MustNewMoney(29.99, AUD), "books")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestNewProduct(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := makeTestProduct(t)
		if p.ID() != "prod-001" {
			t.Errorf("ID = %q", p.ID())
		}
		if p.Status() != ProductDraft {
			t.Error("should be draft")
		}
		if p.CreatedAt().IsZero() {
			t.Error("createdAt")
		}
	})
	t.Run("empty id", func(t *testing.T) {
		_, err := NewProduct("", "Test", MustNewMoney(10, AUD), "cat")
		if err == nil {
			t.Error("expected error")
		}
	})
	t.Run("empty name", func(t *testing.T) {
		_, err := NewProduct("id", "", MustNewMoney(10, AUD), "cat")
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestProduct_StatusTransitions(t *testing.T) {
	t.Run("draft→active", func(t *testing.T) {
		p := makeTestProduct(t)
		if err := p.Activate(); err != nil {
			t.Fatal(err)
		}
		if p.Status() != ProductActive {
			t.Error("should be active")
		}
	})
	t.Run("active→discontinued", func(t *testing.T) {
		p := makeTestProduct(t)
		p.Activate()
		if err := p.Discontinue(); err != nil {
			t.Fatal(err)
		}
		if p.Status() != ProductDiscontinued {
			t.Error("should be discontinued")
		}
	})
	t.Run("cannot re-activate", func(t *testing.T) {
		p := makeTestProduct(t)
		p.Activate()
		if err := p.Activate(); err == nil {
			t.Error("should fail")
		}
	})
	t.Run("cannot discontinue draft", func(t *testing.T) {
		p := makeTestProduct(t)
		if err := p.Discontinue(); err == nil {
			t.Error("should fail")
		}
	})
}

func TestProduct_Metadata(t *testing.T) {
	p := makeTestProduct(t)
	p.SetMetadata("weight", 1.5)
	p.SetMetadata("tags", []string{"go", "programming"})
	p.SetMetadata("featured", true)

	t.Run("get existing", func(t *testing.T) {
		v, ok := p.GetMetadata("weight")
		if !ok {
			t.Fatal("not found")
		}
		w, ok := v.(float64)
		if !ok {
			t.Fatal("not float64")
		}
		if w != 1.5 {
			t.Errorf("weight = %f", w)
		}
	})
	t.Run("get missing", func(t *testing.T) {
		_, ok := p.GetMetadata("color")
		if ok {
			t.Error("should not find")
		}
	})
	t.Run("generic get", func(t *testing.T) {
		v, ok := GetMetadataAs[bool](p, "featured")
		if !ok || !v {
			t.Error("should be true")
		}
	})
	t.Run("generic wrong type", func(t *testing.T) {
		_, ok := GetMetadataAs[string](p, "weight")
		if ok {
			t.Error("should fail")
		}
	})
}

func TestProduct_Equals(t *testing.T) {
	p1 := makeTestProduct(t)
	p2 := makeTestProduct(t)
	p3, _ := NewProduct("prod-002", "Other", MustNewMoney(99, AUD), "other")
	if !p1.Equals(p2) {
		t.Error("same ID")
	}
	if p1.Equals(p3) {
		t.Error("diff ID")
	}
	if p1.Equals(nil) {
		t.Error("nil")
	}
}

func TestDescribeItem(t *testing.T) {
	p := makeTestProduct(t)
	tests := []struct {
		name string
		item any
		want string
	}{
		{"product", p, "Product: Go Book"},
		{"string", "hello", "Raw: hello"},
		{"int", 42, "Unknown item"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DescribeItem(tt.item)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNilInterfaceTrap(t *testing.T) {
	p := makeTestProduct(t)
	if err := ValidateProduct(p); err != nil {
		t.Fatal(err)
	}

	err := CheckProduct(p)
	if err == nil {
		t.Fatal("nil interface trap: err is NOT nil because interface={type:*ProductError, value:nil}")
	}
	t.Run("calling Error() panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		_ = err.Error()
	})
}

func TestProduct_JSON(t *testing.T) {
	p := makeTestProduct(t)
	p.Activate()
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var dto ProductDTO
	json.Unmarshal(data, &dto)
	if dto.ID != "prod-001" {
		t.Errorf("ID = %q", dto.ID)
	}
	if dto.Status != "active" {
		t.Errorf("Status = %q", dto.Status)
	}
}
