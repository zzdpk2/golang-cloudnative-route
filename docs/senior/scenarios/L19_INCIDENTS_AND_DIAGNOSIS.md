# L19: Incidents and Diagnosis

## Mission

Diagnose production behavior from symptoms and evidence before reading the
suspected implementation. Restore service, explain the mechanism, and prevent
recurrence without hiding the signal.

## Incident Set

Build or ask a reviewer to seed at least four incidents:

1. a goroutine leak caused by cancellation or channel ownership;
2. lock contention on a hot Order or inventory path;
3. a cache stampede after expiry;
4. a retry storm that increases dependency load during failure;
5. an optional queue backlog caused by a slow or poison event.

The investigator initially receives only traffic shape, user symptoms, logs,
metrics, traces, profiles, deployment history, and dependency status.

## Response Procedure

1. State impact, severity, and the current safety risk.
2. Build a timeline from evidence.
3. List hypotheses and a discriminating observation for each.
4. Mitigate user impact before optimizing the final fix.
5. Prove the root cause with a minimal reproduction or profile.
6. Add a regression or fault test.
7. Write the incident report and update the runbook.

## Required Evidence

- time to detection, acknowledgement, mitigation, and recovery;
- CPU, heap, goroutine, mutex, or block profiles as relevant;
- request and dependency saturation correlated on one timeline;
- commands used to diagnose and mitigate;
- explanation of why existing tests and alerts missed the issue;
- corrective actions split into immediate, preventive, and systemic work;
- an owner and due date for every action.

## Acceptance

- mitigation is reversible and does not corrupt Orders;
- the fix removes the mechanism, not only the symptom;
- an alert is tied to user impact or exhausted capacity;
- the runbook lets another engineer perform the first safe actions;
- the same injected incident is detected earlier on the next run.

## Review Trap

A larger timeout, more retries, or a periodic restart can be a valid temporary
mitigation. It is not a root-cause fix unless the failure mechanism and added
load are accounted for.
