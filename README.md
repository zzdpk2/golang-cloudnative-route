# Commerce to llm-d Router Exercises

This repository is a red, implementation-first course with one required path:

```text
Go foundations -> Commerce Order -> llm-d prerequisites -> llm-d Router
-> upstream llm-d capstone
```

## Start Here

Start the dynamic [SvelteKit learning console](web/learning-path/README.md), or
open the zero-install [HTML tracker](docs/learning-path.html) and clickable
[Markdown path](docs/LEARNING_PATH.md). Then run:

```bash
go run ./cmd/exercise next
```

Implement the reported contract yourself, run its focused command, and ask for
`next` again. `go run ./cmd/exercise check` succeeds only after all 36 required
gates pass. Tests define behavior; production files contain hints and TODO
starters, never completed exercise answers.

The runner synchronizes passing gate IDs to both trackers. Keep either tracker
open and it will refresh test-verified progress automatically. Six dashed
optional branches are placed inside their relevant foundation or Commerce
phase. Run `go run ./cmd/exercise optional` when you want to refresh their
independent evidence; they never affect the required 36-gate percentage.

## Four Internal Areas

- `internal/foundation`: consolidated Go language, testing, concurrency, and
  optional mechanism exercises.
- `internal/commerce`: the Order domain and its application, transport,
  persistence, event, and extension packages.
- `internal/inference`: Data Layer, scheduling, flow control, configuration,
  EPP, Proxy, runtime, and observability exercises.
- `internal/senior`: optional Commerce production-evidence contracts.

Delivery exercises live in `deploy/inference-lab`; black-box contracts live in
`test/acceptance`. Folder order is not learning order: the runner and
[learning path](docs/LEARNING_PATH.md) are authoritative.

Supporting references:

- [Router production labs](docs/llmd/PRODUCTION_LABS.md)
- [Router module responsibilities](docs/llmd/README.md)
- [Pinned upstream capstone](docs/llmd/UPSTREAM.md)
- [No-answer exercise policy](docs/llmd/EXERCISE_RULES.md)
- [Optional senior release train](docs/senior/README.md)

The Commerce Order and Inference Routing contexts share engineering practices,
but they do not share a domain model.
