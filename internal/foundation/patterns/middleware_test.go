package patterns

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestChain(t *testing.T) {
	var order []string

	mw1 := MiddlewareFunc[string, string](func(next Handler[string, string]) Handler[string, string] {
		return func(ctx context.Context, req string) (string, error) {
			order = append(order, "mw1-before")
			resp, err := next(ctx, req)
			order = append(order, "mw1-after")
			return resp, err
		}
	})
	mw2 := MiddlewareFunc[string, string](func(next Handler[string, string]) Handler[string, string] {
		return func(ctx context.Context, req string) (string, error) {
			order = append(order, "mw2-before")
			resp, err := next(ctx, req)
			order = append(order, "mw2-after")
			return resp, err
		}
	})

	handler := Handler[string, string](func(ctx context.Context, req string) (string, error) {
		order = append(order, "handler")
		return "result:" + req, nil
	})

	chained := Chain(handler, mw1, mw2)
	got, err := chained(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if got != "result:test" {
		t.Errorf("got %q", got)
	}

	want := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i, w := range want {
		if order[i] != w {
			t.Errorf("[%d] = %q, want %q", i, order[i], w)
		}
	}
}

func TestAuthMiddleware(t *testing.T) {
	validate := func(token string) (string, error) {
		if token == "valid-token" {
			return "user-123", nil
		}
		return "", fmt.Errorf("invalid token")
	}

	handler := Handler[string, string](func(ctx context.Context, req string) (string, error) {
		return "ok", nil
	})

	authed := AuthMiddleware[string, string](validate)(handler)

	t.Run("no token", func(t *testing.T) {
		_, err := authed(context.Background(), "req")
		if err == nil {
			t.Error("should fail without token")
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		ctx := WithAuthToken(context.Background(), "bad-token")
		_, err := authed(ctx, "req")
		if err == nil {
			t.Error("should fail with invalid token")
		}
	})

	t.Run("valid token", func(t *testing.T) {
		ctx := WithAuthToken(context.Background(), "valid-token")
		got, err := authed(ctx, "req")
		if err != nil {
			t.Fatal(err)
		}
		if got != "ok" {
			t.Errorf("got %q", got)
		}
	})
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(100, 5) // 5 burst
	defer rl.Close()

	allowed := 0
	for i := 0; i < 10; i++ {
		if rl.Allow() {
			allowed++
		}
	}
	if allowed != 5 {
		t.Errorf("allowed = %d, want 5 (burst)", allowed)
	}

	time.Sleep(100 * time.Millisecond)
	if !rl.Allow() {
		t.Error("should have replenished token")
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	rl := NewRateLimiter(1000, 1) // 1 burst
	defer rl.Close()

	err := rl.Wait(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = rl.Wait(ctx)
	t.Logf("Wait result: %v", err)
}

func TestRateLimiter_ContextCancel(t *testing.T) {
	rl := NewRateLimiter(1, 1) // See the corresponding tests for the intended behavior.
	defer rl.Close()

	rl.Allow() // See the corresponding tests for the intended behavior.

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // See the corresponding tests for the intended behavior.

	err := rl.Wait(ctx)
	if err == nil {
		t.Error("should fail on cancelled context")
	}
}

func TestMetricsMiddleware(t *testing.T) {
	metrics := &Metrics{}

	handler := Handler[int, int](func(ctx context.Context, n int) (int, error) {
		if n < 0 {
			return 0, fmt.Errorf("negative")
		}
		return n * 2, nil
	})

	metered := MetricsMiddleware[int, int](metrics)(handler)

	for i := 0; i < 5; i++ {
		metered(context.Background(), i)
	}
	metered(context.Background(), -1)

	if metrics.RequestCount() != 6 {
		t.Errorf("requests = %d", metrics.RequestCount())
	}
	if metrics.ErrorCount() != 1 {
		t.Errorf("errors = %d", metrics.ErrorCount())
	}
	if metrics.AvgDuration() == 0 {
		t.Error("avg duration should be > 0")
	}
}

func TestFullMiddlewareChain(t *testing.T) {
	metrics := &Metrics{}
	rl := NewRateLimiter(1000, 100)
	defer rl.Close()

	validate := func(token string) (string, error) {
		if token == "ok" {
			return "user", nil
		}
		return "", fmt.Errorf("bad")
	}

	handler := Handler[string, string](func(ctx context.Context, req string) (string, error) {
		return "processed:" + req, nil
	})

	full := Chain(handler,
		MetricsMiddleware[string, string](metrics),
		AuthMiddleware[string, string](validate),
		RateLimitMiddleware[string, string](rl),
	)

	ctx := WithAuthToken(context.Background(), "ok")
	got, err := full(ctx, "data")
	if err != nil {
		t.Fatal(err)
	}
	if got != "processed:data" {
		t.Errorf("got %q", got)
	}
	if metrics.RequestCount() != 1 {
		t.Errorf("requests = %d", metrics.RequestCount())
	}
}
