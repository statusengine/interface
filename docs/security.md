# Security

This is the review that was done before putting the interface on the
open internet with a public demo account that may submit commands, and
the checklist for doing that. It says what was tested, what it found,
what changed, and what is left for the deployment to get right.

Dated against the commit that introduced the demo command allowlist.

---

## 1. What is being protected, and from whom

The attacker to design against is an anonymous visitor with a working
demo account: they can reach every read endpoint and, if the deployment
allows it, submit commands. They are not a browser bug and not a
compromised operator; they are simply a stranger with time.

What they could want, in the order that would hurt:

| Target | What it would look like |
|---|---|
| **The monitoring core** | Commands at HTTP speed: forced checks, downtimes, silenced objects. The interface can generate load on Naemon and on everything it monitors. |
| **The people on call** | A custom notification mails and pages the real contacts of whatever is monitored. One request, real phones. |
| **The database** | Rows written without limit - passive results, audit entries, sessions - until the disk fills. |
| **This process** | Connections, goroutines and file descriptors held open; expensive queries repeated. |
| **The data** | Other people's monitoring data, the worker's API keys, the audit trail's names and addresses. |
| **The host** | Remote code execution, file reads, SQL injection. |

## 2. What was tested

Against a running instance, with a session and without:

- Every endpoint unauthenticated (401 except `/healthz`, `/readyz`,
  `/meta` and the login routes, which is intended).
- Path traversal against the static handler: `/../seid.yaml`,
  `/%2e%2e/%2e%2e/etc/passwd`, `/i18n/../../seid.yaml`, encoded and
  unencoded. All either 404 or the SPA's `index.html`; no file outside
  the bundle is reachable.
- SQL injection through every parameter shape a list takes: `sort`
  (whitelist, refused with the allowed list), `state` (parsed as a
  number), `q` (a `LIKE` pattern, escaped), `from`/`to`, `limit`,
  `offset`, `hours`.
- Command injection: `;`, newlines and control characters in comments,
  authors, hostnames and service descriptions - refused with a message
  saying why.
- Oversized input: a 3 MB body (400), `limit=100000` (400), 1001
  targets in one command (400).
- Login brute force: refused after ten attempts a minute from one
  address.
- Passwordless demo login, repeated.
- Command flooding: twenty-five submissions in 232 ms.
- 150 concurrent event-stream connections.
- Response headers, cookie flags, error bodies, `/meta` contents.
- The frontend for unescaped HTML (`innerHTML`, `bypassSecurityTrust*`):
  none, anywhere.

## 3. What it found, and what changed

### 3.1 Commands had no rate limit at all — fixed

Twenty-five command submissions in under a quarter of a second were all
accepted. Each one becomes a message on the broker and work for the
monitoring core; a visitor with a loop could have driven Naemon and
everything it checks.

Now: a per-caller budget (`command_rate_limit`, 60/minute by default)
and a smaller one for the demo account (`demo_command_rate_limit`,
10/minute), keyed by account **and** address so one visitor cannot spend
another's. Above that, a ceiling across all demo traffic at once, six
times the per-visitor budget, so a crowd or a browser farm cannot sum up
to the same flood. Refusals answer `429` with `Retry-After`.

### 3.2 A demo account could only be read-only or fully privileged — fixed

The demo account was on the `guest` role, which refuses every command,
and the passwordless login refused any other role outright. Granting
commands meant granting all of them, including custom notifications.

Now: `demo_commands` is an allowlist by name. It builds a `demo` role -
read everything, plus exactly those commands - which is rebuilt from the
configuration on every start, so the way to change what a stranger may
do is the config file and nothing else. Three layers hold it:

1. the config refuses `notify` by name and rejects any unknown action;
2. no mapping from `notify` to a permission exists, so nothing can grant
   it to the demo role even if the first layer were bypassed;
3. the passwordless login re-checks the account's permissions against
   the allowlist and refuses to open a session if the database says
   something wider.

`notify` is the one command a demo must never have: it sends to the real
contacts of the monitored objects.

### 3.3 The passwordless login had no rate limit — fixed

Twelve calls, twelve sessions, no password involved: unbounded row
growth in `sei_sessions` at no cost to the caller. It is now throttled
per address like any other login.

### 3.4 The event stream had no ceiling — fixed

150 connections cost about 1.6 MB and seven threads, so memory is not
the limit - file descriptors are, and nothing stopped a client from
opening them until the process ran out. `max_event_clients` (500)
refuses beyond that with `503` **before** any header is written, so the
client falls back to polling instead of holding a stream that never
carries anything.

### 3.5 A bulk of a thousand objects from a stranger — fixed

The 1000-command ceiling is right for an operator who selected them.
`demo_max_targets` (25) applies to the demo account instead.

### 3.6 Command fields had no length limit — fixed

Naemon reads a command line into a fixed buffer and truncates silently
past it, so a long comment was lost work. It was also unlimited graffiti
in front of every other visitor. Fields are capped at 255 characters.

### 3.7 No HSTS — fixed

Sent as `max-age=15552000` when `secure_cookies` is on, which is the
deployment saying the browser reaches it over TLS.

### 3.8 The audit grew forever and recorded visitors' addresses — fixed

`sei_command_audit` had no retention, and every command stored the
caller's IP. On an internal deployment that is the point; on a public
demo it is strangers' personal data kept indefinitely.

`audit_retention_days` (0 = keep, which stays the default) prunes older
rows every six hours. `audit_client_ip` (default true) can be switched
off to keep the trail without the addresses.

## 4. What was already sound

Verified rather than assumed:

- **SQL.** Every caller-supplied value is a placeholder. Three things
  are ever spliced into SQL text: column expressions from a whitelist
  map, the `?,?,?` run for an `IN` list built from a count, and two
  table names that are compile-time literals. `LIKE` patterns are
  escaped. See `docs/code-tour.md` §6.2 for the greps.
- **Naemon command construction.** The author always comes from the
  session, never the request. Every field is checked for `;`, newlines
  and control characters before assembly.
- **Passwords and sessions.** argon2id (RFC 9106 second option), PHC
  format, constant-time comparison, and the same work spent on an
  unknown username. The session token is 32 bytes from `crypto/rand`;
  only its SHA-256 is stored. Cookie: `HttpOnly`, `SameSite=Lax`,
  `Secure` when configured. No token in `localStorage`.
- **Authorization.** Every route declares its permission where it is
  registered, in one table, enforced by middleware. The UI hides what a
  role cannot use; the server refuses it regardless.
- **Static files.** No path outside the embedded bundle is reachable.
- **Headers.** `default-src 'self'`, `script-src 'self'` with no
  `unsafe-inline`, `frame-ancestors 'none'`, `nosniff`,
  `Referrer-Policy: same-origin`. No CORS headers at all, so the API is
  same-origin only.
- **XSS.** Angular escapes every interpolation and the codebase contains
  no `innerHTML` and no `bypassSecurityTrust*`. Plugin output, command
  responses and comments are all rendered as text.
- **Error messages.** A code and a sentence; the driver error and the
  query go to the log, never to the client.
- **Debug surface.** No `pprof`, no `expvar`, no metrics endpoint.
- **Proxy headers.** `X-Forwarded-For` is honoured only when the direct
  peer is loopback or a private address, so a header from the open
  internet cannot forge the address a rate limit is keyed on.
- **Worker credentials.** Never leave the process: the browser talks
  only to this server, and `/meta` carries no secret.

## 5. Accepted risks

These are consequences of what a demo is, not defects. They are listed
so the decision is deliberate.

- **A demo visitor can make the monitoring quieter.** Acknowledging,
  scheduling downtime or switching notifications off are real changes,
  visible to everyone and reversible by anyone. Point the demo at hosts
  nobody depends on.
- **Comments are public writing.** Anything a visitor types into an
  acknowledgement is shown to every other visitor until it is removed.
  Capped at 255 characters; not moderated.
- **`submit-result` puts arbitrary text into the monitoring data** if it
  is allowlisted. It is the most interesting command for a demo and the
  one most worth thinking twice about.
- **The demo shares one account.** The audit shows which commands came
  from it and from which address (unless that is switched off), but not
  which visitor.

## 6. Checklist for a public deployment

**Before it is reachable:**

- [ ] TLS in front, and `secure_cookies: true` so the cookie is
      `Secure` and HSTS is sent.
- [ ] `listen_addr` on loopback or a private interface, with the
      reverse proxy as the only thing in front of it. The default is
      `127.0.0.1` on purpose.
- [ ] A dedicated MySQL account: `SELECT` on `statusengine_*`, full
      rights on `sei_*` only. The interface never writes to the
      worker's tables, so the grant should not allow it to.
- [ ] The worker's command and event ports (8081, 8080) reachable from
      this process and from nothing else. They authenticate with a
      static key.
- [ ] A real administrator account created with `seid user create`.
      There is no default administrator and no default password.

**For the demo itself:**

- [ ] `demo_mode: true`, `demo_user: guest`.
- [ ] `demo_commands` naming exactly what a stranger may do. Start with
      `[acknowledge, downtime, reschedule]`; add `toggle` if the demo
      is about configuration, `submit-result` only if you have read
      §5.
- [ ] `demo_command_rate_limit` (10/minute per visitor is the default)
      and `demo_max_targets` (25).
- [ ] `audit_retention_days: 30` and, if visitors' addresses should not
      be kept, `audit_client_ip: false`.
- [ ] A privacy notice if addresses are kept: they are personal data.
- [ ] Point it at monitored objects that can be acknowledged, silenced
      and put in downtime by strangers without anybody's afternoon
      being ruined.

**Around it:**

- [ ] Rate limiting and request size limits in the reverse proxy as
      well. The limits here protect the monitoring core; the proxy
      protects the process.
- [ ] Log rotation, and an eye on `sei_sessions` and
      `sei_command_audit` growth for the first week.
- [ ] Keep the proxy patched. TLS and HTTP/2 are terminated there, so
      its bugs are the deployment's bugs.

**Verify after the first deploy:**

```bash
curl -sI https://<host>/ | grep -i strict-transport   # HSTS present
curl -si https://<host>/api/v1/auth/login/demo | grep -i set-cookie   # Secure, HttpOnly
curl -s https://<host>/api/v1/hosts -o /dev/null -w '%{http_code}\n'  # 401
```

Then sign in as the demo account and confirm the login page lists
exactly the commands you allowed - it is generated from the same
configuration the server enforces, so a difference there means the
config is not what you think it is.

## 7. Reporting something

Security reports belong in an email to the maintainers rather than a
public issue. A report that names the request, the response and what
was expected is one that can be fixed the same day.
