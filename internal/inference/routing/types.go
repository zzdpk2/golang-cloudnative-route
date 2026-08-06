package routing

import "errors"

var (
	ErrNoEndpoint       = errors.New("no eligible inference endpoint")
	ErrUnknownEndpoint  = errors.New("unknown inference endpoint")
	ErrInvalidScheduler = errors.New("invalid scheduler configuration")
)

// Request is the Scheduler's transport-independent input.
type Request struct {
	ID         string `json:"id,omitempty"`
	Model      string `json:"model"`
	PrefixKey  string `json:"prefix_key,omitempty"`
	Priority   int    `json:"priority,omitempty"`
	FairnessID string `json:"fairness_id,omitempty"`
}

func (r Request) Validate() error {
	panic("TODO(exercise): validate required model and priority invariants")
}

// Endpoint is an immutable scheduling snapshot of one Model Server port.
type Endpoint struct {
	ID                 string   `json:"id"`
	URL                string   `json:"url"`
	Models             []string `json:"models"`
	Ready              bool     `json:"ready"`
	QueueDepth         int      `json:"queue_depth"`
	KVCacheUtilization float64  `json:"kv_cache_utilization"`
	CachedPrefixes     []string `json:"cached_prefixes,omitempty"`
}

func (e Endpoint) Validate() error {
	panic("TODO(exercise): validate identity, URL, models, queue depth, and KV-cache ratio")
}

func (e Endpoint) SupportsModel(model string) bool {
	panic("TODO(exercise): report whether this immutable snapshot serves the model")
}

func (e Endpoint) HasPrefix(prefix string) bool {
	panic("TODO(exercise): an empty prefix must never create affinity")
}

func cloneEndpoint(endpoint Endpoint) Endpoint {
	panic("TODO(exercise): return a defensive copy of every reference field")
}

type ScoredEndpoint struct {
	Endpoint Endpoint           `json:"endpoint"`
	Score    float64            `json:"score"`
	Scores   map[string]float64 `json:"scores"`
}

type Decision struct {
	Selected   ScoredEndpoint   `json:"selected"`
	Candidates []ScoredEndpoint `json:"candidates"`
}
