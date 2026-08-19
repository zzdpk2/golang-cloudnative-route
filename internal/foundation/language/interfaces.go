package language

import (
	"errors"
	"io"
)

// Speaker is deliberately tiny. Router-facing interfaces should be owned by
// callers and expose only the behavior a decision needs.
type Speaker interface {
	Speak() string
}

// ValueSpeaker demonstrates a value-receiver method set.
type ValueSpeaker struct {
	Name string
}

func (s ValueSpeaker) Speak() string {
	return "value:" + s.Name
}

// PointerSpeaker demonstrates a pointer-receiver method set. A PointerSpeaker
// value does not implement Speaker; *PointerSpeaker does.
type PointerSpeaker struct {
	Name string
}

func (s *PointerSpeaker) Speak() string {
	return "pointer:" + s.Name
}

// NamedSpeaker demonstrates embedding and promoted methods without adding
// another forwarding implementation.
type NamedSpeaker struct {
	ValueSpeaker
	Role string
}

// SpeakAll calls each dependency polymorphically and preserves input order.
func SpeakAll(speakers []Speaker) []string {
	res := make([]string, 0, len(speakers))

	for _, speaker := range speakers {
		res = append(res, speaker.Speak())
	}

	return res
}

// DynamicKind is a stable classification returned by InspectDynamic.
type DynamicKind string

const (
	DynamicNil     DynamicKind = "nil"
	DynamicInteger DynamicKind = "integer"
	DynamicText    DynamicKind = "text"
	DynamicSpeaker DynamicKind = "speaker"
	DynamicOther   DynamicKind = "other"
)

// DynamicReport keeps both type-switch behavior and the typed-nil trap visible.
// NilUnderlying is true for nil and for supported typed nil pointers.
type DynamicReport struct {
	Kind          DynamicKind
	NilUnderlying bool
}

// InspectDynamic classifies nil, int, string, Speaker, and other values.
// A Speaker holding (*PointerSpeaker)(nil) is a non-nil interface whose
// underlying pointer is nil; report that without calling Speak.
func InspectDynamic(value any) DynamicReport {
	// panic("TODO")
	switch v := value.(type) {
	case nil:
		return DynamicReport{Kind: DynamicNil, NilUnderlying: true}

	case int:
		return DynamicReport{Kind: DynamicInteger}

	case string:
		return DynamicReport{Kind: DynamicText}

	case Speaker:
		report := DynamicReport{Kind: DynamicSpeaker}
		if p, ok := v.(*PointerSpeaker); ok && p == nil {
			report.NilUnderlying = true
		}
		return report
	default:
		return DynamicReport{Kind: DynamicOther}
	}

}

// Reader and Writer demonstrate interface composition without recreating a
// complete bytes.Buffer exercise.
type Reader interface {
	Read([]byte) (int, error)
}

type Writer interface {
	Write([]byte) (int, error)
}

type ReadWriter interface {
	Reader
	Writer
}

// Transfer copies through narrow interfaces using the provided scratch space.
// Process bytes returned with a terminal read error before returning that error.
// Preserve partial progress on write errors, convert a nil-error short write to
// io.ErrShortWrite, tolerate transient (0, nil) reads, and reject empty scratch.
func Transfer(dst Writer, src Reader, scratch []byte) (int64, error) {
	if len(scratch) == 0 {
		return 0, errors.New("Buffer error!")
	}

	total := 0

	for {
		n, readErr := src.Read(scratch)
		if n > 0 {
			written, writeErr := dst.Write(scratch[:n])
			total += written

			if writeErr != nil {
				return int64(total), writeErr
			}

			if written < n {
				return int64(total), io.ErrShortWrite
			}
		}

		if readErr == io.EOF {
			return int64(total), io.EOF
		}

		if readErr != nil && readErr != io.EOF {
			return int64(total), readErr
		}
	}
}

// Handler and HandlerFunc use the same adapter pattern as net/http.HandlerFunc.
type Handler interface {
	Handle(string) string
}

type HandlerFunc func(string) string

func (f HandlerFunc) Handle(request string) string {
	return f(request)
}

func ApplyHandler(handler Handler, request string) string {

	return handler.Handle(request)
}

var _ Speaker = ValueSpeaker{}
var _ Speaker = (*PointerSpeaker)(nil)
var _ Speaker = NamedSpeaker{}
var _ Handler = HandlerFunc(nil)
