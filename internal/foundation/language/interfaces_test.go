package language

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

var _ ReadWriter = (*bytes.Buffer)(nil)

type terminalDataReader struct {
	data   []byte
	stalls int
}

func (r *terminalDataReader) Read(buffer []byte) (int, error) {
	if r.stalls > 0 {
		r.stalls--
		return 0, nil
	}
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := copy(buffer, r.data)
	r.data = r.data[n:]
	if len(r.data) == 0 {
		return n, io.EOF
	}
	return n, nil
}

type shortWriter struct{}

func (shortWriter) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	return len(data) - 1, nil
}

type partialErrorWriter struct {
	err error
}

func (w partialErrorWriter) Write(data []byte) (int, error) {
	if len(data) < 2 {
		return 0, w.err
	}
	return 2, w.err
}

func TestInterfaceFoundations(t *testing.T) {
	t.Run("method sets and embedding", func(t *testing.T) {
		pointer := &PointerSpeaker{Name: "beta"}
		got := SpeakAll([]Speaker{
			ValueSpeaker{Name: "alpha"},
			pointer,
			NamedSpeaker{ValueSpeaker: ValueSpeaker{Name: "gamma"}, Role: "worker"},
		})
		if want := []string{"value:alpha", "pointer:beta", "value:gamma"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("SpeakAll() = %v, want %v", got, want)
		}
	})

	t.Run("type switches preserve the typed-nil distinction", func(t *testing.T) {
		var pointer *PointerSpeaker
		var speaker Speaker = pointer
		if speaker == nil {
			t.Fatal("an interface containing a typed nil pointer is not nil")
		}
		cases := []struct {
			name  string
			value any
			want  DynamicReport
		}{
			{name: "nil", want: DynamicReport{Kind: DynamicNil, NilUnderlying: true}},
			{name: "integer", value: 7, want: DynamicReport{Kind: DynamicInteger}},
			{name: "text", value: "router", want: DynamicReport{Kind: DynamicText}},
			{name: "speaker", value: ValueSpeaker{Name: "ready"}, want: DynamicReport{Kind: DynamicSpeaker}},
			{name: "typed nil speaker", value: speaker, want: DynamicReport{Kind: DynamicSpeaker, NilUnderlying: true}},
			{name: "other", value: 3.14, want: DynamicReport{Kind: DynamicOther}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				if got := InspectDynamic(tc.value); got != tc.want {
					t.Fatalf("InspectDynamic(%#v) = %#v, want %#v", tc.value, got, tc.want)
				}
			})
		}
	})

	t.Run("interface composition supports streaming and terminal errors", func(t *testing.T) {
		source := strings.NewReader("router-ready")
		var destination bytes.Buffer
		written, err := Transfer(&destination, source, make([]byte, 3))
		if !errors.Is(err, io.EOF) {
			t.Fatalf("terminal error = %v, want io.EOF", err)
		}
		if written != int64(len("router-ready")) || destination.String() != "router-ready" {
			t.Fatalf("written=%d destination=%q", written, destination.String())
		}
		if _, err := Transfer(&destination, source, nil); err == nil {
			t.Fatal("empty scratch buffer must be rejected")
		}

		destination.Reset()
		terminal := &terminalDataReader{data: []byte("tail"), stalls: 2}
		written, err = Transfer(&destination, terminal, make([]byte, 8))
		if !errors.Is(err, io.EOF) || written != 4 || destination.String() != "tail" {
			t.Fatalf("data with EOF = (%d, %v, %q)", written, err, destination.String())
		}

		written, err = Transfer(shortWriter{}, strings.NewReader("abc"), make([]byte, 3))
		if !errors.Is(err, io.ErrShortWrite) || written != 2 {
			t.Fatalf("short write = (%d, %v), want (2, io.ErrShortWrite)", written, err)
		}

		writeFailure := errors.New("write failed")
		written, err = Transfer(partialErrorWriter{err: writeFailure}, strings.NewReader("abc"), make([]byte, 3))
		if !errors.Is(err, writeFailure) || written != 2 {
			t.Fatalf("partial write error = (%d, %v)", written, err)
		}
	})

	t.Run("function types adapt to interfaces", func(t *testing.T) {
		handler := HandlerFunc(strings.ToUpper)
		if got := ApplyHandler(handler, "route"); got != "ROUTE" {
			t.Fatalf("ApplyHandler() = %q, want ROUTE", got)
		}
	})
}
