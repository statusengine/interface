# Running it

## Deployment

One binary. It serves the API and the frontend on the same port, and it
needs nothing on the machine but a CA bundle and a clock.

```
seid migrate            # apply the sei_* schema, then exit
seid user create -username admin -role admin
seid serve
```

### systemd

`listen_addr` defaults to loopback. Put a reverse proxy in front of it
for TLS, or set `SEI_LISTEN_ADDR` and accept that the process holds a key
that can drive the monitoring core.

```ini
[Unit]
Description=Statusengine Web Interface
After=network-online.target mysql.service
Wants=network-online.target

[Service]
Type=simple
User=statusengine
Environment=SEI_CONFIG=/etc/statusengine/seid.yaml
ExecStart=/usr/local/bin/seid serve
Restart=on-failure
RestartSec=5s

# It reads one config file and talks to MySQL and the worker. Nothing
# else is needed, so nothing else is allowed.
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadOnlyPaths=/etc/statusengine

[Install]
WantedBy=multi-user.target
```

Behind nginx, one setting matters beyond the usual proxy headers:

```nginx
location /api/v1/events {
    proxy_pass http://127.0.0.1:8090;
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_buffering off;      # without this the event stream is held in a buffer
    proxy_read_timeout 1h;    # the stream is meant to stay open
}
```

The daemon already sends `X-Accel-Buffering: no`, which nginx honours,
but `proxy_read_timeout` is still yours to set.

## When something breaks

| Symptom | What it means | What the interface does |
|---|---|---|
| `/readyz` reports `mysql: unreachable` | The database is down or unreachable | `/healthz` still answers 200, so an orchestrator does not restart a healthy process. Signed-in users get `503`, **not** `401` - their sessions survive and come back with the database. |
| Every page says "could not verify your session" | Same as above, seen from a browser | Wait. Nothing needs restarting; the pool reconnects on its own. |
| A bulk answers `400` naming one object | One selected object failed validation | Nothing was submitted. Fix that object, or deselect it. |
| A command answers `502` with "could not reach the Statusengine worker" | The worker is down, or `worker_command_url` is wrong | Monitoring data keeps working - it comes from MySQL, not the worker. The attempt is still recorded in `sei_command_audit`. |
| A command answers `503` with "external commands are switched off" | No `worker_command_key` | The worker leaves `/commands` unserved without one. |
| The top bar says "Polling" instead of "Live" | The event stream is not available | The UI refreshes every 30 seconds instead. This is a working state, not an error - a deployment with no `worker_events_key` never leaves it. |
| A list answers `504` "took longer than 20s" | A query hit `query_timeout` | Narrow the window, name a host, or ask for fewer rows. See the performance notes below. |
| The log entries page is empty | `LogData` is not enabled in the broker | Add `LogData = "statusngin_logentries"` to `statusengine.toml` and restart Naemon. |

### Overlapping downtimes

Windows stack. Scheduling a downtime on an object that is already in one
increments `scheduled_downtime_depth` rather than replacing it, and
notifications stay suppressed until the last of them ends. The detail
page lists every window on the object for that reason, and says so when
there is more than one - cancelling one of three changes nothing an
operator can see, and a page that showed only the first would make that
look like a bug.

### Limits on a bulk

The worker accepts at most 1 000 commands in one submission. A downtime
that covers a host's services is two commands per host, so the ceiling
in objects can be half that; the refusal says which limit was hit and by
how much.

Above five objects the interface stops polling each one for
confirmation - the list refreshes anyway, from the event stream or the
polling fallback, and fifty status requests to confirm one command
would cost more than the command did.

## What it costs

Measured against MySQL 8.4 with a 512 MB buffer pool, on a database
holding 5 000 hosts, 60 000 services, 1.8 million service checks, 1.7
million performance-data points and 500 000 log entries. Warm cache,
best of three:

| Page | Time |
|---|---|
| Hosts, sorted by severity | 10 ms |
| Services, sorted by severity | 18 ms |
| Services, free-text search | 16 ms |
| Problems (the host/service union) | 39 ms |
| Dashboard summary | 19 ms |
| Log entries, 24 hours | 9 ms |
| Performance chart, 24 hours downsampled | 8 ms |
| Check history **scoped to one service**, 30 days | 8 ms |
| Check history **across everything**, 6 hours | 246 ms + 212 ms for the count |

The last row is the one to understand. The history tables are clustered
on an object-first primary key, so every check for one service sits
physically together: reading one object is a range read, reading every
object over the same window is not. That is why the API allows a 90-day
window when a host is named and caps it at 6 hours when none is, and why
the history pages lead with the object filter.

On a cold buffer pool - a freshly restarted MySQL - the unscoped history
query takes seconds rather than milliseconds. `query_timeout` (20s by
default) bounds it and answers with something actionable instead of
holding the connection.

## Backups

The interface owns the `sei_*` tables and nothing else:

```
sei_schema_migrations  sei_roles  sei_users  sei_sessions  sei_command_audit
```

Backing those up preserves accounts, roles and the command audit trail.
Everything else in the database belongs to the Statusengine worker and
is reproduced from the monitoring core.

`sei_sessions` is safe to truncate at any time; it signs everyone out.

## Upgrades

`seid serve` applies pending migrations at startup, under a MySQL
advisory lock, so two instances starting together cannot half-apply a
schema. `seid migrate` does the same and exits, for a deployment that
would rather migrate as a separate step.

Migrations only ever create or alter `sei_*` objects. An upgrade never
touches the worker's tables.

## Accounts

```
seid user create -username ops -role operator
seid user passwd -username ops     # also ends that user's sessions
seid user role   -username ops -role admin
seid user list
```

There is no default administrator and no default password. Passwords are
argon2id; the session cookie carries a random token and the database
stores only its SHA-256, so a dump of `sei_sessions` is not a set of live
credentials.
