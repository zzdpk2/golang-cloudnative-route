package patterns

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNormalizeAndMatchRoute(t *testing.T) {
	if got := normalizePath("users/"); got != "/users" {
		t.Fatalf("normalizePath = %q", got)
	}

	params, ok := matchRoute("/users/:id/books/:bookID", "/users/42/books/99")
	if !ok {
		t.Fatal("expected route to match")
	}
	if params["id"] != "42" || params["bookID"] != "99" {
		t.Fatalf("params = %#v", params)
	}

	if _, ok := matchRoute("/users/:id", "/users/42/books"); ok {
		t.Fatal("route with different segment count should not match")
	}
}

func TestRouteParamsQueryAndJSON(t *testing.T) {
	e := New()
	e.GET("/users/:id", func(c *Context) {
		c.JSON(http.StatusOK, map[string]string{
			"id":  c.Param("id"),
			"tab": c.Query("tab"),
		})
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/42?tab=posts", nil)
	e.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}

	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["id"] != "42" || body["tab"] != "posts" {
		t.Fatalf("body = %#v", body)
	}
}

func TestGroupMiddlewareOrderAndAbort(t *testing.T) {
	e := New()
	var calls []string

	e.Use(func(c *Context) {
		calls = append(calls, "global-before")
		c.Next()
		calls = append(calls, "global-after")
	})

	api := e.Group("/api", func(c *Context) {
		calls = append(calls, "group-before")
		c.Next()
		calls = append(calls, "group-after")
	})

	api.GET("/ping", func(c *Context) {
		calls = append(calls, "handler")
		c.String(http.StatusOK, "pong")
	})

	rr := httptest.NewRecorder()
	e.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/ping", nil))

	want := []string{"global-before", "group-before", "handler", "group-after", "global-after"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}

	blocked := New()
	ranHandler := false
	blocked.Use(func(c *Context) {
		c.Abort(http.StatusUnauthorized)
	})
	blocked.GET("/secret", func(c *Context) {
		ranHandler = true
	})

	unauthorized := httptest.NewRecorder()
	blocked.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/secret", nil))

	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", unauthorized.Code)
	}
	if ranHandler {
		t.Fatal("handler ran after Abort")
	}
}

func TestBindJSONAndContextKeys(t *testing.T) {
	e := New()
	e.POST("/echo", func(c *Context) {
		var in struct {
			Name string `json:"name"`
		}
		if err := c.BindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		c.Set("name", in.Name)
		v, ok := c.Get("name")
		if !ok {
			c.String(http.StatusInternalServerError, "missing key")
			return
		}

		c.JSON(http.StatusCreated, map[string]string{"name": v.(string)})
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/echo", stringsReader(`{"name":"rex"}`))
	e.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestRecoveryRequestIDLoggerAndRecorder(t *testing.T) {
	e := New()
	sink := NewAsyncLogSink(8)

	var tick int64
	base := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	now := func() time.Time {
		n := atomic.AddInt64(&tick, 1)
		return base.Add(time.Duration(n) * time.Millisecond)
	}

	e.Use(
		RecoveryMiddleware(),
		RequestIDMiddleware(e),
		LoggerMiddleware(sink, now),
	)
	e.GET("/panic", func(c *Context) {
		panic("boom")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set("X-Request-ID", "client-123")
	e.ServeHTTP(rr, req)
	sink.Close()

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rr.Code)
	}
	if got := rr.Header().Get("X-Request-ID"); got != "client-123" {
		t.Fatalf("X-Request-ID = %q", got)
	}

	events := sink.Events()
	if len(events) != 1 {
		t.Fatalf("events len = %d", len(events))
	}
	if events[0].RequestID != "client-123" || events[0].Status != http.StatusInternalServerError {
		t.Fatalf("event = %#v", events[0])
	}

	rec := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	if _, err := rec.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if rec.Status() != http.StatusOK || rec.bytes != 5 {
		t.Fatalf("rec status=%d bytes=%d", rec.Status(), rec.bytes)
	}
}

func TestStatsMiddlewareIsConcurrencySafe(t *testing.T) {
	e := New()
	stats := NewStats()
	e.Use(StatsMiddleware(stats))
	e.GET("/items/:id", func(c *Context) {
		c.String(http.StatusOK, c.Param("id"))
	})

	const n = 200
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/items/"+itoa(i), nil)
			e.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Errorf("status = %d", rr.Code)
			}
		}()
	}

	wg.Wait()
	snap := stats.Snapshot()
	if snap.Total != n {
		t.Fatalf("total = %d, want %d", snap.Total, n)
	}
	if snap.ByPath["/items/:id"] != n {
		t.Fatalf("byPath = %#v", snap.ByPath)
	}
}

func TestTimeoutMiddlewareAddsDeadline(t *testing.T) {
	e := New()
	e.Use(TimeoutMiddleware(50 * time.Millisecond))
	e.GET("/deadline", func(c *Context) {
		if _, ok := c.RequestContext().Deadline(); !ok {
			c.String(http.StatusInternalServerError, "missing deadline")
			return
		}
		c.String(http.StatusOK, "ok")
	})

	rr := httptest.NewRecorder()
	e.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/deadline", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
}

func stringsReader(s string) *strings.Reader {
	return strings.NewReader(s)
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
