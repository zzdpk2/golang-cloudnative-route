# Repository Guidance for Claude

This repository is a hands-on learning path for Go, concurrency, DDD/TDD,
cloud-native engineering, Kubernetes, and llm-d Router development.

## Learning integrity

- Do not implement learner-facing `TODO` exercises unless the user explicitly asks for a full solution.
- Prefer explanations, diagnostic questions, and progressively stronger hints.
- Preserve failing tests that intentionally define unfinished exercises.
- Do not weaken, skip, delete, or rewrite tests merely to make an exercise pass.
- Keep required and optional exercises distinct. Optional work must not affect required progress.

## Change discipline

- Read the nearest README and tests before changing an exercise contract.
- Keep the learning path single-directional and update its tracker when exercise metadata changes.
- Keep all repository content in English.
- Make focused changes and avoid unrelated restructuring.
- Run the narrowest relevant tests first, then broader checks when appropriate.

## Repository hygiene

- Never commit secrets, credentials, local environment files, generated build output, dependency directories, local progress, or archived attempts.
- Use feature branches for remote changes and submit changes for review.
- Explain any production trade-offs involving concurrency, cancellation, ownership, reliability, observability, or Kubernetes behavior.
