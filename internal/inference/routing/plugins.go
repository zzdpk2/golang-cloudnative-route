package routing

import "context"

type ReadyFilter struct{}

func (ReadyFilter) Name() string { panic("TODO") }
func (ReadyFilter) Keep(context.Context, Request, Endpoint) (bool, error) {
	panic("TODO(exercise): filter unready Endpoint snapshots")
}

type ModelFilter struct{}

func (ModelFilter) Name() string { panic("TODO") }
func (ModelFilter) Keep(context.Context, Request, Endpoint) (bool, error) {
	panic("TODO(exercise): filter Endpoints that do not serve the requested Model")
}

type QueueDepthScorer struct{}

func (QueueDepthScorer) Name() string { panic("TODO") }
func (QueueDepthScorer) Score(context.Context, Request, Endpoint) (float64, error) {
	panic("TODO(exercise): map non-negative queue depth monotonically into [0,1]")
}

type KVCacheUtilizationScorer struct{}

func (KVCacheUtilizationScorer) Name() string { panic("TODO") }
func (KVCacheUtilizationScorer) Score(context.Context, Request, Endpoint) (float64, error) {
	panic("TODO(exercise): prefer available KV-cache headroom")
}

type PrefixAffinityScorer struct{}

func (PrefixAffinityScorer) Name() string { panic("TODO") }
func (PrefixAffinityScorer) Score(context.Context, Request, Endpoint) (float64, error) {
	panic("TODO(exercise): reward reusable prefix state without treating empty prefixes as hits")
}

type MaxScorePicker struct{}

func (MaxScorePicker) Pick(
	context.Context,
	Request,
	[]ScoredEndpoint,
) (ScoredEndpoint, error) {
	panic("TODO(exercise): pick the maximum score with deterministic tie-breaking")
}
