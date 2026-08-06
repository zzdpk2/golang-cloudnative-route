# L22: Capacity and Cost

## Mission

Establish the service's capacity envelope, make overload behavior intentional,
and connect technical limits to a cost and reliability budget.

## Work

1. Build a demand model covering steady traffic, peak traffic, hot keys, event
   fan-out, data growth, and retention.
2. Define latency and error SLOs for at least one write and one read journey.
3. Load test until the first three saturation points are observed.
4. Add backpressure, concurrency limits, queue bounds, or load shedding at the
   correct ownership boundaries.
5. Define graceful degradation for non-critical behavior.
6. Estimate monthly cost at normal, peak, and failure-amplified load.
7. Control metric, log, trace, and tenant-label cardinality.

## Required Experiments

- steady load long enough to expose leaks;
- step load to locate the knee of the latency curve;
- burst load above admitted capacity;
- one hot Customer, Product, or Order key;
- slow database and slow event consumer;
- cache cold start;
- retry amplification during dependency degradation.

## Acceptance Evidence

- capacity model with assumptions and error bounds;
- reproducible load scripts and versioned test data;
- throughput, p50, p95, p99, errors, saturation, and cost on one report;
- a stated maximum admitted load and behavior above it;
- queue and concurrency bounds justified by memory and latency budgets;
- scaling test showing when more replicas help and when they do not;
- cost per successful Order at normal and peak load.

## Decision Questions

- Which work is rejected first under overload?
- Which customer-visible features may degrade?
- Is horizontal scaling limited by a shared lock, database, partition, or vendor?
- Can one tenant consume another tenant's capacity?
- Which telemetry is too expensive or too high-cardinality to retain?

## Invalid Shortcuts

- reporting only average latency;
- stopping the test before saturation;
- increasing every pool and queue size;
- autoscaling on CPU without observing the actual bottleneck;
- ignoring failed-request cost.
