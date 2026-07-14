# Code Overview — `ginrest` template (Gin REST API)

## Project Purpose

A starter REST API template written in Go using the **Gin** framework, part of **SunGo Project Manager**. It provides a ready-made HTTP server structure with basic CRUD operations, middleware, and a graceful shutdown mechanism.

---

## 1. Data Structure — `Item`

```go
type Item struct {
    ID    int
    Name  string
    Value string
}
```

A simple model representing a single resource. Data is stored **in memory** (`items []Item`), with no database — this is a placeholder meant to be replaced by your own persistence layer.

- `nextID` — a counter used to assign IDs to new items.
- `mu sync.RWMutex` — **(added in the fix)** protects `items`/`nextID` from concurrent access across goroutines handling HTTP requests.

---

## 2. Handlers (request-handling functions)

| Function | Method/Path | Behavior |
|---|---|---|
| `healthHandler` | `GET /health` | Returns server status, project name, and current UTC time — a typical health-check endpoint (monitoring, Kubernetes). |
| `listItems` | `GET /api/v1/items` | Returns the full list of items along with a count. Read protected by `mu.RLock()`. |
| `getItem` | `GET /api/v1/items/:id` | Returns a single item by ID. |
| `createItem` | `POST /api/v1/items` | Validates the incoming JSON (`name`, `value` required), creates a new `Item`, and appends it to the list. Write protected by `mu.Lock()`. |
| `deleteItem` | `DELETE /api/v1/items/:id` | Removes the item with the given ID. Write protected by `mu.Lock()`. |

### Fixes Applied to the Code

#### 🐛 Bug #1 — Incorrect ID comparison

**Original:**
```go
if id == string(rune('0'+item.ID)) {
```
This code compared the URL string against a single character produced by converting a digit to a `rune`. It only worked by coincidence for IDs 0–9; for IDs ≥ 10 it produced an incorrect/unpredictable result, so the item would never be found or deleted.

**Fix:**
```go
import "strconv"
...
if id == strconv.Itoa(item.ID) {
```
The item's ID is now correctly converted to a string and compared against the URL parameter — this works for any number of digits. Applied in both `getItem` and `deleteItem`.

#### 🐛 Bug #2 — Missing synchronization (race condition)

**Problem:** Gin handles requests concurrently by default (each in its own goroutine). The global variables `items` and `nextID` were being modified without any locking — under simultaneous `POST`/`DELETE` requests this could lead to:
- lost updates (overwritten changes),
- a panic from concurrent `append` and slice iteration,
- non-deterministic `nextID` values.

**Fix:** added a `sync.RWMutex`:
- `mu.RLock()` / `mu.RUnlock()` for reads (`listItems`, `getItem`),
- `mu.Lock()` / `mu.Unlock()` for writes (`createItem`, `deleteItem`).

This guarantees that only one goroutine at a time can modify the store, while reads can proceed concurrently with each other (but not with writes).

---

## 3. Router — `setupRouter()`

Builds a Gin instance (`gin.New()` — without the default middleware) and manually attaches:

1. **`gin.Logger()`** — logs every HTTP request (method, path, status, duration).
2. **`gin.Recovery()`** — recovers from panics in handlers, converting them into a `500` response instead of crashing the whole server.
3. **CORS** (`gin-contrib/cors`) — allows requests from any origin (`AllowOrigins: []string{"*"}`), with allowed methods `GET, POST, PUT, DELETE, OPTIONS`. `AllowCredentials: false` is consistent with `*` (browsers block `*` combined with `credentials: true` anyway).

Routes:
- `GET /health` — outside the API group.
- `/api/v1/...` — versioned API group with 4 CRUD endpoints for `items`.

---

## 4. `main()` Function — Server Startup and Shutdown

### Server Startup

```go
port := os.Getenv("PORT")
if port == "" { port = "8080" }
```
The port is configurable via an environment variable, defaulting to `8080`.

The server is constructed manually (`http.Server`) instead of via `router.Run()`, which allows setting:
- `ReadTimeout`, `WriteTimeout` — 10 s (protection against slow clients, e.g. slowloris-style attacks).
- `IdleTimeout` — 60 s for keep-alive connections.

The server listens in a **separate goroutine**, so the main thread can wait for a shutdown signal.

### Graceful Shutdown

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
```
The program blocks on the `quit` channel until it receives a `SIGINT` (Ctrl+C) or `SIGTERM` (e.g. `docker stop`, `kill`) signal.

Once the signal arrives:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
srv.Shutdown(ctx)
```
The server gets **5 seconds** to finish any in-flight requests before being forcibly shut down — a standard pattern for services running in containers/Kubernetes.

---

## 5. Request Flow

```
Client → CORS middleware → Logger → Recovery → Router (route matching)
       → Handler (e.g. createItem) → JSON validation → operation on []Item (under mutex)
       → JSON response
```

---

## 6. Further Considerations for a Real Project

- **Persistence**: `items` only lives in process memory — restarting the server means losing all data. Eventually hook up a real database (Postgres/SQLite/etc.).
- **ID validation**: there's currently no check that `:id` is actually numeric — `strconv.Itoa` on the comparison side works correctly, but it's worth also calling `strconv.Atoi(id)` on input and returning `400 Bad Request` for non-numeric IDs.
- **CORS `*`**: fine for dev/testing; in production it's worth restricting to specific origins.
- **Missing `PUT`/`PATCH`**: there's currently no endpoint to update an existing item.
- **Pagination**: `listItems` returns the entire list at once — with a larger number of records, consider adding `limit`/`offset` or `page` parameters.
