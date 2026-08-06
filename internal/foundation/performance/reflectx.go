// Package reflectx is the reflection laboratory.
//
// It used to live inside a package called `unsafe`, which was misleading —
// reflection is not unsafe, it is just slow and untyped. It is also the machine
// underneath things you already use every day: encoding/json reads your struct
// tags with exactly these calls, and so does every ORM, config loader, and
// validation library in Go.
//
// The subject here is the real DTOs — entity.ProductDTO and
// application.AddressDTO — because that is where reflection actually earns its
// place in this codebase: turning a domain object into a wire shape and back.
//
// # The rule to hold on to
//
// Reflection trades compile-time checking for run-time flexibility, and the
// trade is worse than it looks. A field renamed in a struct breaks a
// reflection-based mapper at run time, in production, on the one code path
// nobody tested. The compiler cannot help you, because from its point of view
// nothing is wrong.
//
// So: reflection is right for a *library* that cannot know its caller's types,
// and wrong for application code that can. If you find yourself reaching for it
// inside the order service, write the mapping by hand instead — it is longer,
// and the compiler checks it.
//
// # Cost
//
// Reflection is roughly one to two orders of magnitude slower than direct field
// access. The benchmark in this package measures it on ProductDTO so you have a
// real number rather than a rumour.
package performance

import (
	"fmt"
	"reflect"
)

// FieldInfo describes one struct field.
type FieldInfo struct {
	Name   string
	Type   string
	Tag    string
	Offset uintptr
	Size   uintptr
}

// InspectStruct reports every exported field of a struct.
//
//	InspectStruct(entity.ProductDTO{}) → one FieldInfo per field, in
//	                                     declaration order
//	InspectStruct(&entity.ProductDTO{}) → the same; a pointer is dereferenced
//	InspectStruct(42)                   → nil; not a struct
//	InspectStruct(nil)                  → nil
//
// Name is the Go field name ("ID"), Type is its type as a string ("string"),
// Tag is the whole raw tag (`json:"id"`), Offset is the byte position inside
// the struct, and Size is the field's size.
//
// Accept both a value and a pointer — callers pass either, and requiring one is
// the kind of API friction that makes people stop using a helper. reflect.Ptr
// and Elem are how you unwrap.
//
// The offsets are the same numbers lab/layout computes; seeing them arrive by a
// completely different route is worth a moment.
func InspectStruct(v any) []FieldInfo {
	panic("TODO")
}

// GetTag reads one key out of one field's struct tag.
//
//	GetTag(entity.ProductDTO{}, "ID", "json")          → "id", true
//	GetTag(entity.ProductDTO{}, "Description", "json") → "description,omitempty", true
//	GetTag(entity.ProductDTO{}, "ID", "xml")           → "", false
//	GetTag(entity.ProductDTO{}, "Nope", "json")        → "", false
//
// Note the second example: the value is the *raw* tag content, comma and all.
// Splitting "description,omitempty" into a name and its options is the caller's
// job, and forgetting that is why hand-rolled mappers mis-handle omitempty.
//
// Use Tag.Lookup rather than Tag.Get. Get returns "" both for an absent key and
// for a key that is genuinely empty, and the comma-ok form is what distinguishes
// them — the same lesson as Product.GetMetadata in L2.
func GetTag(v any, fieldName, tagKey string) (string, bool) {
	panic("TODO")
}

// ToMap flattens a struct into a map keyed by json tag name, falling back to the
// field name when there is no tag.
//
//	ToMap(entity.ProductDTO{ID: "p1", Name: "Mug", Price: 9.99})
//	  → map[string]any{"id": "p1", "name": "Mug", "price": 9.99, ...}
//
//	ToMap(42)  → nil
//	ToMap(nil) → nil
//
// Use only the tag *name*, so "description,omitempty" becomes the key
// "description". Skip unexported fields — reflect will not let you read them
// anyway, and trying panics.
//
// This is roughly what encoding/json does on the way out. Writing it is the
// fastest way to stop treating json.Marshal as magic.
func ToMap(v any) map[string]any {
	panic("TODO")
}

// FromMap fills a struct from a map, matching on the same keys ToMap produces.
//
//	var dto entity.ProductDTO
//	FromMap(map[string]any{"id": "p1", "name": "Mug"}, &dto)
//	  → dto.ID == "p1", dto.Name == "Mug", other fields untouched
//
//	FromMap(m, entity.ProductDTO{})  → error: not a pointer, nothing to write to
//	FromMap(m, nil)                  → error
//	FromMap(map[string]any{"id": 42}, &dto) → error: id is a string field
//
// The direction that is genuinely harder, for three reasons the errors above
// hint at. You need a *pointer* or there is nothing to assign to. You need
// CanSet, which is false for unexported fields. And the map holds `any`, so the
// type has to be checked before assigning or reflect panics rather than
// returning an error.
//
// A key with no matching field is not an error — ignore it. Decide for yourself
// whether that is the right call, and note that encoding/json makes the same
// one by default and offers DisallowUnknownFields for when it is not.
func FromMap(m map[string]any, ptr any) error {
	panic("TODO")
}

// CallMethod invokes a method by name.
//
//	CallMethod(vo.Zero(vo.AUD), "IsZero")     → []any{true}, nil
//	CallMethod(vo.Zero(vo.AUD), "Cents")      → []any{int64(0)}, nil
//	CallMethod(vo.Zero(vo.AUD), "Nonexistent") → nil, error
//
// This is how a test framework finds your TestXxx functions and how an RPC
// server dispatches a method name off the wire.
//
// It is also where reflection is at its most dangerous: a wrong argument count
// or type panics rather than returning an error, so validate before you Call.
// The test passes a bad method name and expects an error, not a crash.
func CallMethod(obj any, methodName string, args ...any) ([]any, error) {
	panic("TODO")
}

var (
	_ = reflect.TypeOf
	_ = fmt.Sprintf
)
