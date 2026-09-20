/**
 * How this app writes times, with the wording passed in.
 *
 * Pure functions rather than methods on the pipes: the pipes now need
 * the active language, which means dependency injection, and a
 * formatting decision should still be testable without a TestBed. It is
 * also the only way `SincePipe` can reuse the duration logic - it used
 * to do it by constructing the other pipe by hand, which stops working
 * the moment a pipe has a dependency.
 */

/** The words a duration needs. Everything else is numbers and units. */
export interface TimeWords {
  /** Nothing has happened yet; the worker writes 0 for this. */
  never: string;
  /** Under five seconds, where an exact count is noise. */
  justNow: string;
  /** A time still ahead, given the duration: "in 3m". */
  ahead: (duration: string) => string;
}

/**
 * A span of seconds as an operator says it: "4d 3h", "12m", "just now".
 *
 * Two units, never three. "4d 3h 17m 9s" is precise and unreadable, and
 * nobody acting on a five-day outage cares about the seconds. The unit
 * letters stay as they are in both languages: they are notation, like a
 * unit on an axis, and "4T 3Std" is not what anyone writes in a ticket.
 */
export function formatDuration(seconds: number | null | undefined, justNow: string): string {
  if (seconds === null || seconds === undefined || !Number.isFinite(seconds)) {
    return '';
  }
  const total = Math.max(0, Math.floor(seconds));
  if (total < 5) {
    return justNow;
  }

  const d = Math.floor(total / 86400);
  const h = Math.floor((total % 86400) / 3600);
  const m = Math.floor((total % 3600) / 60);
  const s = total % 60;

  if (d > 0) return h > 0 ? `${d}d ${h}h` : `${d}d`;
  if (h > 0) return m > 0 ? `${h}h ${m}m` : `${h}h`;
  if (m > 0) return s > 0 ? `${m}m ${s}s` : `${m}m`;
  return `${s}s`;
}

/**
 * The time since a Unix timestamp, or the time until it when it is
 * ahead - `next_check` is normally in the future, and "in 3m" is the
 * right reading of that, not "-3m".
 */
export function formatSince(
  unixSeconds: number | null | undefined,
  words: TimeWords,
  now?: number,
): string {
  if (!unixSeconds) {
    return words.never;
  }
  const reference = now ?? Math.floor(Date.now() / 1000);
  const elapsed = reference - unixSeconds;
  if (elapsed < 0) {
    return words.ahead(formatDuration(-elapsed, words.justNow));
  }
  return formatDuration(elapsed, words.justNow);
}

/**
 * A Unix timestamp as a local date and time, in the reader's language.
 *
 * The locale is the language they chose, not the browser's default.
 * `08/01/2026` in a German interface is not a date, it is a riddle with
 * two answers.
 */
export function formatTimestamp(
  unixSeconds: number | null | undefined,
  locale: string,
  style: 'full' | 'short',
  never: string,
): string {
  if (!unixSeconds) {
    return never;
  }
  const date = new Date(unixSeconds * 1000);
  if (style === 'short') {
    return date.toLocaleTimeString(locale, {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  }
  return date.toLocaleString(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

/**
 * One hour bucket as a label, told how wide the window around it is.
 *
 * The span decides how much of the date has to be there for the label
 * to identify one bar and no other. Over a day, two bars read "19:00";
 * over a week, two read "Sun 19:00". Both make a complete axis look
 * broken.
 */
export function formatHour(t: number, locale: string, spanHours: number): string {
  const date = new Date(t * 1000);
  if (spanHours >= 6 * 24) {
    return date.toLocaleString(locale, {
      day: '2-digit',
      month: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  }
  if (spanHours >= 24) {
    return date.toLocaleString(locale, { weekday: 'short', hour: '2-digit', minute: '2-digit' });
  }
  return date.toLocaleTimeString(locale, { hour: '2-digit', minute: '2-digit' });
}
