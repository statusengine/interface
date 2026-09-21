/**
 * Formatting for the metric chart.
 *
 * Separate from the component because these are pure decisions worth
 * testing on their own - and because importing the chart pulls in uPlot,
 * which reaches for matchMedia the moment it loads.
 */

/** A short human duration for the resolution label. */
export function formatDuration(seconds: number): string {
  if (seconds % 86400 === 0) return `${seconds / 86400}d`;
  if (seconds % 3600 === 0) return `${seconds / 3600}h`;
  if (seconds % 60 === 0) return `${seconds / 60}m`;
  return `${seconds}s`;
}

/**
 * Values with their unit, at a precision that matches their magnitude.
 * Three decimals on a disk in megabytes is noise; none on a load average
 * loses the reading.
 */
export function formatValue(value: number, unit: string): string {
  if (!Number.isFinite(value)) {
    return '—';
  }
  const abs = Math.abs(value);
  const decimals = abs >= 100 ? 0 : abs >= 1 ? 2 : 3;
  const text = value.toFixed(decimals).replace(/\.?0+$/, '');
  return unit ? `${text} ${unit}` : text;
}

export function formatAxisValue(value: number): string {
  const abs = Math.abs(value);
  if (abs >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`;
  if (abs >= 1_000) return `${(value / 1_000).toFixed(1)}k`;
  if (abs >= 100) return value.toFixed(0);
  if (abs >= 1) return value.toFixed(1);
  return value.toFixed(2);
}

/**
 * A translucent version of a resolved colour, for the min-max band.
 * The tokens resolve to hex, so this handles #rgb and #rrggbb and falls
 * back to the colour itself rather than drawing nothing.
 */
export function withAlpha(colour: string, alpha: number): string {
  const hex = colour.trim();
  const match = /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.exec(hex);
  if (!match) {
    return hex;
  }
  const full =
    match[1].length === 3
      ? match[1]
          .split('')
          .map((c) => c + c)
          .join('')
      : match[1];
  const r = parseInt(full.slice(0, 2), 16);
  const g = parseInt(full.slice(2, 4), 16);
  const b = parseInt(full.slice(4, 6), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

/** Whether a surface colour is a dark one. Rough on purpose: this
 *  decides an alpha, not a contrast ratio. */
export function isDarkSurface(surface: string): boolean {
  const match = /^#([0-9a-f]{6})$/i.exec(surface.trim());
  if (!match) {
    return false;
  }
  const n = parseInt(match[1], 16);
  const r = (n >> 16) & 255;
  const g = (n >> 8) & 255;
  const b = n & 255;
  return (r * 299 + g * 587 + b * 114) / 1000 < 128;
}

/**
 * How strongly to shade under a line, if at all.
 *
 * One series can carry a visible wash; several cannot, because where
 * they overlap the alphas add up and the chart turns to soup. Past
 * three, the lines have to speak for themselves.
 *
 * The numbers are higher on a dark surface: the same alpha that reads
 * as a tint on white is a rumour at night, because a translucent colour
 * carries much further over paper than over ink.
 *
 * Top and floor are close together on purpose. This is a light fill
 * with a slight fade, not a gradient - a wash that drops away to
 * nothing puts all its colour in the few pixels under the highest peak
 * and leaves the rest of the line looking unfilled.
 */
export function washFor(count: number, dark: boolean): { top: number; floor: number } | null {
  if (count === 1) {
    return dark ? { top: 0.34, floor: 0.16 } : { top: 0.24, floor: 0.1 };
  }
  if (count <= 3) {
    return dark ? { top: 0.16, floor: 0.07 } : { top: 0.11, floor: 0.05 };
  }
  return null;
}
