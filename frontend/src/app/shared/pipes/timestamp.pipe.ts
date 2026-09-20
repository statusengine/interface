import { Pipe, type PipeTransform, inject, LOCALE_ID } from '@angular/core';

/**
 * A Unix timestamp as a local date and time. Zero renders as "never",
 * because that is what the worker writes when something has not happened.
 */
@Pipe({ name: 'seiTimestamp' })
export class TimestampPipe implements PipeTransform {
  private readonly locale = inject(LOCALE_ID);

  transform(unixSeconds: number | null | undefined, style: 'full' | 'short' = 'full'): string {
    if (!unixSeconds) {
      return 'never';
    }
    const date = new Date(unixSeconds * 1000);
    if (style === 'short') {
      return date.toLocaleTimeString(this.locale, {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
    }
    return date.toLocaleString(this.locale, {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  }
}
