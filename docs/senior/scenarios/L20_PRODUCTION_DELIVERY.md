# L20: Production Delivery

## Mission

Deliver the Order service through a repeatable pipeline and operate it during
rollout, dependency failure, and rollback.

## Tooling

Use tools because they satisfy the acceptance criteria, not because their names
belong on a checklist. A reasonable local stack is:

- a multi-stage container image running as a non-root user;
- Kind or k3d for a reproducible Kubernetes environment;
- Helm for environment-specific release configuration;
- GitHub Actions for build, test, security, and deployment gates;
- Prometheus and Grafana for metrics and dashboards;
- OpenTelemetry for traces and context propagation;
- k6 for load and rollout traffic;
- Toxiproxy or equivalent fault injection when useful.

Equivalent tools are acceptable when the evidence remains reproducible.

## Work

1. Define service-level indicators and an SLO for the Order write path.
2. Separate startup, liveness, and readiness semantics.
3. Implement graceful shutdown and bound all draining periods.
4. Create a versioned deployment artifact and immutable image reference.
5. Gate rollout on tests, migration safety, and live service indicators.
6. Perform a canary or staged rollout followed by automatic or manual promotion.
7. Roll back application code after a forward database migration.

## Required Failure Drills

- terminate a pod while requests are in flight;
- make the database unavailable without causing a restart loop;
- make telemetry unavailable and observe business-path behavior;
- deploy an intentionally bad version and stop the rollout using SLO evidence;
- exhaust a memory or CPU limit and explain the observed scheduling behavior;
- rotate a secret without rebuilding the image.

## Acceptance Evidence

- zero unexplained failed requests during a normal rolling update;
- readiness removes traffic before shutdown begins;
- liveness does not restart healthy processes merely because a dependency is
  unavailable;
- every deployed image maps to source and test evidence;
- the rollback procedure has been executed;
- alerts name an owner and link to a tested runbook;
- dashboards distinguish traffic, errors, latency, saturation, and dependency
  contribution.

## Invalid Shortcuts

- `latest` image tags;
- liveness and readiness using the same dependency-heavy check;
- a Helm chart whose only proof is `helm template`;
- a green pipeline that cannot identify which migration or image is running;
- declaring zero downtime without sending traffic during rollout.
