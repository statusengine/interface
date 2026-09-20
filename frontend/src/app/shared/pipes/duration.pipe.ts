import { Pipe, type PipeTransform } from '@angular/core';

/**
 * How long something has been in its current state, as an operator says
 * it: "4d 3h", "12m", "just now".
 *
 * Two units, never three. "4d 3h 17m 9s" is precise and unreadable, and
 * nobody acting on a five-day outage cares about the seconds.
 */
@Pipe({ name: 'seiDuration' })
export class DurationPipe implements PipeTransform {
  transform(seconds: number | null | undefined): string {
    if (seconds === null || seconds === undefined || !Number.isFinite(seconds)) {
      return '';
    }
    const total = Math.max(0, Math.floor(seconds));
    if (total < 5) {
      return 'just now';
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
}
