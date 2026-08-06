# SvelteKit Learning Path

This is the dynamic UI for the repository's single Commerce-to-llm-d Router
journey. It does not own exercise state. The Go runner remains authoritative
and publishes verified gates to `docs/learning-progress.js`; this app reads that
evidence through a read-only route handler.

## Run

From `web/learning-path`:

```bash
npm ci
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

On Windows PowerShell with script execution disabled, use `npm.cmd` instead of
`npm`.

```powershell
npm.cmd run dev
```

For a production build, keep the same working directory:

```powershell
npm.cmd run build
npm.cmd start
```

In a second terminal at the Go module root, refresh verified evidence with:

```bash
go run ./cmd/exercise next
```

The six dashed optional branches have separate evidence. Refresh them only
when you choose to work on them:

```bash
go run ./cmd/exercise optional
```

An incomplete optional command may exit non-zero to show that its own tests are
still open. It never changes required progress and never blocks `next`.

The console polls every five seconds and also provides an immediate **Sync
progress** button. Overall progress, all four phases, every gate row, and the
selected gate show explicit test-evidence progress bars. Optional progress uses
an independent dashed meter and is excluded from the required percentage.
Manual notes stay in browser local storage and never become test evidence.

## Verification

```bash
npm run check
npm run validate
npm run build
```

The source viewer serves only the registered gate source files and two learning
guides, rejects every other repository path, and caps responses at one
megabyte. The original `docs/learning-path.html` remains a zero-install
fallback.
