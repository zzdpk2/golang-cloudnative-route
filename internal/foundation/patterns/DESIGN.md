# minigin Exercise Design

This package is a small HTTP framework core exercise. It is not intended to replace `net/http` or Gin. Its purpose is to expose the routing, context, middleware, recovery, and concurrency mechanisms that frameworks such as Gin build upon.

## Request Flow

```text
client
  |
  v
Engine.ServeHTTP(w, r)
  |
  |-- matchRoute(method, path)
  |      |
  |      +-- pattern: /users/:id
  |      +-- path:    /users/42
  |      +-- params:  {"id":"42"}
  |
  v
newContext(w, r, pattern, params, handlers)
  |
  v
global middleware
  |
  v
group middleware
  |
  v
route handler
  |
  v
Context.JSON / Context.String
  |
  v
ResponseWriter
```

## Core Model

```text
Engine
  - routes []routeEntry
  - global []HandlerFunc
  - stats *Stats
  - nextID uint64

RouterGroup
  - prefix string
  - handlers []HandlerFunc

Context
  - Writer http.ResponseWriter
  - Request *http.Request
  - Params map[string]string
  - handlers []HandlerFunc
  - index int
  - keys map[string]any
```

Every middleware and final route handler has the same type:

```go
type HandlerFunc func(*Context)
```

A middleware calls `c.Next()` to run the remaining chain, allowing work both before and after downstream handlers:

```go
func(c *Context) {
    // Run before downstream handlers.
    c.Next()
    // Run after downstream handlers.
}
```

## Exercise Stages

1. Routing fundamentals: implement `normalizePath`, `matchRoute`, `Engine.Handle`, and `Engine.ServeHTTP`. Use `TestNormalizeAndMatchRoute` and `TestRouteParamsQueryAndJSON`.
2. Context fundamentals: implement `newContext`, `Next`, `Abort`, response helpers, JSON binding, and context keys. Use `TestGroupMiddlewareOrderAndAbort` and `TestBindJSONAndContextKeys`.
3. Gin-style middleware chains: implement global and group middleware with the order `global -> group -> handler -> group after -> global after`.
4. HTTP details: implement `responseRecorder`, request IDs, recovery, and logging. Use `TestRecoveryRequestIDLoggerAndRecorder`.
5. Concurrency fundamentals: implement `Stats`, `StatsMiddleware`, and `AsyncLogSink`. Use `TestStatsMiddlewareIsConcurrencySafe` and tools such as `sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, channels, and atomics.
6. Request context: implement `TimeoutMiddleware`, `RequestContext`, and `WithRequestContext`. Use `TestTimeoutMiddlewareAddsDeadline`.

## Recommended Implementation Order

```powershell
go test ./pkg/minigin -run TestNormalizeAndMatchRoute
go test ./pkg/minigin -run TestRouteParamsQueryAndJSON
go test ./pkg/minigin -run TestGroupMiddlewareOrderAndAbort
go test ./pkg/minigin -run TestBindJSONAndContextKeys
go test ./pkg/minigin -run TestRecoveryRequestIDLoggerAndRecorder
go test ./pkg/minigin -run TestStatsMiddlewareIsConcurrencySafe
go test ./pkg/minigin -run TestTimeoutMiddlewareAddsDeadline
go test ./pkg/minigin
```

## Important Details

- `Context.Next()` controls the nested execution of the middleware chain.
- `Abort()` is not the same as `return`. It changes chain state, but the current function continues unless it explicitly returns.
- `responseRecorder` wraps and forwards to `ResponseWriter` while recording status and byte count.
- `Stats.Snapshot()` must return a map copy so callers cannot mutate internal state.
- `AsyncLogSink.Close()` must wait until its background goroutine drains the channel.
- `Request.Context()` represents the standard-library request lifetime. This package's `Context` is a framework-level request wrapper; they are distinct types with different responsibilities.
