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
