package routing

import "context"

type Filter interface {
	Name() string
	Keep(context.Context, Request, Endpoint) (bool, error)
}

type Scorer interface {
	Name() string
	Score(context.Context, Request, Endpoint) (float64, error)
}

type WeightedScorer struct {
	Scorer Scorer
	Weight float64
}

type Picker interface {
	Pick(context.Context, Request, []ScoredEndpoint) (ScoredEndpoint, error)
}

// Scheduler owns the Filter -> Score -> Pick lifecycle.
type Scheduler struct{}

func NewScheduler(filters []Filter, scorers []WeightedScorer, picker Picker) (*Scheduler, error) {
	panic("TODO(exercise): reject invalid plugins, weights, and duplicate names; copy configuration")
}

func NewDefaultScheduler() *Scheduler {
	panic("TODO(exercise): assemble the documented learning Scheduling Profile")
}

func (s *Scheduler) Schedule(
	ctx context.Context,
	request Request,
	endpoints []Endpoint,
) (Decision, error) {
	panic("TODO(exercise): filter, score, rank, pick, and preserve explainability without mutating input")
}
