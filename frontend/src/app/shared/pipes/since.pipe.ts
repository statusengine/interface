import { Pipe, type PipeTransform } from '@angular/core';
import { formatSince } from './time-format';
import { injectTimeContext } from './time-words';

/**
 * The elapsed time since a Unix timestamp.
 *
 * A zero timestamp means "never" in the worker's tables, not 1 January
 * 1970. Rendering that date is how a monitoring UI ends up claiming a
 * host was last checked during the Nixon administration.
 */
@Pipe({ name: 'seiSince' })
export class SincePipe implements PipeTransform {
  private readonly context = injectTimeContext();

  transform(unixSeconds: number | null | undefined, now?: number): string {
    return formatSince(unixSeconds, this.context().words, now);
  }
}
