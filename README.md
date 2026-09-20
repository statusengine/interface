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
| 3 | History pages, performance charts, metrics provider abstraction | next |
| 4 | External commands, live updates with polling fallback | planned |
| 5 | Accessibility pass, edge cases, hardening | planned |

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
login page as one click. It signs in through the ordinary session path, and
the server refuses every command it submits - the UI hides those controls
as a courtesy, but the refusal is what enforces it.

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
- **The standard schema has no partitions.** Some installations add
  them. This project does not depend on them either way; if you
  partition by `time DIV 86400`, note that MySQL does not prune for that
  expression, so the clustering above is still what carries the cost.

## Testing

```bash
make test        # Go and frontend
make test-go
make test-ui
make lint
```

The repository tests can also run against a real Statusengine schema.
They only read, and they are skipped unless a DSN is given:

```bash
make test-integration SEI_TEST_DSN='user:pass@tcp(127.0.0.1:3306)/statusengine'
```

They are worth running after any change to a query: a builder test proves
the clause reads correctly, but only a database proves MySQL accepts it
and that the column list and the scan targets still line up.
