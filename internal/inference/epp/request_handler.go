// Package epp models the request-side responsibilities of an Endpoint Picker.
// It is intentionally not an Envoy ext-proc implementation.
package epp

import (
	"context"
	"errors"
	"net/http"

	"github.com/rex/go-ddd-tdd/internal/inference/routing"
)

var (
	ErrInvalidRequest = errors.New("invalid inference request")
	ErrOverloaded     = errors.New("endpoint picker overloaded")
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

func (r ChatCompletionRequest) Validate() error {
	panic("TODO(exercise): validate the supported OpenAI-style request subset")
}

type RequestState struct {
	Request   routing.Request
	BodyBytes int
}

type Parser interface {
	Parse(context.Context, []byte) (RequestState, error)
}

type DataProducer interface {
	Name() string
	Produce(context.Context, http.Header, *RequestState) error
}

type Admitter interface {
	Name() string
	Admit(context.Context, RequestState) (release func(), err error)
}

type OpenAIChatParser struct{}

func (OpenAIChatParser) Parse(context.Context, []byte) (RequestState, error) {
	panic("TODO(exercise): decode one bounded strict JSON value and preserve stable error classification")
}

type RoutingHeaderProducer struct{}

func (RoutingHeaderProducer) Name() string { panic("TODO") }
func (RoutingHeaderProducer) Produce(
	context.Context,
	http.Header,
	*RequestState,
) error {
	panic("TODO(exercise): derive request ID, prefix, priority, and fairness without transport leakage")
}

type BodySizeAdmitter struct {
	MaxBytes int
}

func (BodySizeAdmitter) Name() string { panic("TODO") }
func (BodySizeAdmitter) Admit(context.Context, RequestState) (func(), error) {
	panic("TODO(exercise): enforce the EPP admission limit independently of the Proxy limit")
}

// ConcurrencyAdmitter is the basic fail-fast exercise. The production Flow
// Control exercise lives in internal/inference/flowcontrol.
type ConcurrencyAdmitter struct{}

func NewConcurrencyAdmitter(limit int) (*ConcurrencyAdmitter, error) {
	panic("TODO(exercise): construct a bounded fail-fast admitter")
}

func (*ConcurrencyAdmitter) Name() string { panic("TODO") }
func (a *ConcurrencyAdmitter) Admit(context.Context, RequestState) (func(), error) {
	panic("TODO(exercise): acquire or fail fast; release exactly once; honor cancellation")
}

type RequestHandler struct{}

func NewRequestHandler(
	parser Parser,
	producers []DataProducer,
	admitters []Admitter,
) (*RequestHandler, error) {
	panic("TODO(exercise): validate and defensively copy the Request Handler configuration")
}

func NewDefaultRequestHandler() *RequestHandler {
	panic("TODO(exercise): assemble the basic parser, producer, and admitters")
}

func (h *RequestHandler) Handle(
	ctx context.Context,
	headers http.Header,
	body []byte,
) (RequestState, func(), error) {
	panic("TODO(exercise): parse, produce, admit, and release acquired capacity in reverse order")
}
