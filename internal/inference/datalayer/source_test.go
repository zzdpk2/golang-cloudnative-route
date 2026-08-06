package datalayer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileSourceValidatesBeforePublishing(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "endpoints.json")
	fixture := `{
	  "version": 7,
	  "observed_at": "2026-07-31T00:00:00Z",
	  "endpoints": [
	    {"id":"model-b","url":"http://model-b","models":["tiny-llm"],"ready":true},
	    {"id":"model-a","url":"http://model-a","models":["tiny-llm"],"ready":true}
	  ]
	}`
	if err := os.WriteFile(path, []byte(fixture), 0o600); err != nil {
		t.Fatal(err)
	}
	source, err := NewFileSource(path)
	if err != nil {
		t.Fatal(err)
	}
	var updates []Update
	if err := source.Run(context.Background(), func(update Update) error {
		updates = append(updates, update)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(updates) != 2 || updates[0].EndpointID != "model-a" || updates[1].EndpointID != "model-b" {
		t.Fatalf("updates = %+v, want deterministic Endpoint ID order", updates)
	}
	for _, update := range updates {
		if update.Version != 7 || !update.ObservedAt.Equal(time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("fixture metadata was not preserved: %+v", update)
		}
	}
}

func TestFileSourcePublishesNothingFromInvalidFixture(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(path, []byte(`{
	  "version": 1,
	  "observed_at": "2026-07-31T00:00:00Z",
	  "endpoints": [
	    {"id":"valid","url":"http://valid","models":["tiny-llm"],"ready":true},
	    {"id":"","url":"http://invalid","models":["tiny-llm"],"ready":true}
	  ]
	}`), 0o600); err != nil {
		t.Fatal(err)
	}
	source, err := NewFileSource(path)
	if err != nil {
		t.Fatal(err)
	}
	published := 0
	err = source.Run(context.Background(), func(Update) error {
		published++
		return nil
	})
	if !errors.Is(err, ErrInvalidSource) || published != 0 {
		t.Fatalf("error=%v published=%d, want ErrInvalidSource and zero partial updates", err, published)
	}
}

func TestEventSourcePreservesOrderAndSinkFailure(t *testing.T) {
	t.Parallel()
	input := make(chan Update, 3)
	input <- Update{EndpointID: "one", Version: 1}
	input <- Update{EndpointID: "two", Version: 2}
	input <- Update{EndpointID: "three", Version: 3}
	close(input)
	source, err := NewEventSource(input)
	if err != nil {
		t.Fatal(err)
	}
	sinkErr := errors.New("sink stopped")
	var got []string
	err = source.Run(context.Background(), func(update Update) error {
		got = append(got, update.EndpointID)
		if update.EndpointID == "two" {
			return sinkErr
		}
		return nil
	})
	if !errors.Is(err, sinkErr) {
		t.Fatalf("Run error = %v, want sink failure", err)
	}
	if len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("delivery order = %v", got)
	}
}

func TestEventSourceHonorsCancellation(t *testing.T) {
	t.Parallel()
	input := make(chan Update)
	source, err := NewEventSource(input)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := source.Run(ctx, func(Update) error { return nil }); !errors.Is(err, context.Canceled) {
		t.Fatalf("Run error = %v, want context.Canceled", err)
	}
}
