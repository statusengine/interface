# Statusengine Web Interface

An operations interface for [Statusengine](https://statusengine.org) and Naemon.
It reads monitoring data from MySQL, submits external commands through the
Statusengine worker, and ships as a single binary with the frontend embedded.

Naemon, the broker module, the Statusengine worker and MySQL are upstream
systems. This repository is the web interface and its API.

## Status

| Phase | Scope | State |
|---|---|---|
| 1 | Scaffolding, authentication, roles, demo mode, responsive shell | done |
| 2 | Dashboard, hosts, services, problems, downtimes, acknowledgements, log entries | done |
| 3 | History pages, performance charts, metrics provider abstraction | done |
| 4 | External commands, live updates with polling fallback | done |
| 5 | Accessibility pass, edge cases, hardening | done |

Pages from a later phase are already routed and permission-guarded; they
say which phase they belong to rather than showing a spinner.

## Requirements

- Go 1.24 or newer (built and tested with 1.26)
- Node 20.19+ or 22.12+ (Angular 21)
- MySQL 8.0 or newer - the database the Statusengine worker writes to
- A running Statusengine worker, for external commands and live updates

## Quick start

```bash
cp seid.example.yaml seid.yaml    # then edit mysql_dsn and the worker keys
make deps                         # install frontend dependencies
make build                        # build the frontend and embed it in ./seid

./seid migrate                    # create the sei_* tables
./seid user create -username admin -role admin
./seid serve
```

The interface is then on <http://127.0.0.1:8090>.

There is no default administrator and no default password: the first
account is made with `seid user create`, so a shipped credential never
exists to be forgotten.

### Developing

Run the API and the Angular dev server side by side; the dev server proxies
`/api` to the backend, so the browser sees one origin and the session
cookie works normally.

```bash
make dev-api     # :8090, API only
make dev-ui      # :4200, with hot reload
```

### Docker

```bash
docker compose -f deploy/docker-compose.yml up --build
```

That brings up MySQL and the interface. The worker stays outside the stack -
set `SEI_WORKER_COMMAND_KEY` and `SEI_WORKER_EVENTS_KEY` in the environment
to connect to it.

## Documentation

- [Architecture](docs/architecture.md) - the shape of it, and why the
  queries look the way they do.
- [Running it](docs/operations.md) - deployment, an outage-symptom table,
  measured page costs at 60 000 services, backups and upgrades.

## Configuration

Settings resolve in this order, each beating the one before it:

1. built-in defaults
2. the YAML file named by `-config` or `SEI_CONFIG`
3. environment variables (`SEI_` + the key in upper case)
4. command line flags

See [`seid.example.yaml`](seid.example.yaml) for every key. A misspelled key
in the file is a startup error rather than a silent default.

Two settings are worth calling out:

- **`listen_addr` defaults to loopback.** This process holds a key that can
  drive the monitoring core. Reaching the network should be a decision, not
  something that happens by leaving a setting alone. The container image
  binds `:8090` instead and leaves exposure to the port mapping.
- **`worker_command_key` and `worker_events_key` are empty by default.**
  Without them the interface starts with commands and live updates switched
  off, and the UI says so rather than offering buttons that always fail.
- **`query_timeout` (20s) bounds one request's database work.** Without
  it a query that cannot finish holds a pooled connection and a browser
  tab indefinitely. The live event stream is exempt; it is meant to stay
  open.

## Accounts and roles

| Role | What it can do |
|---|---|
| `admin` | Everything, including managing accounts |
| `operator` | Read everything and submit external commands |
| `guest` | Read only; every external command is refused |

```bash
seid user create -username ops -role operator
seid user passwd -username ops          # also ends that user's sessions
seid user role   -username ops -role admin
seid user list
```

Passwords are hashed with argon2id. Sessions are server-side; the cookie
carries a random token and the database stores only its SHA-256, so a dump
of `sei_sessions` is not a set of live credentials.

### Demo mode

Setting `demo_mode: true` creates a `guest` account and offers it on the
login page as one click. It signs in through the ordinary session path.

By default it can only read, and the server refuses every command it
submits - the UI hides those controls as a courtesy, but the refusal is
what enforces it. A deployment that wants visitors to try commands names
them one by one:

```yaml
demo_commands: [acknowledge, downtime, reschedule]
```

That builds a `demo` role of read access plus exactly those, with its
own rate limit and a smaller ceiling on how many objects one command may
address. `notify` is refused whatever you write there, because a custom
notification reaches the real contacts of the monitored objects.
`docs/security.md` has the reasoning and the checklist for putting this
on the open internet.

## The dashboard

The top of the dashboard answers the questions somebody asks who is not
going to read a table: how much of the estate is available, how much of
what is broken nobody has taken on, how much of it moved at all in the
window, and how often somebody was actually alerted - with the hour by
hour shape of those alerts and a comparison against the window before.
Every percentage carries its denominator, because "50% of hosts are
down" has meant one host of two more than once.

The window is 24 hours or 7 days. Both are cheap: the figures come from
the status tables, which hold one row per object, and from an indexed
range on the notification tables. Counting individual state changes
would mean scanning a history table that has no index on its time column
alone, so the dashboard says how many objects changed rather than how
many times - an object that flapped forty times counts once, and the
page stays fast on an estate of any size.

## External commands

Operator actions go to the worker's `/commands` endpoint, which hands
them to the broker and on to Naemon. Supported: acknowledge and remove
an acknowledgement, schedule and delete a downtime, force a check,
submit a passive result, send a custom notification, and switch active
checks, passive checks, notifications, flap detection or the event
handler on and off per object.

Those five switches sit on the detail page as controls, not readings:
the value is the button, and clicking it sends the matching
`ENABLE_`/`DISABLE_` command and waits for the object to report the new
value back. A reader without the permission sees the same five values
as plain text.

A host or service that is in a downtime or acknowledged says so on its
detail page, with the record behind it: who set it, what they wrote,
and how long it lasts - plus the way to cancel the window or remove the
acknowledgement. Both suppress notifications, and the reason is the
part an operator needs.

Hosts, services and problems can be ticked and acted on together:
select rows, then acknowledge, schedule a downtime, or force a check on
all of them. The three that read the same for one object and for fifty -
toggling notifications across a mixed selection does not, so it is not
offered there.

Three things worth knowing:

- **A `202` means the command reached the broker, not that Naemon ran
  it.** The queue acknowledges the publish, the broker module has no
  reply path. So the UI says "submitted", then watches the object for
  about twelve seconds and only then says "confirmed" - and says
  "submitted, not confirmed" when it cannot see the change.
- **A bulk is all or nothing.** If one selected object fails validation,
  nothing is submitted and the error names it. A partial success over
  fifty objects leaves an operator working out which three did not take,
  at the moment they can least afford it.
- **Comments cannot contain a semicolon.** Naemon splits command fields
  on it with no escape, so one inside a comment would truncate the field
  and shift everything after it. The interface refuses rather than
  silently rewriting what you typed.

Every submission is recorded in `sei_command_audit`, including the ones
that were refused, with the submitting account and the response - one
row per object, so a bulk of fifty leaves fifty rows.

The **Command log** page reads that back: who sent what, at which
object, from which address, what the broker answered, and the payload as
it was sent. It filters by action, by person, by window, or down to the
failures alone, and a host or service page links into it pre-filtered to
that object. Administrators and operators can read it; guests cannot,
because it names people and their addresses.

## Live updates

With `worker_events_key` set, the daemon holds one WebSocket to the
worker and fans changes out to browsers over Server-Sent Events. The
browser never talks to the worker directly: its key would have to reach
the page, and browser JavaScript cannot set a header on a WebSocket
handshake.

The stream says *what* changed, not what it changed to - the UI refetches
the endpoint it is already rendering, so there is one source of truth.
Without the key, or when the stream drops, the UI polls every 30 seconds
and the indicator in the top bar says which mode it is in. Polling is a
working state, not an error.

## Database

The interface **reads** the worker's `statusengine_*` tables and never
writes to them. It **owns** the tables it creates itself, all prefixed
`sei_`:

| Table | Holds |
|---|---|
| `sei_schema_migrations` | Which migrations have run |
| `sei_roles` | Role definitions and their permissions |
| `sei_users` | Accounts |
| `sei_sessions` | Active sessions, keyed by token hash |
| `sei_command_audit` | Every external command submitted, including refused ones |

Migrations are embedded in the binary and applied at startup, under a MySQL
advisory lock so two instances starting together cannot half-apply a schema.

## What the upstream data does and does not contain

Worth knowing before filing a bug:

- **There is no object configuration in MySQL.** No host groups, service
  groups, contacts, host addresses, `notes_url` or parent relationships -
  the worker stores runtime status only. Lists are therefore flat, and
  filtering is by name, state and flags. Naemon's `objects.cache` has all of
  it and can be read later behind an `ObjectProvider`; the repository
  signatures already leave room for a group filter.
- **Log entries need a broker setting.** `statusengine_logentries` stays
  empty unless `LogData` is enabled in `statusengine.toml`:

  ```toml
  LogData = "statusngin_logentries"
  NotificationData = "statusngin_notifications"
  ```

  Both are commented out in the shipped example. Naemon needs a restart
  after the change.
- **Some data does not exist at all.** The worker has no topic for
  comments, flapping history or program status, so there is no comments
  page and no global "core is healthy" tile. Acknowledgement and downtime
  comments are stored and are shown.
- **The history tables are clustered on an object-first primary key,**
  for example `(hostname, service_description, start_time, …)` on
  `statusengine_servicechecks`. InnoDB stores rows in that order, so
  every check for one service sits physically together: reading one
  object over a window is a contiguous range read (8 ms against 321k
  rows here), while reading every object over the same window is not
  (244 ms, growing with the table). The history pages are built around
  that - per-object views are the fast path, and global views lead with
  an object filter and a capped window.
- **History requests always carry a time window,** and the ceiling
  depends on whether they name an object: 90 days for one host or
  service, 6 hours across everything. That follows directly from the
  clustering above, and the interface says which limit is in force
  rather than letting you find out from a rejected request.
- **The standard schema has no partitions.** Some installations add
  them. This project does not depend on them either way; if you
  partition by `time DIV 86400`, note that MySQL does not prune for that
  expression, so the clustering above is still what carries the cost.

## Testing

```bash
make test        # Go, frontend and translations
make test-go
make test-ui
make test-i18n
make lint
```

`make test-i18n` reads both translation files and the code that uses
them: same keys in both languages, every key the code names resolves,
nothing empty, and the keys assembled at runtime - states, error codes,
command actions - complete against what the backend can send.

Accessibility is checked with axe-core against every page in both
themes, plus the four command dialogs, and a keyboard-only walkthrough
that asserts every tab stop is visible and has a focus ring. The colour
tokens are not eyeballed: every text colour is computed against all
three surfaces of its theme and must clear 4.5:1.

The repository tests can also run against a real Statusengine schema.
They are skipped unless a DSN is given:

```bash
make test-integration SEI_TEST_DSN='user:pass@tcp(127.0.0.1:3306)/statusengine'
```

They are worth running after any change to a query: a builder test proves
the clause reads correctly, but only a database proves MySQL accepts it
and that the column list and the scan targets still line up.

The repository tests only read. The one exception is the audit test,
which writes its own rows to `sei_command_audit` under a username
nothing else uses and deletes exactly those again - point `SEI_TEST_DSN`
at a test database, not at production.
