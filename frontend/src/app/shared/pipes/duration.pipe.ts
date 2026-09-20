import { Pipe, type PipeTransform } from '@angular/core';
import { formatDuration } from './time-format';
import { injectTimeContext } from './time-words';

/** How long something has been in its current state: "4d 3h", "12m". */
@Pipe({ name: 'seiDuration' })
export class DurationPipe implements PipeTransform {
  private readonly context = injectTimeContext();

  transform(seconds: number | null | undefined): string {
    return formatDuration(seconds, this.context().words.justNow);
  }
}
