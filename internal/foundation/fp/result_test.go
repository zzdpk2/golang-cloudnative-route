package fp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// ---- Result Tests ----

func TestFromTuple(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		r := FromTuple(strconv.Atoi("42"))
		if !r.IsOk() {
			t.Fatal("should be ok")
		}
		if r.Unwrap() != 42 {
			t.Errorf("got %d", r.Unwrap())
		}
	})
	t.Run("error", func(t *testing.T) {
		r := FromTuple(strconv.Atoi("abc"))
		if !r.IsErr() {
			t.Fatal("should be err")
		}
	})
}

func TestResult_Unwrap(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		if Ok(42).Unwrap() != 42 {
			t.Error("wrong value")
		}
	})
	t.Run("panics on err", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		Err[int](errors.New("boom")).Unwrap()
	})
}

func TestResult_UnwrapOr(t *testing.T) {
	if Ok(42).UnwrapOr(0) != 42 {
		t.Error("should return value")
	}
	if Err[int](errors.New("x")).UnwrapOr(99) != 99 {
		t.Error("should return default")
	}
}

func TestResult_UnwrapOrElse(t *testing.T) {
	r := Err[string](errors.New("missing"))
	got := r.UnwrapOrElse(func(err error) string {
		return "fallback: " + err.Error()
	})
	if got != "fallback: missing" {
		t.Errorf("got %q", got)
	}
}

func TestResult_ToTuple(t *testing.T) {
	v, err := Ok(42).ToTuple()
	if err != nil || v != 42 {
		t.Error("bad tuple")
	}

	_, err = Err[int](errors.New("bad")).ToTuple()
	if err == nil {
		t.Error("should have error")
	}
}

func TestResult_Map(t *testing.T) {
	got := Ok(5).Map(func(n int) int { return n * 2 })
	if got.Unwrap() != 10 {
		t.Error("map")
	}

	errR := Err[int](errors.New("x")).Map(func(n int) int { return n * 2 })
	if errR.IsOk() {
		t.Error("should stay err")
	}
}

func TestMapResult(t *testing.T) {
	r := Ok(42)
	got := MapResult(r, strconv.Itoa)
	if got.Unwrap() != "42" {
		t.Errorf("got %q", got.Unwrap())
	}

	errR := Err[int](errors.New("x"))
	got2 := MapResult(errR, strconv.Itoa)
	if got2.IsOk() {
		t.Error("should propagate error")
	}
}

func TestResult_FlatMap(t *testing.T) {
	safeDiv := func(n int) Result[int] {
		if n == 0 {
			return Err[int](errors.New("div by zero"))
		}
		return Ok(100 / n)
	}

	got := Ok(5).FlatMap(safeDiv)
	if got.Unwrap() != 20 {
		t.Errorf("got %d", got.Unwrap())
	}

	got2 := Ok(0).FlatMap(safeDiv)
	if got2.IsOk() {
		t.Error("should be err")
	}

	got3 := Err[int](errors.New("x")).FlatMap(safeDiv)
	if got3.IsOk() {
		t.Error("should stay err")
	}
}

func TestFlatMapResult(t *testing.T) {
	parseInt := func(s string) Result[int] {
		return FromTuple(strconv.Atoi(s))
	}

	got := FlatMapResult(Ok("42"), parseInt)
	if got.Unwrap() != 42 {
		t.Error("should parse")
	}

	got2 := FlatMapResult(Ok("abc"), parseInt)
	if got2.IsOk() {
		t.Error("should fail")
	}
}

func TestResult_MapErr(t *testing.T) {
	r := Err[int](errors.New("raw"))
	got := r.MapErr(func(err error) error {
		return fmt.Errorf("wrapped: %w", err)
	})
	if !strings.Contains(got.Error().Error(), "wrapped") {
		t.Errorf("got %v", got.Error())
	}
}

func TestResult_Recover(t *testing.T) {
	r := Err[int](errors.New("oops"))
	got := r.Recover(func(err error) Result[int] {
		return Ok(0) // See the corresponding tests for the intended behavior.
	})
	if got.Unwrap() != 0 {
		t.Error("should recover")
	}

	ok := Ok(42).Recover(func(err error) Result[int] { return Ok(0) })
	if ok.Unwrap() != 42 {
		t.Error("should keep ok value")
	}
}

func TestMatch(t *testing.T) {
	msg := Match(
		Ok(42),
		func(v int) string { return fmt.Sprintf("got %d", v) },
		func(err error) string { return "error: " + err.Error() },
	)
	if msg != "got 42" {
		t.Errorf("got %q", msg)
	}

	msg2 := Match(
		Err[int](errors.New("boom")),
		func(v int) string { return "ok" },
		func(err error) string { return "error: " + err.Error() },
	)
	if msg2 != "error: boom" {
		t.Errorf("got %q", msg2)
	}
}

func TestResult_ChainedOperations(t *testing.T) {
	parseAndValidate := func(input string) Result[int] {
		return FromTuple(strconv.Atoi(input)).
			FlatMap(func(n int) Result[int] {
				if n < 0 || n > 100 {
					return Err[int](fmt.Errorf("out of range: %d", n))
				}
				return Ok(n)
			}).
			Map(func(n int) int { return n * 2 })
	}

	if parseAndValidate("25").Unwrap() != 50 {
		t.Error("valid")
	}
	if parseAndValidate("abc").IsOk() {
		t.Error("parse fail")
	}
	if parseAndValidate("200").IsOk() {
		t.Error("range fail")
	}
}

// ---- Catch Tests ----

func TestCatch(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		r := Catch(func() int { return 42 })
		if r.Unwrap() != 42 {
			t.Error("should catch value")
		}
	})
	t.Run("catches panic error", func(t *testing.T) {
		r := Catch(func() int { panic(errors.New("boom")) })
		if r.IsOk() {
			t.Fatal("should be err")
		}
		if r.Error().Error() != "boom" {
			t.Errorf("got %v", r.Error())
		}
	})
	t.Run("catches panic string", func(t *testing.T) {
		r := Catch(func() int { panic("oops") })
		if r.IsOk() {
			t.Fatal("should be err")
		}
		if !strings.Contains(r.Error().Error(), "oops") {
			t.Errorf("got %v", r.Error())
		}
	})
}

// ---- Option Tests ----

func TestOption_FlatMap(t *testing.T) {
	safeSqrt := func(n float64) Option[float64] {
		if n < 0 {
			return None[float64]()
		}
		return Some(n * n) // See the corresponding tests for the intended behavior.
	}

	got := Some(4.0).FlatMap(safeSqrt)
	if !got.IsSome() {
		t.Error("should be some")
	}

	got2 := Some(-1.0).FlatMap(safeSqrt)
	if got2.IsSome() {
		t.Error("should be none")
	}

	got3 := None[float64]().FlatMap(safeSqrt)
	if got3.IsSome() {
		t.Error("none flatmap should be none")
	}
}

func TestOption_ToResult(t *testing.T) {
	got := Some(42).ToResult(errors.New("missing"))
	if got.Unwrap() != 42 {
		t.Error("some → ok")
	}

	got2 := None[int]().ToResult(errors.New("missing"))
	if got2.IsOk() {
		t.Error("none → err")
	}
	if got2.Error().Error() != "missing" {
		t.Error("wrong error")
	}
}

func TestFromPtr(t *testing.T) {
	x := 42
	got := FromPtr(&x)
	if !got.IsSome() || got.Unwrap() != 42 {
		t.Error("from non-nil ptr")
	}

	got2 := FromPtr[int](nil)
	if got2.IsSome() {
		t.Error("from nil ptr")
	}
}

// ---- Either Tests ----

func TestEither(t *testing.T) {
	t.Run("right", func(t *testing.T) {
		e := Right[string, int](42)
		v, ok := e.Right()
		if !ok || v != 42 {
			t.Error("should be right")
		}
		if e.IsLeft() {
			t.Error("not left")
		}
	})
	t.Run("left", func(t *testing.T) {
		e := Left[string, int]("error")
		v, ok := e.Left()
		if !ok || v != "error" {
			t.Error("should be left")
		}
		if e.IsRight() {
			t.Error("not right")
		}
	})
}

func TestFold(t *testing.T) {
	r := Right[string, int](42)
	got := Fold(r,
		func(s string) string { return "left: " + s },
		func(n int) string { return fmt.Sprintf("right: %d", n) },
	)
	if got != "right: 42" {
		t.Errorf("got %q", got)
	}
}
