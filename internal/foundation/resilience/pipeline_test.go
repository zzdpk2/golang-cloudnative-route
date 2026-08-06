package resilience

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestThen(t *testing.T) {
	parseInt := Step[string, int](func(ctx context.Context, s string) (int, error) {
		n := 0
		for _, c := range s {
			if c < '0' || c > '9' {
				return 0, fmt.Errorf("not a number")
			}
			n = n*10 + int(c-'0')
		}
		return n, nil
	})
	double := Step[int, int](func(ctx context.Context, n int) (int, error) {
		return n * 2, nil
	})

	pipeline := Then(parseInt, double)
	got, err := pipeline(context.Background(), "21")
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}

	_, err = pipeline(context.Background(), "abc")
	if err == nil {
		t.Error("should fail on bad input")
	}
}

func TestThen_ThreeSteps(t *testing.T) {
	step1 := Step[string, string](func(ctx context.Context, s string) (string, error) {
		return strings.TrimSpace(s), nil
	})
	step2 := Step[string, string](func(ctx context.Context, s string) (string, error) {
		if len(s) < 3 {
			return "", fmt.Errorf("too short")
		}
		return strings.ToUpper(s), nil
	})
	step3 := Step[string, string](func(ctx context.Context, s string) (string, error) {
		return "[" + s + "]", nil
	})

	pipeline := Then(Then(step1, step2), step3)
	got, err := pipeline(context.Background(), "  hello  ")
	if err != nil {
		t.Fatal(err)
	}
	if got != "[HELLO]" {
		t.Errorf("got %q", got)
	}
}

func TestWithLogging(t *testing.T) {
	var logs []string
	logFn := func(msg string) { logs = append(logs, msg) }

	step := Step[int, int](func(ctx context.Context, n int) (int, error) {
		return n * 2, nil
	})

	logged := WithLogging[int, int](logFn)(step)
	got, err := logged(context.Background(), 21)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
	if len(logs) < 2 {
		t.Errorf("expected at least 2 log entries, got %d", len(logs))
	}
}

func TestWithTiming(t *testing.T) {
	var recorded time.Duration
	step := Step[int, int](func(ctx context.Context, n int) (int, error) {
		time.Sleep(10 * time.Millisecond)
		return n, nil
	})

	timed := WithTiming[int, int](func(d time.Duration) { recorded = d })(step)
	timed(context.Background(), 1)

	if recorded < 10*time.Millisecond {
		t.Errorf("timing too low: %v", recorded)
	}
}

func TestWithRetry(t *testing.T) {
	callCount := 0
	failTwice := Step[int, int](func(ctx context.Context, n int) (int, error) {
		callCount++
		if callCount <= 2 {
			return 0, fmt.Errorf("fail #%d", callCount)
		}
		return n * 10, nil
	})

	retried := WithRetry[int, int](3, 1*time.Millisecond)(failTwice)
	got, err := retried(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if got != 50 {
		t.Errorf("got %d", got)
	}
	if callCount != 3 {
		t.Errorf("callCount = %d", callCount)
	}
}

func TestWithRetry_AllFail(t *testing.T) {
	alwaysFail := Step[int, int](func(ctx context.Context, n int) (int, error) {
		return 0, fmt.Errorf("always fails")
	})

	retried := WithRetry[int, int](2, 1*time.Millisecond)(alwaysFail)
	_, err := retried(context.Background(), 1)
	if err == nil {
		t.Error("should fail after all retries")
	}
	if !strings.Contains(err.Error(), "2 retries") {
		t.Errorf("error = %q", err)
	}
}

func TestWithRecover(t *testing.T) {
	panicky := Step[int, int](func(ctx context.Context, n int) (int, error) {
		panic("boom")
	})

	safe := WithRecover[int, int]()(panicky)
	_, err := safe(context.Background(), 1)
	if err == nil {
		t.Fatal("should catch panic")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("error = %q", err)
	}
}

func TestApply_MiddlewareChain(t *testing.T) {
	var logs []string
	var elapsed time.Duration

	step := Step[string, string](func(ctx context.Context, s string) (string, error) {
		return strings.ToUpper(s), nil
	})

	pipeline := Apply(step,
		WithLogging[string, string](func(msg string) { logs = append(logs, msg) }),
		WithTiming[string, string](func(d time.Duration) { elapsed = d }),
		WithRecover[string, string](),
	)

	got, err := pipeline(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if got != "HELLO" {
		t.Errorf("got %q", got)
	}
	if len(logs) == 0 {
		t.Error("should have logs")
	}
	if elapsed == 0 {
		t.Error("should have timing")
	}
}

func TestParallel(t *testing.T) {
	step1 := Step[int, string](func(ctx context.Context, n int) (string, error) {
		return fmt.Sprintf("a:%d", n), nil
	})
	step2 := Step[int, string](func(ctx context.Context, n int) (string, error) {
		return fmt.Sprintf("b:%d", n*2), nil
	})

	parallel := Parallel(step1, step2)
	results, err := parallel(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("len = %d", len(results))
	}
	if results[0] != "a:5" {
		t.Errorf("[0] = %q", results[0])
	}
	if results[1] != "b:10" {
		t.Errorf("[1] = %q", results[1])
	}
}

func TestFallback(t *testing.T) {
	primary := Step[string, string](func(ctx context.Context, s string) (string, error) {
		return "", fmt.Errorf("primary down")
	})
	backup := Step[string, string](func(ctx context.Context, s string) (string, error) {
		return "from backup: " + s, nil
	})

	fb := Fallback(primary, backup)
	got, err := fb(context.Background(), "data")
	if err != nil {
		t.Fatal(err)
	}
	if got != "from backup: data" {
		t.Errorf("got %q", got)
	}
}

func TestFallback_AllFail(t *testing.T) {
	fail1 := Step[int, int](func(ctx context.Context, n int) (int, error) {
		return 0, fmt.Errorf("fail1")
	})
	fail2 := Step[int, int](func(ctx context.Context, n int) (int, error) {
		return 0, fmt.Errorf("fail2")
	})

	fb := Fallback(fail1, fail2)
	_, err := fb(context.Background(), 1)
	if err == nil {
		t.Error("should fail")
	}
}

func TestRealWorldPipeline(t *testing.T) {

	clean := Step[string, string](func(ctx context.Context, s string) (string, error) {
		return strings.TrimSpace(strings.ToLower(s)), nil
	})
	validate := Step[string, string](func(ctx context.Context, s string) (string, error) {
		if len(s) == 0 {
			return "", fmt.Errorf("empty input")
		}
		if len(s) > 50 {
			return "", fmt.Errorf("input too long")
		}
		return s, nil
	})
	process := Step[string, string](func(ctx context.Context, s string) (string, error) {
		return "processed:" + s, nil
	})

	pipeline := Then(Then(clean, validate), process)

	pipeline = Apply(pipeline,
		WithRecover[string, string](),
	)

	got, err := pipeline(context.Background(), "  HELLO  ")
	if err != nil {
		t.Fatal(err)
	}
	if got != "processed:hello" {
		t.Errorf("got %q", got)
	}

	_, err = pipeline(context.Background(), "   ")
	if err == nil {
		t.Error("empty should fail")
	}
}
