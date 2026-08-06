package integrations

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

func TestProtoLikeMarshal(t *testing.T) {
	req := CreateOrderRequest{CustomerID: "cust-1"}
	data, err := MarshalProtoLike(req)
	if err != nil {
		t.Fatal(err)
	}

	var decoded CreateOrderRequest
	if err := UnmarshalProtoLike(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.CustomerID != "cust-1" {
		t.Fatalf("decoded = %+v", decoded)
	}
}

func TestStatusError(t *testing.T) {
	err := NewStatusError(CodeNotFound, "order missing")
	if err.Error() != "NOT_FOUND: order missing" {
		t.Fatalf("error = %q", err.Error())
	}
	if CodeFromError(err) != CodeNotFound {
		t.Fatalf("code = %s", CodeFromError(err))
	}
	if CodeFromError(nil) != CodeOK {
		t.Fatal("nil error should be OK")
	}
	if CodeFromError(errors.New("plain")) != CodeUnknown {
		t.Fatal("plain error should be UNKNOWN")
	}
}

func TestMetadata(t *testing.T) {
	ctx := NewOutgoingContext(context.Background(), Metadata{
		"request-id": {"req-1"},
	})

	md, ok := MetadataFromContext(ctx)
	if !ok {
		t.Fatal("metadata missing")
	}
	if md["request-id"][0] != "req-1" {
		t.Fatalf("metadata = %v", md)
	}
}

func TestUnaryInterceptors(t *testing.T) {
	var calls []string

	a := func(ctx context.Context, req any, handler UnaryHandler) (any, error) {
		calls = append(calls, "a-before")
		resp, err := handler(ctx, req)
		calls = append(calls, "a-after")
		return resp, err
	}
	b := func(ctx context.Context, req any, handler UnaryHandler) (any, error) {
		calls = append(calls, "b-before")
		resp, err := handler(ctx, req)
		calls = append(calls, "b-after")
		return resp, err
	}
	final := func(ctx context.Context, req any) (any, error) {
		calls = append(calls, "handler")
		return "ok", nil
	}

	resp, err := ChainUnaryInterceptors(final, a, b)(context.Background(), "req")
	if err != nil {
		t.Fatal(err)
	}
	if resp != "ok" {
		t.Fatalf("resp = %v", resp)
	}

	want := []string{"a-before", "b-before", "handler", "b-after", "a-after"}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("calls = %v, want %v", calls, want)
		}
	}
}

func TestDeadlineInterceptor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := DeadlineInterceptor(ctx, nil, func(ctx context.Context, req any) (any, error) {
		return "should not run", nil
	})
	if CodeFromError(err) != CodeDeadlineExceeded && CodeFromError(err) != CodeCanceled {
		t.Fatalf("code = %s err=%v", CodeFromError(err), err)
	}
}

func TestMetadataInterceptor(t *testing.T) {
	_, err := MetadataInterceptor(context.Background(), nil, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	if CodeFromError(err) != CodeInvalidArgument {
		t.Fatalf("missing metadata code = %s", CodeFromError(err))
	}

	ctx := NewOutgoingContext(context.Background(), Metadata{"request-id": {"req-1"}})
	resp, err := MetadataInterceptor(ctx, nil, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	if err != nil || resp != "ok" {
		t.Fatalf("resp=%v err=%v", resp, err)
	}
}

type fakeBulkStream struct {
	ctx      context.Context
	requests []*CreateOrderRequest
	reply    *BulkCreateReply
	index    int
}

func (f *fakeBulkStream) Context() context.Context { return f.ctx }
func (f *fakeBulkStream) Recv() (*CreateOrderRequest, error) {
	if f.index >= len(f.requests) {
		return nil, io.EOF
	}
	req := f.requests[f.index]
	f.index++
	return req, nil
}
func (f *fakeBulkStream) SendAndClose(reply *BulkCreateReply) error {
	f.reply = reply
	return nil
}

func TestFakeBulkStreamShape(t *testing.T) {
	stream := &fakeBulkStream{
		ctx: context.Background(),
		requests: []*CreateOrderRequest{
			{CustomerID: "a"},
			{CustomerID: "b"},
		},
	}

	req, err := stream.Recv()
	if err != nil || req.CustomerID != "a" {
		t.Fatalf("req=%v err=%v", req, err)
	}

	_ = stream.SendAndClose(&BulkCreateReply{Created: 2})
	if stream.reply.Created != 2 {
		t.Fatalf("reply = %+v", stream.reply)
	}
}

func TestContextDeadlineShape(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
	case <-time.After(20 * time.Millisecond):
		t.Fatal("deadline should fire")
	}
}
