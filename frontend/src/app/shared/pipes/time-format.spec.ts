import { describe, expect, it } from 'vitest';
import {
  formatDuration,
  formatHour,
  formatSince,
  formatTimestamp,
  type TimeWords,
} from './time-format';

const en: TimeWords = {
  never: 'never',
  justNow: 'just now',
  ahead: (d) => `in ${d}`,
};

const de: TimeWords = {
  never: 'nie',
  justNow: 'gerade eben',
  ahead: (d) => `in ${d}`,
};

describe('formatDuration', () => {
  it('never shows more than two units', () => {
    // "4d 3h 17m 9s" is precise and unreadable, and nobody acting on a
    // five-day outage cares about the seconds.
    expect(formatDuration(4 * 86400 + 3 * 3600 + 17 * 60 + 9, en.justNow)).toBe('4d 3h');
    expect(formatDuration(3 * 3600 + 17 * 60 + 9, en.justNow)).toBe('3h 17m');
    expect(formatDuration(17 * 60 + 9, en.justNow)).toBe('17m 9s');
  });

  it('drops a zero second unit', () => {
    expect(formatDuration(4 * 86400, en.justNow)).toBe('4d');
    expect(formatDuration(3 * 3600, en.justNow)).toBe('3h');
    expect(formatDuration(5 * 60, en.justNow)).toBe('5m');
  });

  it('collapses the first few seconds, in the reader s language', () => {
    expect(formatDuration(0, en.justNow)).toBe('just now');
    expect(formatDuration(4, de.justNow)).toBe('gerade eben');
    expect(formatDuration(5, en.justNow)).toBe('5s');
  });

  it('renders nothing for a missing value', () => {
    expect(formatDuration(null, en.justNow)).toBe('');
    expect(formatDuration(undefined, en.justNow)).toBe('');
    expect(formatDuration(Number.NaN, en.justNow)).toBe('');
  });

  it('treats a negative duration as zero rather than counting backwards', () => {
    expect(formatDuration(-100, en.justNow)).toBe('just now');
  });
});

describe('formatSince', () => {
  const now = 1_700_000_000;

  // The worker writes 0 for "this has not happened". Rendering that as a
  // date is how a monitoring UI ends up claiming a host was last checked
  // in 1970.
  it('says never for a zero timestamp, in the reader s language', () => {
    expect(formatSince(0, en, now)).toBe('never');
    expect(formatSince(null, de, now)).toBe('nie');
    expect(formatSince(undefined, de, now)).toBe('nie');
  });

  it('measures elapsed time', () => {
    expect(formatSince(now - 90, en, now)).toBe('1m 30s');
    expect(formatSince(now - 7200, en, now)).toBe('2h');
  });

  // next_check is normally in the future, and "in 3m" is the right
  // reading of that, not "-3m".
  it('reads a future timestamp forwards', () => {
    expect(formatSince(now + 180, en, now)).toBe('in 3m');
  });
});

describe('formatTimestamp', () => {
  const noon = Date.UTC(2026, 7, 1, 12, 0, 0) / 1000;

  it('writes the date the way the chosen language writes dates', () => {
    const asGerman = formatTimestamp(noon, 'de', 'full', de.never);
    const asEnglish = formatTimestamp(noon, 'en-US', 'full', en.never);

    // 08/01/2026 in a German interface is not a date, it is a riddle
    // with two answers.
    expect(asGerman).toContain('01.08.2026');
    expect(asEnglish).toContain('08/01/2026');
  });

  it('says never rather than 1970', () => {
    expect(formatTimestamp(0, 'de', 'full', de.never)).toBe('nie');
    expect(formatTimestamp(null, 'en', 'full', en.never)).toBe('never');
  });

  it('shortens to a time of day when asked', () => {
    const short = formatTimestamp(noon, 'de', 'short', de.never);
    expect(short).not.toContain('2026');
    expect(short).toMatch(/\d{2}:\d{2}:\d{2}/);
  });
});

describe('formatHour', () => {
  const noon = Date.UTC(2026, 7, 2, 12, 0, 0) / 1000;

  // The label has to identify one bar and no other. Over a day two bars
  // read "19:00"; over a week, two read "Sun 19:00".
  it('adds as much of the date as the window needs', () => {
    expect(formatHour(noon, 'de', 6)).toMatch(/^\d{2}:\d{2}$/);
    expect(formatHour(noon, 'de', 25)).toMatch(/^So\.?,? \d{2}:\d{2}$/);
    expect(formatHour(noon, 'de', 169)).toMatch(/^02\.08\.,? \d{2}:\d{2}$/);
  });
});
