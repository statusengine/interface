import { describe, expect, it } from 'vitest';
import {
  axisRange,
  formatAxisValue,
  formatDuration,
  formatValue,
  isDarkSurface,
  washFor,
  withAlpha,
} from './metric-format';

describe('formatDuration', () => {
  it('uses the largest whole unit', () => {
    expect(formatDuration(30)).toBe('30s');
    expect(formatDuration(300)).toBe('5m');
    expect(formatDuration(3600)).toBe('1h');
    expect(formatDuration(86400)).toBe('1d');
    expect(formatDuration(7 * 86400)).toBe('7d');
  });

  it('falls back to seconds when nothing divides evenly', () => {
    expect(formatDuration(90)).toBe('90s');
  });
});

describe('formatValue', () => {
  // Three decimals on a disk in megabytes is noise; none on a load
  // average loses the reading.
  it('scales precision to magnitude', () => {
    expect(formatValue(47643, 'MB')).toBe('47643 MB');
    expect(formatValue(1.5, '')).toBe('1.5');
    expect(formatValue(0.142, '')).toBe('0.142');
  });

  it('trims trailing zeros rather than padding', () => {
    expect(formatValue(2, '')).toBe('2');
    expect(formatValue(2.5, '')).toBe('2.5');
  });

  it('appends the unit only when there is one', () => {
    expect(formatValue(5, '%')).toBe('5 %');
    expect(formatValue(5, '')).toBe('5');
  });

  it('does not render a non-finite value as a number', () => {
    expect(formatValue(Number.NaN, 'MB')).toBe('—');
    expect(formatValue(Number.POSITIVE_INFINITY, '')).toBe('—');
  });
});

describe('formatAxisValue', () => {
  it('abbreviates large magnitudes so ticks stay narrow', () => {
    expect(formatAxisValue(2_500_000)).toBe('2.5M');
    expect(formatAxisValue(4096)).toBe('4.1k');
    expect(formatAxisValue(250)).toBe('250');
    expect(formatAxisValue(2.5)).toBe('2.5');
    expect(formatAxisValue(0.25)).toBe('0.25');
  });

  it('keeps negatives readable', () => {
    expect(formatAxisValue(-4096)).toBe('-4.1k');
  });
});

describe('withAlpha', () => {
  it('turns a resolved token into a translucent fill', () => {
    expect(withAlpha('#2a78d6', 0.16)).toBe('rgba(42, 120, 214, 0.16)');
  });

  it('accepts the short form and surrounding whitespace', () => {
    expect(withAlpha('  #abc ', 0.5)).toBe('rgba(170, 187, 204, 0.5)');
  });

  // A token that resolves to something else must not silently make the
  // band invisible.
  it('passes anything it cannot parse straight through', () => {
    expect(withAlpha('rebeccapurple', 0.2)).toBe('rebeccapurple');
    expect(withAlpha('', 0.2)).toBe('');
  });
});

describe('washFor', () => {
  // One series can carry a visible wash; several cannot, because where
  // they overlap the alphas add up.
  it('fades out as the series multiply, and gives up past three', () => {
    const one = washFor(1, false)!;
    const three = washFor(3, false)!;
    expect(one.top).toBeGreaterThan(three.top);
    expect(washFor(4, false)).toBeNull();
    expect(washFor(8, true)).toBeNull();
  });

  // The same alpha that reads as a tint on white is a rumour at night.
  it('shades harder on a dark surface', () => {
    expect(washFor(1, true)!.top).toBeGreaterThan(washFor(1, false)!.top);
    expect(washFor(2, true)!.floor).toBeGreaterThan(washFor(2, false)!.floor);
  });

  // A wash that drops away to nothing puts all its colour under the
  // highest peak and leaves the rest of the line looking unfilled.
  it('keeps a floor under the fade', () => {
    for (const count of [1, 2, 3]) {
      for (const dark of [true, false]) {
        const wash = washFor(count, dark)!;
        expect(wash.floor).toBeGreaterThan(0);
        expect(wash.floor).toBeLessThan(wash.top);
        expect(wash.top).toBeLessThan(0.5); // a fill, not a block of colour
      }
    }
  });
});

describe('isDarkSurface', () => {
  it('reads the theme off the surface colour', () => {
    expect(isDarkSurface('#11161f')).toBe(true);
    expect(isDarkSurface('#ffffff')).toBe(false);
    expect(isDarkSurface('#f7f8fa')).toBe(false);
  });

  // The token can be missing or in a notation this does not parse; a
  // light surface is the safer guess, because the lighter wash is the
  // one that cannot drown a line.
  it('assumes light when it cannot tell', () => {
    expect(isDarkSurface('')).toBe(false);
    expect(isDarkSurface('oklch(0.2 0.02 250)')).toBe(false);
  });
});

describe('axisRange', () => {
  // A chart that fits itself to the data answers a different question
  // than the one that was asked.
  it('draws the window that was requested', () => {
    expect(axisRange(1000, 5000, 3000, 4000, 60)).toEqual([1000, 5000]);
    // Even when the data is only a sliver of it: the gap is the point.
    expect(axisRange(1000, 5000, 4900, 4950, 60)).toEqual([1000, 5000]);
  });

  // A zero-width range is where uPlot invents one of its own, and one
  // sample renders as an axis running from 2026 to 2029.
  it('widens a single point into something drawable', () => {
    expect(axisRange(0, 0, 1790095680, 1790095680, 60)).toEqual([1790095620, 1790095740]);
    // No bucket size to go by: a minute either side.
    expect(axisRange(0, 0, 100, 100, 0)).toEqual([40, 160]);
  });

  it('falls back to the data when there is no window', () => {
    expect(axisRange(0, 0, 100, 200, 60)).toEqual([100, 200]);
    // A window that makes no sense is not a window.
    expect(axisRange(500, 100, 100, 200, 60)).toEqual([100, 200]);
  });
});
