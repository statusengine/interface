# Code tour

This is the document to read before auditing this codebase. It says what
runs when, which file owns which decision, and where to look for each
class of risk. `architecture.md` explains *why* the shape is what it is;
this one explains *where everything is*.

Read it in order once. After that, section 8 is the checklist.

---

## 1. The shape in one paragraph

One Go binary (`seid`) serves a JSON API and the Angular bundle that is
compiled into it. It reads monitoring data from MySQL tables the
Statusengine worker writes, and it sends operator commands to the worker's
HTTP endpoint, which puts them on Gearman for Naemon. It owns five tables
of its own, all prefixed `sei_`. Browsers talk only to `seid`; nothing in
the browser ever holds a worker credential.

```
Browser ──HTTP/JSON──┐
   │                 ├──> seid ──SQL (read)───> MySQL statusengine_*
   └──SSE────────────┘    │  └──SQL (write)──> MySQL sei_*
                          ├──POST /commands───> worker :8081 ──> Gearman ──> Naemon
                          └──WS   /ws─────────> worker :8080  (one connection per process)
```

## 2. The map

| Path | Lines | What it owns |
|---|---:|---|
| `cmd/seid/` | 495 | `serve`, `migrate`, `user`, `version`; flag parsing and shutdown |
| `internal/config/` | 309 | One `Config` struct; file, then env, then flags; `Validate()` |
| `internal/database/` | 44 | Opening the pool, nothing else |
| `internal/migrate/` | 171 | Embedded SQL, applied under a MySQL advisory lock |
| `internal/auth/` | 1086 | argon2id, sessions, roles, permissions, login rate limit |
| `internal/domain/` | 479 | The types the API speaks; no SQL, no HTTP |
| `internal/repository/mysql/` | 1829 | Every SQL statement against `statusengine_*` |
| `internal/metrics/` | 302 | `Provider` interface, MySQL adapter, Graphite stub |
| `internal/commands/` | 869 | Naemon command lines, worker client, audit |
| `internal/events/` | 575 | Worker WebSocket consumer, fan-out hub, SSE endpoint |
| `internal/httpapi/` | 2547 | Routing, middleware, handlers, response envelope |
| `internal/webui/` | 37 | `go:embed` of the built frontend |
| `frontend/src/` | 9411 | Angular 21 app (6565 TS + 2846 HTML) |

Tests: 2904 lines of Go, 1086 of TypeScript, plus `tools/a11y/` (three
Playwright scripts).

Dependency direction is strictly downward: `httpapi` → `repository` /
`commands` / `events` / `auth` → `domain`. Nothing under `internal/`
imports `httpapi`, and `domain` imports nothing of ours at all. A new
import that points the other way is the first thing to question in a
review.

## 3. Startup, in order

`cmd/seid/main.go:75` (`serve`) is the whole sequence, and it is worth
reading top to bottom before anything else:

1. `config.Load(args)` — defaults, then the YAML file, then `SEI_*`
   environment variables, then flags; `Validate()` last. The config file
   path itself can only come from a flag or the environment
   (`config.go:112`).
2. `logging.New` — slog, text or JSON.
3. `signal.NotifyContext` — SIGINT/SIGTERM cancel the context that every
   long-lived goroutine watches.
4. `database.Open` — pool limits from config, ping before continuing.
5. `migrate.Run` — pending `sei_*` migrations, under an advisory lock so
   two instances starting together cannot half-apply a schema.
6. `auth.Bootstrap` — creates the three built-in roles if missing, and
   reconciles their permissions with the definitions in code, logging
   what changed (`internal/auth/bootstrap.go`). Creates the demo account
   when `demo_mode` is on. **Never creates an administrator**: the first
   one is made with `seid user create`, so no default credential exists.
7. `auth.NewService` + `go authSvc.PruneSessions(ctx, time.Hour)`.
8. Warnings for the three states an operator should know about: no
   accounts yet, commands disabled, live updates disabled.
9. The event hub and the single worker WebSocket, only when
   `worker_events_key` is set.
10. `httpapi.New(...)` then `ListenAndServe(ctx)`, which on cancellation
    gives in-flight requests 15 seconds.

Shutdown is the reverse and is explicit. The single `os.Exit` in the
program is in `main()` (`main.go:30`), for an error before anything is
serving.

## 4. A request, outside in

`internal/httpapi/server.go:routes()` builds two muxes: `/api/` and
everything else (the UI). The chain wraps both, outermost first:

```
requestID → logging → recover → securityHeaders → timeout → mux
```

- **requestID** puts an ID in the context and the response header; every
  log line for that request carries it.
- **recover** turns a panic into a 500 plus a stack in the log, so one bad
  row cannot take the process down.
- **securityHeaders** (`middleware.go:144`) sets `nosniff`,
  `X-Frame-Options: DENY`, `Referrer-Policy: same-origin` and a CSP with
  `default-src 'self'`, `script-src 'self'`, `frame-ancestors 'none'`.
  There is no `unsafe-inline` for scripts; `style-src` allows inline
  styles because Angular emits them.
- **timeout** applies `query_timeout` to everything except
  `/api/v1/events`, which is meant to stay open (`server.go:isStreamingRequest`).

Inside `/api/`, each route declares its own permission:

```go
authed := func(perm string, h http.HandlerFunc) http.Handler {
    if perm == "" { return chain(h, s.authMiddleware) }
    return chain(h, s.authMiddleware, requirePermission(perm))
}
api.Handle("GET /api/v1/hosts", authed(auth.PermHostsRead, s.handleListHosts))
```

`server.go:125-190` is therefore the complete access-control table — every
route, with the permission it needs, in one screen. That is the first
thing to read in an authorization review, and the only place a route can
be added without a permission by accident.

`authMiddleware` (`middleware.go:160`) reads the session cookie, resolves
it, and distinguishes two failures: no or unknown session → **401**;
anything else, such as the database being unreachable → **503**. The
distinction matters operationally: answering 401 for an outage signs
everyone out of a working installation.

A handler then: parses paging and filters (`response.go:parsePage`, with a
per-endpoint whitelist of sortable columns), calls a repository, and
writes `{"data": [...], "meta": {...}}` or
`{"error": {"code", "message", "field"}}`. Handlers hold no SQL and
repositories hold no HTTP.

## 5. How Angular gets in

**Build.** `frontend/` is a standard Angular workspace. `ng build` writes
to `internal/webui/dist/` (see `frontend/angular.json` → `outputPath`).
`internal/webui/webui.go` embeds that directory:

```go
//go:embed all:dist
var embedded embed.FS
```

`FS()` returns the bundle only if `dist/index.html` actually exists, so a
binary built without a frontend says so instead of serving an empty shell.
The directory keeps a `.gitkeep`, because `go:embed` fails to compile
against a missing directory — a clean clone must still build.

**Serving.** `internal/httpapi/ui.go` handles everything that is not
`/api/`:

- a path that exists **and carries a content hash** (`main-3X6UOOBT.js`,
  `styles-X4HKQZJG.css`, the fonts under `media/`) →
  `Cache-Control: public, max-age=31536000, immutable`;
- a path that exists **without** one — `i18n/de.json`, `favicon.svg` —
  → `Cache-Control: no-cache` plus an ETag over the content. Those names
  do not change between builds, so caching them hard freezes whatever a
  browser loaded first. It did: the language files were served immutably
  for a while, and every page added afterwards rendered bare translation
  keys in any browser that had been there before;
- `index.html` → `Cache-Control: no-cache`, so it is always revalidated;
  otherwise a deploy never reaches an open tab (`ui.go:77`);
- a path that does not exist **and has no file extension** → `index.html`,
  because it is an Angular route;
- a path that does not exist **and has an extension** → 404, so a missing
  `.js` does not come back as HTML and produce a baffling syntax error.

**Development.** `ui_dir` (config, `SEI_UI_DIR`, or `-ui-dir`) serves a
build from disk instead of the embedded one (`main.go:194`); with
neither, `/` answers with a message telling you to run the dev server.
The dev server proxies `/api` to `127.0.0.1:8090`
(`frontend/proxy.conf.json`), so the cookie is same-origin there too.

**Bootstrap.** `frontend/src/main.ts` calls `bootstrapApplication(App, appConfig)`.
`app.config.ts` is the whole runtime wiring:

- `provideZonelessChangeDetection()` — no zone.js; every screen-visible
  value is a signal.
- `provideRouter(routes, withComponentInputBinding(), withInMemoryScrolling(...))`.
- `provideHttpClient(withFetch(), withInterceptors([authInterceptor]))`.
- `provideTransloco(...)` with an HTTP loader for `public/i18n/{de,en}.json`.
- `provideAppInitializer(...)` — resolves theme, language, `GET /meta`
  and `GET /auth/me` **before** the router runs, so the first paint is
  never a flash of the wrong theme or a login page for someone signed in.

Routing is lazy per feature (`app.routes.ts`), every route behind
`authGuard`, most also behind `permissionGuard('<perm>')`
(`core/auth/auth.guard.ts`). The guards are a courtesy: the server
enforces the same permission on every call, and the UI hiding a control
is not access control.

## 6. Four flows worth tracing

### 6.1 Login

`login.ts` → `POST /api/v1/auth/login` → `handlers_auth.go:handleLogin` →
`auth.Service.Login` (`service.go:82`):

1. Per-client attempt limiter; over the limit → `429`.
2. `UserByUsername`, then `VerifyPassword` — argon2id, PHC-formatted,
   parameters re-read from the stored hash, comparison in constant time
   (`password.go`). A missing user still pays the hash cost.
3. `newToken()` — 32 bytes from `crypto/rand`. The **token** goes to the
   browser; only its SHA-256 is stored (`service.go:232-248`). A dump of
   `sei_sessions` is not a set of live credentials.
4. Cookie: `HttpOnly`, `SameSite=Lax`, `Secure` when `secure_cookies` is
   on (`handlers_auth.go:119`). No token in `localStorage`, so XSS cannot
   read it.

CSRF: every state-changing route is `POST` with a JSON body, and
`SameSite=Lax` keeps the cookie off cross-site POSTs. There is no
separate CSRF token; that is the trade-off to check if the deployment
ever puts a browser-reachable origin in front of it.

### 6.2 A list page

`hosts-list.ts` constructs a `ListStore<HostStatus>` with a path, a default
sort and its filter keys. `core/list/list-store.ts` is the single piece of
state machinery behind every list: URL ⇄ filters, sorting, paging,
selection, refetch, error text. Worth reading in full — 358 lines that
decide the behaviour of eleven tables across ten pages.

- Everything that shapes the request lives in the URL, so an operator can
  paste a narrowed list into a ticket.
- An effect refetches when any of those change; a second effect refetches
  on `live.tick()` and on `commands.submitted()`.
- Out-of-order responses are dropped by sequence number, so a fast typist
  cannot be shown the results of a prefix of what they typed.
- `errorText()` translates the server's error code and falls back to the
  server's own sentence (`core/api/error-text.ts`).

Server side: `handleListHosts` → `parsePage` (whitelist) →
`repository/mysql/hosts.go` → `query.go`. **Every value a caller supplies
is a placeholder.** Exactly three things are ever spliced into SQL text,
and none of them can carry a request value:

- column expressions looked up in a `map[string]string` whitelist
  (`orderBy`, `query.go:72`), checked in the handler and again here, so a
  future caller that skips the handler cannot get past it;
- the `?,?,?` run for an `IN` list, built from the *number* of values
  (`query.go:44`);
- two table names in `summary.go:120`, both compile-time literals passed
  from `summary.go:28` and `:32`.

Search values go through `likeEscape` (`query.go:94`) before they reach a
`LIKE`. That is the whole SQL surface.

### 6.3 A command

Click → `core/commands/commands.service.ts` `run({action, targets, body, verify})`
→ `POST /api/v1/commands/<action>` → `handlers_commands.go`:

1. `requirePermission` already ran.
2. `targetsRequest.resolve()` — dedupes, caps at `MaxBulkCommands` (1000).
3. `submit()` (`handlers_commands.go:131`) builds **every** envelope
   first. If one fails validation, nothing is sent and the error names the
   object: a partial success over fifty objects is worse than a refusal.
4. `commands.Bulk` flattens to one submission.
5. `commands.Client.Submit` — `POST` to the worker with `X-Api-Key`, under
   `worker_timeout`.
6. `recordCommands` writes **one audit row per target**, with the payload,
   the HTTP status and the response — including for every failure path
   above.
7. `202` back, with `verify` naming the objects to watch.

The command line itself is built in `internal/commands/naemon.go`. Two
rules live there and are the security-relevant part:

- **The author always comes from the session**, never from the request
  body.
- **`checkField` rejects `;`, newlines and control characters** in every
  field before assembly. Naemon splits command fields on `;` with no
  escape, so one inside a comment would silently shift every field after
  it. The interface refuses instead of rewriting what someone typed.

`202` means the broker accepted it, not that Naemon ran it. The UI then
polls the affected objects and only then says "confirmed"
(`commands.service.ts`).

### 6.4 Live updates

`events.WorkerSource` holds one WebSocket to the worker for the whole
process, with backoff, and pushes into `events.Hub`, which fans out to SSE
clients (`GET /api/v1/events`, session-authenticated like everything
else). The browser never sees the worker: its key would have to reach the
page, and browser JavaScript cannot set a header on a WebSocket
handshake.

The stream says *what* changed, never what it changed to. Consumers
refetch the endpoint they are already rendering, so a pushed payload can
never drift from what the REST endpoint returns. With no key, or when the
stream drops, `core/events/live.service.ts` polls every 30 seconds and the
top bar says which mode is in force.

## 7. The data boundary

| | Read | Write |
|---|---|---|
| `statusengine_*` (worker's) | yes | **never** |
| `sei_*` (ours) | yes | yes |

`internal/migrate/migrate_test.go:74` fails the build if any migration so
much as mentions `statusengine_`. Grepping for `INSERT`, `UPDATE` or
`DELETE` in `internal/repository/mysql/` should return nothing — those
verbs only appear in `internal/auth/store.go` and
`internal/commands/audit.go`, both `sei_*` only.

## 8. Audit checklist

| Question | Where to look |
|---|---|
| Which routes exist and what does each require? | `internal/httpapi/server.go:125-190` — the complete table |
| Can a route be reached unauthenticated? | Only those registered before the `authed` helper: `healthz`, `readyz`, `meta`, the three `auth/*` |
| How are passwords stored? | `internal/auth/password.go` — argon2id, RFC 9106 second option, PHC format |
| How are sessions stored? | `internal/auth/service.go:232-248` — random token out, SHA-256 in the database |
| Cookie flags | `internal/httpapi/handlers_auth.go:119` |
| Brute-force protection | `internal/auth/service.go:82` + `login_rate_limit`, `login_rate_window` |
| SQL injection surface | `internal/repository/mysql/query.go` — placeholders everywhere, whitelist in `orderBy`, `escapeLike` for `LIKE` |
| Command injection surface | `internal/commands/naemon.go` — `checkField`, and the author from the session |
| What is logged, and does it leak? | `internal/httpapi/middleware.go` (request log), `internal/commands/audit.go` (payloads: comments and times, no credentials) |
| Error messages to the client | `internal/httpapi/response.go` + `internalError` — a code and a sentence, never a driver error or a query |
| CSP and headers | `internal/httpapi/middleware.go:144` |
| Secrets | `internal/config/config.go` — `mysql_dsn`, `worker_command_key`, `worker_events_key`. Check `handleMeta` (`handlers_health.go:73`): it is the only thing an anonymous caller reads, and it carries no secret |
| Denial of service | `max_page_size`, `query_timeout`, `MaxBulkCommands`, `maxVerifyTargets`, SSE client buffers in `internal/events/hub.go` |
| Static assets and caching | `internal/httpapi/ui.go` — `fingerprinted` decides immutable vs. revalidate |
| Frontend XSS | No `innerHTML`, no `bypassSecurityTrust*` anywhere in `frontend/src` — Angular's interpolation escapes everything. Worth re-grepping after any change |
| Who can read the command log | `audit:read`: admin and operator, not guest (`internal/auth/bootstrap.go`) |

Three greps that should each return nothing, and are quick to re-run:

```bash
# Nothing in the UI writes unescaped HTML.
grep -rn "bypassSecurityTrust\|innerHTML" frontend/src

# Nothing in the read path writes. The three files that do write are
# internal/auth/store.go, internal/commands/audit.go and
# internal/migrate/migrate.go - all sei_* only.
grep -rnE "INSERT INTO|UPDATE [a-z_]+ SET|DELETE FROM" internal/repository/mysql/

# Two hits, both expected: the IN placeholder run and a map key built
# from a scanned state number. A third is worth a look.
grep -rn "fmt.Sprintf" internal/repository/mysql/*.go | grep -v _test
```

## 9. Invariants a change must not break

1. **The interface never writes to `statusengine_*`.** Enforced for
   migrations by a test; for queries, by review.
2. **Every route names its permission at registration.** A handler that
   checks permissions internally hides the rule from the table.
3. **The author of a command comes from the session.**
4. **No user input reaches SQL except as a placeholder**, with the
   documented exception of whitelisted column expressions.
5. **A list's state lives in the URL**, so links are shareable.
6. **A `202` is never reported as success**; confirmation is observed.
7. **A bulk is all or nothing.**
8. **Both language files stay symmetric** and every key a template uses
   resolves (section 10 has the check). Nothing user-facing is an
   English literal in code: the time pipes take their wording and their
   locale from the active language, and a date is written the way that
   language writes dates.
9. **Only a fingerprinted file may be cached immutably.**
10. **State is never carried by colour alone** — a rail, a label and the
    sort order carry it too.

## 10. Running the checks

```bash
make test                      # Go, frontend, translations
make lint                      # go vet, gofmt, prettier
make test-integration SEI_TEST_DSN='user:pass@tcp(127.0.0.1:3306)/statusengine'
make test-a11y SEI_URL=http://127.0.0.1:8090 SEI_USER=admin SEI_PASS=...
```

What each one actually proves:

- **Go unit tests** (2904 lines): command construction against expected
  Naemon lines, the permission matrix, argon2 round-trips, config
  precedence, the query builder, the SSE hub, migrations.
- **Integration tests**: the SQL against a real schema. The repository
  ones only read; the audit one writes rows under a username nothing else
  uses and deletes exactly those again, so point it at a test database.
- **Frontend tests** (80): the list store, pipes, the API error mapping,
  metric formatting.
- **a11y**: axe-core over every page in both themes, every dialog, and a
  keyboard-only walkthrough asserting each tab stop is visible and has a
  focus ring.
- **Translations** (`tools/i18n/check.mjs`, `make test-i18n`): the two
  files hold the same keys, every key written out in the code resolves in
  both, no key is empty, and the keys built at runtime are complete
  against the values the backend can actually send. That last list is
  maintained by hand next to the Go file it mirrors - adding a state or
  an error code means adding it there too, and the check will say so.
  Currently 418 keys, 307 of them referenced by name.

## 11. What is deliberately not here

- **No host groups, service groups or contacts.** They are not in MySQL;
  only Naemon's `objects.cache` has them, and the worker is the data
  source. Lists filter by name, state and flags.
- **No per-host authorization.** Roles are global, which follows from the
  above.
- **No comments** beyond those attached to acknowledgements and
  downtimes: the worker has no topic for them.
- **Graphite** is a documented stub (`internal/metrics/graphite/`).
- **Downtimes are always "from now", fixed**; the builder supports a
  start, an end and flexible windows, the dialog does not.
- **No acknowledgement expiry**, though Naemon has
  `ACKNOWLEDGE_*_PROBLEM_EXPIRE`.
- **No user management in the UI**; `seid user` only, and nobody can
  change their own password from the browser.
- **The dead placeholder page** (`features/placeholder-page.ts`,
  `shared/ui/not-built-yet.ts`) is no longer routed and can be deleted.
