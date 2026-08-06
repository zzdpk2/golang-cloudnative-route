package datalayer

import (
	"context"
	"errors"
)

var ErrInvalidSource = errors.New("invalid endpoint source")

type Sink func(Update) error

type Source interface {
	Run(context.Context, Sink) error
}

// FileSource is a finite bootstrap source for the documented JSON fixture
// shape. It must validate the complete file before publishing any update.
type FileSource struct{}

func NewFileSource(path string) (*FileSource, error) {
	panic("TODO(exercise): reject an empty path without reading the file yet")
}

func (s *FileSource) Run(ctx context.Context, sink Sink) error {
	panic("TODO(exercise): decode one bounded strict fixture, validate it completely, then publish in deterministic order")
}

// EventSource adapts an ordered update stream into the same Source contract.
type EventSource struct{}

func NewEventSource(updates <-chan Update) (*EventSource, error) {
	panic("TODO(exercise): reject a nil event stream")
}

func (s *EventSource) Run(ctx context.Context, sink Sink) error {
	panic("TODO(exercise): preserve event order, stop on cancellation or sink failure, and handle stream closure")
}
