#!/usr/bin/env node
// Checks the translation files against the code that uses them.
//
// A missing key renders as the key itself: not a crash, just a page that
// says `audit.outcome` to somebody trying to work. That is the kind of
// gap a script catches for free and a person catches by accident, three
// releases later, in the language they do not read.
//
// Run from the repository root: node tools/i18n/check.mjs

import { readFileSync, readdirSync, statSync } from "node:fs";
import { join } from "node:path";

const I18N = "frontend/public/i18n";
const SRC = "frontend/src";

function flatten(value, prefix = "") {
  const out = {};
  if (value && typeof value === "object") {
    for (const [key, child] of Object.entries(value)) {
      Object.assign(out, flatten(child, prefix ? `${prefix}.${key}` : key));
    }
  } else if (typeof value === "string") {
    out[prefix] = value;
  }
  return out;
}

const load = (lang) =>
  flatten(JSON.parse(readFileSync(join(I18N, `${lang}.json`), "utf8")));

function sourceFiles(dir) {
  return readdirSync(dir).flatMap((entry) => {
    const path = join(dir, entry);
    if (statSync(path).isDirectory()) {
      return sourceFiles(path);
    }
    return /\.(ts|html)$/.test(path) && !path.endsWith(".spec.ts")
      ? [path]
      : [];
  });
}

const de = load("de");
const en = load("en");
const sources = sourceFiles(SRC).map((path) => ({
  path,
  text: readFileSync(path, "utf8"),
}));

const problems = [];
const report = (what, items) => {
  if (items.length > 0) {
    problems.push(`${what}:\n  ${items.join("\n  ")}`);
  }
};

// 1. The two files describe the same interface.
report(
  "in de.json but not en.json",
  Object.keys(de).filter((k) => !(k in en)),
);
report(
  "in en.json but not de.json",
  Object.keys(en).filter((k) => !(k in de)),
);

// 2. Every key written out in full resolves. The trailing [,)] keeps
//    `t('settings.' + x)` out; those prefixes are checked below.
const staticKeys = new Map();
for (const { path, text } of sources) {
  for (const match of text.matchAll(
    /\bt(?:ranslate)?\(\s*(['"])([^'"]+)\1\s*[,)]/g,
  )) {
    if (match[2].includes(".")) {
      staticKeys.set(match[2], path);
    }
  }
}
report(
  "used by the code but missing",
  [...staticKeys]
    .filter(([key]) => !(key in de) || !(key in en))
    .map(([key, path]) => `${key}  (${path})`),
);

// 3. Keys built at runtime from a value the API produces or a fixed
//    list. The whole key never appears in the source, so the only way to
//    check them is to enumerate what the other side can send.
const families = {
  // internal/domain/state.go
  "states.": [
    "ok",
    "up",
    "warning",
    "critical",
    "down",
    "unknown",
    "unreachable",
    "pending",
  ],
  // internal/domain/history.go NotificationReason
  "reasons.": [
    "normal",
    "acknowledgement",
    "flapping_start",
    "flapping_stop",
    "flapping_disabled",
    "downtime_start",
    "downtime_end",
    "downtime_cancelled",
    "custom",
    "unknown",
  ],
  // internal/httpapi/response.go, plus the client-only network failure
  "errors.": [
    "bad_request",
    "unauthorized",
    "forbidden",
    "not_found",
    "conflict",
    "rate_limited",
    "unavailable",
    "internal_error",
    "invalid_filter",
    "timeout",
    "network_unreachable",
  ],
  // internal/commands/naemon.go Action constants
  "audit.actions.": [
    "acknowledge",
    "remove_acknowledgement",
    "schedule_downtime",
    "delete_downtime",
    "reschedule",
    "submit_result",
    "custom_notification",
    "toggle",
  ],
  // internal/auth/bootstrap.go
  "roles.": ["admin", "operator", "guest"],
  // core/events/live.service.ts LiveMode. Not 'idle': the badge is
  // hidden in that mode rather than labelled.
  "live.": ["live", "polling", "offline"],
  // features/commands/object-actions.ts Mode
  "commands.": [
    "acknowledgeTitle",
    "downtimeTitle",
    "resultTitle",
    "notifyTitle",
  ],
  // features/commands/object-settings.ts SwitchName
  "settings.": [
    "activeChecks",
    "passiveChecks",
    "notifications",
    "flapDetection",
    "eventHandler",
  ],
};
report(
  "built at runtime but missing",
  Object.entries(families).flatMap(([prefix, values]) =>
    values.map((v) => prefix + v).filter((key) => !(key in de) || !(key in en)),
  ),
);

// 4. An empty string renders as nothing at all, which reads as a bug.
report(
  "empty",
  [...Object.entries(de), ...Object.entries(en)]
    .filter(([, value]) => value.trim() === "")
    .map(([key]) => key),
);

if (problems.length > 0) {
  console.error(problems.join("\n\n"));
  process.exit(1);
}
console.log(
  `translations: ${Object.keys(de).length} keys in both languages, ` +
    `${staticKeys.size} referenced by name, all resolve`,
);
