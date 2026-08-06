package performance

import (
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"testing"
)

func sampleDTO() domain.ProductDTO {
	return domain.ProductDTO{
		ID:       "p1",
		Name:     "Mug",
		Price:    9.99,
		Currency: "AUD",
		Category: "kitchen",
		Status:   "active",
	}
}

func TestInspectStruct_ReportsEveryField(t *testing.T) {
	fields := InspectStruct(sampleDTO())
	if len(fields) == 0 {
		t.Fatal("InspectStruct returned nothing for a ProductDTO")
	}

	byName := map[string]FieldInfo{}
	for _, f := range fields {
		byName[f.Name] = f
	}

	id, ok := byName["ID"]
	if !ok {
		t.Fatalf("no ID field reported; got %v", fields)
	}
	if id.Type != "string" {
		t.Errorf("ID.Type = %q, want %q", id.Type, "string")
	}
	if id.Offset != 0 {
		t.Errorf("ID.Offset = %d, want 0 (it is declared first)", id.Offset)
	}
	if id.Tag == "" {
		t.Error("ID.Tag is empty; the raw tag should include the json key")
	}
}

func TestInspectStruct_AcceptsAPointer(t *testing.T) {
	dto := sampleDTO()
	byValue := InspectStruct(dto)
	byPointer := InspectStruct(&dto)

	if len(byValue) != len(byPointer) {
		t.Errorf("value gave %d fields, pointer gave %d: a pointer must be "+
			"dereferenced, not rejected", len(byValue), len(byPointer))
	}
}

func TestInspectStruct_NonStruct(t *testing.T) {
	if got := InspectStruct(42); got != nil {
		t.Errorf("InspectStruct(42) = %v, want nil", got)
	}
	if got := InspectStruct(nil); got != nil {
		t.Errorf("InspectStruct(nil) = %v, want nil", got)
	}
}

func TestGetTag(t *testing.T) {
	dto := sampleDTO()

	tests := []struct {
		name      string
		field     string
		key       string
		want      string
		wantFound bool
	}{
		{"plain tag", "ID", "json", "id", true},
		{"tag with options", "Description", "json", "description,omitempty", true},
		{"missing key", "ID", "xml", "", false},
		{"missing field", "Nope", "json", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := GetTag(dto, tt.field, tt.key)
			if found != tt.wantFound {
				t.Fatalf("found = %v, want %v", found, tt.wantFound)
			}
			if got != tt.want {
				t.Errorf("GetTag = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToMap_UsesTagNames(t *testing.T) {
	m := ToMap(sampleDTO())
	if m == nil {
		t.Fatal("ToMap returned nil for a ProductDTO")
	}

	if got, ok := m["id"]; !ok || got != "p1" {
		t.Errorf(`m["id"] = %v (present=%v), want "p1"`, got, ok)
	}
	if _, ok := m["ID"]; ok {
		t.Error(`m has key "ID": the json tag name should win over the field name`)
	}
	// The tag is `json:"description,omitempty"` — the key is the name only.
	if _, ok := m["description"]; !ok {
		t.Error(`m has no key "description": strip the options after the comma`)
	}
	if _, ok := m["description,omitempty"]; ok {
		t.Error(`m has key "description,omitempty": the options are not part of the name`)
	}
}

func TestToMap_NonStruct(t *testing.T) {
	if got := ToMap(42); got != nil {
		t.Errorf("ToMap(42) = %v, want nil", got)
	}
}

func TestFromMap_RoundTrips(t *testing.T) {
	original := sampleDTO()
	m := ToMap(original)

	var restored domain.ProductDTO
	if err := FromMap(m, &restored); err != nil {
		t.Fatalf("FromMap returned %v", err)
	}

	if restored.ID != original.ID || restored.Name != original.Name {
		t.Errorf("round trip gave %+v, want %+v", restored, original)
	}
	if restored.Price != original.Price {
		t.Errorf("Price round-tripped to %v, want %v", restored.Price, original.Price)
	}
}

func TestFromMap_RequiresAPointer(t *testing.T) {
	m := map[string]any{"id": "p1"}

	if err := FromMap(m, domain.ProductDTO{}); err == nil {
		t.Error("FromMap into a value returned nil; there is nothing to assign to")
	}
	if err := FromMap(m, nil); err == nil {
		t.Error("FromMap(nil) returned nil")
	}
}

func TestFromMap_RejectsTheWrongType(t *testing.T) {
	var dto domain.ProductDTO
	err := FromMap(map[string]any{"id": 42}, &dto)
	if err == nil {
		t.Error("assigning an int to a string field returned nil. reflect panics " +
			"on this — check the type before you Set, and return an error instead")
	}
}

func TestFromMap_IgnoresUnknownKeys(t *testing.T) {
	var dto domain.ProductDTO
	err := FromMap(map[string]any{"id": "p1", "not_a_field": "x"}, &dto)
	if err != nil {
		t.Fatalf("FromMap returned %v; an unknown key is not an error here", err)
	}
	if dto.ID != "p1" {
		t.Errorf("ID = %q, want %q", dto.ID, "p1")
	}
}

func TestCallMethod(t *testing.T) {
	zero := domain.Zero(domain.AUD)

	out, err := CallMethod(zero, "IsZero")
	if err != nil {
		t.Fatalf("CallMethod(IsZero) returned %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("IsZero returned %d values, want 1", len(out))
	}
	if isZero, ok := out[0].(bool); !ok || !isZero {
		t.Errorf("IsZero = %v, want true", out[0])
	}

	out, err = CallMethod(zero, "Cents")
	if err != nil {
		t.Fatalf("CallMethod(Cents) returned %v", err)
	}
	if cents, ok := out[0].(int64); !ok || cents != 0 {
		t.Errorf("Cents = %v, want int64(0)", out[0])
	}
}

func TestCallMethod_UnknownMethodIsAnErrorNotAPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CallMethod panicked on an unknown method: %v. "+
				"Validate before you Call — that is the whole discipline of "+
				"reflection", r)
		}
	}()

	if _, err := CallMethod(domain.Zero(domain.AUD), "Nonexistent"); err == nil {
		t.Error("CallMethod on a missing method returned nil error")
	}
}

// ---- What it costs ----
//
//	go test ./internal/foundation/performance -bench=. -benchmem
//
// Direct field access against the reflection-based mapper, on the same struct.
// Look at the ratio, then decide where reflection belongs in a request path.

var sinkMap map[string]any
var reflectSinkString string

func BenchmarkDirectFieldAccess(b *testing.B) {
	dto := sampleDTO()
	b.ReportAllocs()
	for range b.N {
		reflectSinkString = dto.ID
	}
}

func BenchmarkToMapReflection(b *testing.B) {
	dto := sampleDTO()
	b.ReportAllocs()
	for range b.N {
		sinkMap = ToMap(dto)
	}
}
