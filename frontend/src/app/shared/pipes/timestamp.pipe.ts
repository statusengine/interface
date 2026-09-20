import { Pipe, type PipeTransform } from '@angular/core';
import { formatTimestamp } from './time-format';
import { injectTimeContext } from './time-words';

/**
 * A Unix timestamp as a local date and time, in the chosen language.
 * Zero renders as "never", because that is what the worker writes when
 * something has not happened.
 */
@Pipe({ name: 'seiTimestamp' })
export class TimestampPipe implements PipeTransform {
  private readonly context = injectTimeContext();

  transform(unixSeconds: number | null | undefined, style: 'full' | 'short' = 'full'): string {
    const { locale, words } = this.context();
    return formatTimestamp(unixSeconds, locale, style, words.never);
  }
}
