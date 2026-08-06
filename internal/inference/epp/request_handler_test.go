package epp

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestRequestHandlerRunsParserProducerAndAdmitter(t *testing.T) {
	t.Parallel()

	concurrency, err := NewConcurrencyAdmitter(1)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewRequestHandler(
		OpenAIChatParser{},
		[]DataProducer{RoutingHeaderProducer{}},
		[]Admitter{concurrency},
	)
	if err != nil {
		t.Fatal(err)
	}
	headers := http.Header{
		"X-Request-Id":  []string{"request-1"},
		"X-Prefix-Key":  []string{"shared"},
		"X-Priority":    []string{"7"},
		"X-Fairness-Id": []string{"tenant-a"},
	}
	body := []byte(`{"model":"tiny-llm","messages":[{"role":"user","content":"hello"}]}`)

	state, release, err := handler.Handle(context.Background(), headers, body)
	if err != nil {
		t.Fatal(err)
	}
	if state.Request.Model != "tiny-llm" ||
		state.Request.PrefixKey != "shared" ||
		state.Request.Priority != 7 ||
		state.Request.FairnessID != "tenant-a" {
		t.Fatalf("unexpected request state: %+v", state.Request)
	}

	_, _, err = handler.Handle(context.Background(), headers, body)
	if !errors.Is(err, ErrOverloaded) {
		t.Fatalf("second request error = %v, want ErrOverloaded", err)
	}
	release()
	release()
	if _, secondRelease, err := handler.Handle(context.Background(), headers, body); err != nil {
		t.Fatalf("capacity was not released: %v", err)
	} else {
		secondRelease()
	}
}

func TestOpenAIChatParserRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	_, err := (OpenAIChatParser{}).Parse(
		context.Background(),
		[]byte(`{"model":"tiny-llm","messages":[{"role":"user","content":"hello"}],"unknown":true}`),
	)
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
}

func TestRoutingHeaderProducerRejectsInvalidPriority(t *testing.T) {
	t.Parallel()

	handler, err := NewRequestHandler(
		OpenAIChatParser{},
		[]DataProducer{RoutingHeaderProducer{}},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = handler.Handle(
		context.Background(),
		http.Header{"X-Priority": []string{"urgent"}},
		[]byte(`{"model":"tiny-llm","messages":[{"role":"user","content":"hello"}]}`),
	)
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
	}
}
