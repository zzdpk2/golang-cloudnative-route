package http

import (
	"bytes"
	"encoding/json"
	"log"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSONAndDecodeJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSON(rr, stdhttp.StatusCreated, map[string]string{"ok": "true"})

	if rr.Code != stdhttp.StatusCreated {
		t.Fatalf("status = %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}

	var decoded map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["ok"] != "true" {
		t.Fatalf("decoded = %v", decoded)
	}

	req := httptest.NewRequest(stdhttp.MethodPost, "/", bytes.NewBufferString(`{"name":"rex"}`))
	var dst map[string]string
	if err := decodeJSON(req, &dst); err != nil {
		t.Fatalf("decodeJSON: %v", err)
	}
	if dst["name"] != "rex" {
		t.Fatalf("dst = %v", dst)
	}
}

func TestStatusRecorder(t *testing.T) {
	rr := httptest.NewRecorder()
	rec := &statusRecorder{ResponseWriter: rr}

	n, err := rec.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 || rec.bytes != 5 {
		t.Fatalf("bytes = %d n=%d", rec.bytes, n)
	}
	if rec.status != stdhttp.StatusOK {
		t.Fatalf("status = %d", rec.status)
	}
}

func TestChainOrder(t *testing.T) {
	var calls []string

	base := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		calls = append(calls, "handler")
	})

	a := func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			calls = append(calls, "a-before")
			next.ServeHTTP(w, r)
			calls = append(calls, "a-after")
		})
	}
	b := func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			calls = append(calls, "b-before")
			next.ServeHTTP(w, r)
			calls = append(calls, "b-after")
		})
	}

	Chain(base, a, b).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(stdhttp.MethodGet, "/", nil))

	want := []string{"a-before", "b-before", "handler", "b-after", "a-after"}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("calls = %v, want %v", calls, want)
		}
	}
}

func TestMiddleware(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	handler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusAccepted)
		_, _ = w.Write([]byte("ok"))
	})

	wrapped := Chain(handler, LoggingMiddleware(logger), CORSMiddleware("*"))
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, httptest.NewRequest(stdhttp.MethodGet, "/", nil))

	if rr.Code != stdhttp.StatusAccepted {
		t.Fatalf("status = %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("CORS header missing")
	}
	if logBuf.Len() == 0 {
		t.Fatal("logging middleware should log")
	}
}

func TestAuthMiddleware(t *testing.T) {
	handler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	})
	wrapped := AuthMiddleware(func(token string) bool { return token == "secret" })(handler)

	unauthorized := httptest.NewRecorder()
	wrapped.ServeHTTP(unauthorized, httptest.NewRequest(stdhttp.MethodGet, "/", nil))
	if unauthorized.Code != stdhttp.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", unauthorized.Code)
	}

	authorized := httptest.NewRecorder()
	req := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer secret")
	wrapped.ServeHTTP(authorized, req)
	if authorized.Code != stdhttp.StatusNoContent {
		t.Fatalf("authorized status = %d", authorized.Code)
	}
}
