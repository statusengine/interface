import { Pipe, type PipeTransform } from '@angular/core';
import { DurationPipe } from './duration.pipe';

const duration = new DurationPipe();

/**
 * The elapsed time since a Unix timestamp.
 *
 * A zero timestamp means "never" in the worker's tables, not 1 January
 * 1970. Rendering that date is how a monitoring UI ends up claiming a
 * host was last checked during the Nixon administration.
 */
@Pipe({ name: 'seiSince' })
export class SincePipe implements PipeTransform {
  transform(unixSeconds: number | null | undefined, now?: number): string {
    if (!unixSeconds) {
      return 'never';
    }
    const reference = now ?? Math.floor(Date.now() / 1000);
    const elapsed = reference - unixSeconds;
    if (elapsed < 0) {
      // A next_check in the future is the normal case for that column.
      return 'in ' + duration.transform(-elapsed);
    }
    return duration.transform(elapsed);
  }
}
