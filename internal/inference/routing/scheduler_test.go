package routing

import (
	"context"
	"errors"
	"math"
	"testing"
)

func TestDefaultSchedulerBalancesPrefixAffinityAndCacheHeadroom(t *testing.T) {
	t.Parallel()

	endpoints := []Endpoint{
		{
			ID:                 "cached",
			URL:                "http://cached.example",
			Models:             []string{"tiny-llm"},
			Ready:              true,
			KVCacheUtilization: 0.80,
			CachedPrefixes:     []string{"shared"},
		},
		{
			ID:                 "headroom",
			URL:                "http://headroom.example",
			Models:             []string{"tiny-llm"},
			Ready:              true,
			KVCacheUtilization: 0.10,
		},
	}
	scheduler := NewDefaultScheduler()

	withPrefix, err := scheduler.Schedule(
		context.Background(),
		Request{Model: "tiny-llm", PrefixKey: "shared"},
		endpoints,
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := withPrefix.Selected.Endpoint.ID; got != "cached" {
		t.Fatalf("prefix-aware selection = %q, want cached", got)
	}

	withoutPrefix, err := scheduler.Schedule(
		context.Background(),
		Request{Model: "tiny-llm"},
		endpoints,
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := withoutPrefix.Selected.Endpoint.ID; got != "headroom" {
		t.Fatalf("headroom selection = %q, want headroom", got)
	}
}

func TestSchedulerFiltersUnreadyAndWrongModelEndpoints(t *testing.T) {
	t.Parallel()

	scheduler := NewDefaultScheduler()
	_, err := scheduler.Schedule(context.Background(), Request{Model: "wanted"}, []Endpoint{
		{ID: "unready", URL: "http://unready.example", Models: []string{"wanted"}, Ready: false},
		{ID: "wrong-model", URL: "http://wrong.example", Models: []string{"other"}, Ready: true},
	})
	if !errors.Is(err, ErrNoEndpoint) {
		t.Fatalf("error = %v, want ErrNoEndpoint", err)
	}
}

func TestMaxScorePickerUsesEndpointIDForDeterministicTies(t *testing.T) {
	t.Parallel()

	picker := MaxScorePicker{}
	selected, err := picker.Pick(context.Background(), Request{}, []ScoredEndpoint{
		{Endpoint: Endpoint{ID: "z"}, Score: 0.5},
		{Endpoint: Endpoint{ID: "a"}, Score: 0.5},
	})
	if err != nil {
		t.Fatal(err)
	}
	if selected.Endpoint.ID != "a" {
		t.Fatalf("selected %q, want a", selected.Endpoint.ID)
	}
}

func TestSchedulerRejectsOutOfRangePluginScore(t *testing.T) {
	t.Parallel()

	scheduler, err := NewScheduler(
		nil,
		[]WeightedScorer{{Scorer: fixedScorer{name: "broken", score: 1.1}, Weight: 1}},
		MaxScorePicker{},
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = scheduler.Schedule(context.Background(), Request{Model: "tiny-llm"}, []Endpoint{{
		ID: "one", URL: "http://one.example", Models: []string{"tiny-llm"}, Ready: true,
	}})
	if err == nil {
		t.Fatal("expected out-of-range score error")
	}
}

func TestSchedulerRejectsNonFiniteValues(t *testing.T) {
	t.Parallel()

	_, err := NewScheduler(
		nil,
		[]WeightedScorer{{Scorer: fixedScorer{name: "one", score: 1}, Weight: math.Inf(1)}},
		MaxScorePicker{},
	)
	if err == nil {
		t.Fatal("expected infinite weight error")
	}

	scheduler, err := NewScheduler(
		nil,
		[]WeightedScorer{{Scorer: fixedScorer{name: "nan", score: math.NaN()}, Weight: 1}},
		MaxScorePicker{},
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = scheduler.Schedule(context.Background(), Request{Model: "tiny-llm"}, []Endpoint{{
		ID: "one", URL: "http://one.example", Models: []string{"tiny-llm"}, Ready: true,
	}})
	if err == nil {
		t.Fatal("expected NaN score error")
	}

	endpoint := Endpoint{
		ID: "nan", URL: "http://nan.example", Models: []string{"tiny-llm"}, Ready: true,
		KVCacheUtilization: math.NaN(),
	}
	if err := endpoint.Validate(); err == nil {
		t.Fatal("expected NaN KV-cache utilization error")
	}
}

type fixedScorer struct {
	name  string
	score float64
}

func (s fixedScorer) Name() string { return s.name }
func (s fixedScorer) Score(context.Context, Request, Endpoint) (float64, error) {
	return s.score, nil
}
