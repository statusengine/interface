import { describe, expect, it } from 'vitest';
import { DurationPipe } from './duration.pipe';
import { SincePipe } from './since.pipe';

describe('DurationPipe', () => {
  const pipe = new DurationPipe();

  it('never shows more than two units', () => {
    // "4d 3h 17m 9s" is precise and unreadable, and nobody acting on a
    // five-day outage cares about the seconds.
    expect(pipe.transform(4 * 86400 + 3 * 3600 + 17 * 60 + 9)).toBe('4d 3h');
    expect(pipe.transform(3 * 3600 + 17 * 60 + 9)).toBe('3h 17m');
    expect(pipe.transform(17 * 60 + 9)).toBe('17m 9s');
  });

  it('drops a zero second unit', () => {
    expect(pipe.transform(4 * 86400)).toBe('4d');
    expect(pipe.transform(3 * 3600)).toBe('3h');
    expect(pipe.transform(5 * 60)).toBe('5m');
  });

  it('collapses the first few seconds', () => {
    expect(pipe.transform(0)).toBe('just now');
    expect(pipe.transform(4)).toBe('just now');
    expect(pipe.transform(5)).toBe('5s');
  });

  it('renders nothing for a missing value', () => {
    expect(pipe.transform(null)).toBe('');
    expect(pipe.transform(undefined)).toBe('');
    expect(pipe.transform(Number.NaN)).toBe('');
  });

  it('treats a negative duration as zero rather than counting backwards', () => {
    expect(pipe.transform(-100)).toBe('just now');
  });
});

describe('SincePipe', () => {
  const pipe = new SincePipe();
  const now = 1_700_000_000;

  // The worker writes 0 for "this has not happened". Rendering that as a
  // date is how a monitoring UI ends up claiming a host was last checked
  // in 1970.
  it('says never for a zero timestamp', () => {
    expect(pipe.transform(0, now)).toBe('never');
    expect(pipe.transform(null, now)).toBe('never');
    expect(pipe.transform(undefined, now)).toBe('never');
  });

  it('measures elapsed time', () => {
    expect(pipe.transform(now - 90, now)).toBe('1m 30s');
    expect(pipe.transform(now - 7200, now)).toBe('2h');
  });

  // next_check is normally in the future, and "in 3m" is the right
  // reading of that, not "-3m".
  it('reads a future timestamp forwards', () => {
    expect(pipe.transform(now + 180, now)).toBe('in 3m');
  });
});
