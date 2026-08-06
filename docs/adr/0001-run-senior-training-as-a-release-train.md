# Run senior training as a release train

The senior track evolves one long-lived Order service through sequential
releases instead of using isolated exercises. Each release must preserve
committed data and explicitly manage compatibility, migration, rollout, and
rollback because those constraints create the time dimension that the earlier
levels intentionally lack.

## Considered Options

- Isolated infrastructure exercises are easier to reset, but they do not expose
  interactions between old data, old clients, and new code.
- One final production project has realism, but delays feedback and makes weak
  areas difficult to diagnose.
- A release train preserves history while keeping each assessment bounded.

## Consequences

Historical fixtures, API contracts, migration records, and incident reports
become immutable inputs to later releases. Rewriting history to make a new
release pass invalidates that release.
