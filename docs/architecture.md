# Architecture

## The shape of it

```
Browser ──HTTP/JSON──┐
   │                 ├──> seid (Go) ──SQL──> MySQL (statusengine_*, sei_*)
   └──SSE────────────┘        │
                              ├──POST /commands──> worker :8081  (external commands)
                              └──WS   /ws────────> worker :8080  (live events)
```

One process serves the API and the Angular bundle. Go, because the
Statusengine worker is Go, the deployment stays a single file, and the
event consumer plus fan-out belongs in a long-lived process rather than in
a request handler.

The browser never talks to the worker. Two reasons, and the second is the
one that decides it:

1. The worker's keys would have to reach the page, and a key in the page is
   a key in everybody's hands.
2. Browser JavaScript cannot set headers on a WebSocket handshake, so a
   browser client would be pushed onto the `?api_key=` query parameter -
   which the worker's own documentation calls a browser fallback, and which
   leaks into proxy logs and history.

So `seid` holds one authenticated WebSocket to the worker, subscribes to
the topics the UI actually uses, and fans them out to browsers over SSE
with the ordinary session cookie.

## Layers

| Package | Responsibility |
|---|---|
| `internal/config` | Defaults, YAML, environment, flags, validation |
| `internal/database` | The MySQL pool |
| `internal/migrate` | Embedded `sei_*` schema, applied under an advisory lock |
| `internal/auth` | argon2id, sessions, roles, permissions |
| `internal/httpapi` | Routing, middleware, handlers, static bundle |
| `internal/domain` | Types shared across repositories and handlers |
| `internal/repository/mysql` | Every query against `statusengine_*` |
| `internal/metrics` | `Provider` interface; MySQL now, Graphite later |
| `internal/commands` | Naemon command construction, worker client, audit |
| `internal/events` | Worker WebSocket consumer, hub, SSE endpoint |
| `internal/webui` | The embedded Angular bundle |

Three interfaces carry the parts that are meant to be replaceable:

- **`metrics.Provider`** — the worker can already route performance data to
  Graphite instead of MySQL, so this abstraction has a real second
  implementation waiting rather than being a speculative one.
- **`commands.Transport`** — the worker's HTTP endpoint today; replaceable
  with a direct Gearman publish or the `naemon.cmd` pipe, and with a fake
  in tests.
- **`events.Source`** — the worker's WebSocket today; a polling source
  covers the case where it is unreachable.

## Queries against a partitioned schema

The worker partitions its history tables by `time DIV 86400`:
`*checks`, `*_statehistory`, `*_notifications`, `*_notifications_log` and
`logentries`. A query without a bound on the partitioning column scans
every partition, which on a table of a few hundred thousand rows is the
difference between a page that loads and one that does not.

So every history endpoint **requires** a time window and defaults to the
last 24 hours when the caller does not give one. That is not a UI
convenience; it is what makes partition pruning happen.

`statusengine_perfdata` is not partitioned, but it carries
`metric (hostname, service_description, label, timestamp_unix)`. The chart
query is written in exactly that column order, and downsamples in SQL:

```sql
SELECT label, unit,
       FLOOR(timestamp_unix / ?) * ? AS bucket,
       AVG(value), MIN(value), MAX(value)
FROM statusengine_perfdata
WHERE hostname = ? AND service_description = ?
  AND timestamp_unix BETWEEN ? AND ?
GROUP BY label, unit, bucket
ORDER BY bucket
```

## Sorting and SQL

Sorting is the one place a query parameter would otherwise reach SQL as an
identifier, where placeholders cannot help. Every list endpoint declares a
whitelist of sortable columns, and anything outside it is a `400` naming
the allowed values. Everything else - filters, search, pagination - is a
bound parameter.

## Command semantics

The worker's `/commands` endpoint takes the broker's own envelope, with
`Command` and `Data` capitalised. Three operations have typed forms
(`schedule_check`, `check_result`, `delete_downtime`); everything else goes
as `raw`, a Naemon external command line, and the worker prefixes the
timestamp.

The author of a command always comes from the session, never from the
request body. Every field is checked for `;`, newlines and control
characters before assembly - the worker rejects control characters itself,
but a `;` inside a comment would silently shift every field after it.

**A `202` means the command reached the broker, not that Naemon ran it.**
Gearman acknowledges queueing; the broker has no reply path. So the API
returns `202` with the audit record written, and the UI shows "submitted",
then watches the affected object for a few seconds - via SSE, or by polling -
before it says "confirmed". When confirmation does not arrive, it says that
instead of claiming success.

## Frontend

Angular 21, standalone components, signals, zoneless change detection.
Tailwind v4 for the design tokens, Angular CDK for overlay and
accessibility primitives. No component library: the visual language here is
dense operational tables, which is not what a general-purpose kit is shaped
for.

Two rules run through the styling:

1. **State is structure.** A status is carried by a rail at the row's
   leading edge, a label, and the sort order - not by colour alone. The
   rail is solid for a hard state and striped for a soft one, which is the
   distinction that decides whether an operator acts now or waits.
2. **Monospace means a machine wrote it.** Hostnames, service descriptions,
   plugin output, performance data, command lines. Everything a person
   wrote is sans. An operator can tell a literal string from our prose
   without reading it.

Fonts are vendored rather than fetched from a CDN. A monitoring system
routinely runs on a segment with no route to the internet, and a UI that
degrades there degrades exactly when someone needs it.

Angular's critical-CSS inliner is switched off in the production build: it
emits `<link ... onload="this.media='all'">`, an inline event handler that
our `script-src 'self'` policy blocks, which would leave the real
stylesheet unloaded. The inliner saves one round trip on a same-host
deployment; the policy is worth more.
