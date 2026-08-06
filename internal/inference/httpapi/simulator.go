package httpapi

import (
	"net/http"
	"time"

	"github.com/rex/go-ddd-tdd/internal/inference/epp"
)

type Simulator struct{}

func NewSimulator(id, model string, delay time.Duration) (*Simulator, error) {
	panic("TODO(exercise): validate the simulator identity, model, and delay")
}

func (s *Simulator) Handler() http.Handler {
	panic("TODO(exercise): expose completion, health, readiness, and metrics routes")
}

func (s *Simulator) handleChatCompletions(http.ResponseWriter, *http.Request) {
	panic("TODO(exercise): validate input, honor cancellation, and emit an OpenAI-style response")
}

func (s *Simulator) handleMetrics(http.ResponseWriter, *http.Request) {
	panic("TODO(exercise): expose bounded active-request metrics")
}

func estimateTokens(messages []epp.Message) int {
	panic("TODO(exercise): use the documented deterministic approximation")
}
