# Robot Framework Acceptance Tests

These tests observe running services through public protocols. They must not
import Go packages or inspect internal maps.

## Setup

Install Robot Framework and any libraries declared by the suites, then start
the target service separately.

```bash
python -m pip install robotframework robotframework-requests
go run ./cmd/orderd
```

## Run

```bash
robot --dryrun --output NONE --log NONE --report NONE test/acceptance
robot --outputdir .robot-results test/acceptance
```

Start with smoke and one business flow. Add validation, conflict, cancellation,
and failure recovery only when the public behavior exists. Keep shared protocol
keywords in resource files and business expectations in test cases.

All suites and shared keywords are learner-owned red starters. Documentation
describes observable contracts and trade-offs; it does not contain completed
Robot implementations.

A useful failure message says what the user sent and observed. It does not say
which private function should have been called.

## Inference Router track

Router suites are in `test/acceptance/inference` and are learner-owned red
starters. See [../../docs/llmd/PRODUCTION_LABS.md](../../docs/llmd/PRODUCTION_LABS.md).
