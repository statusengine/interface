import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';

/**
 * The filter strip above a list. It wraps on narrow screens rather than
 * hiding controls behind a menu: the filters are how an operator narrows
 * an incident down, and burying them costs more than the space they take.
 */
@Component({
  selector: 'sei-toolbar',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective],
  template: `
    <div
      class="flex flex-wrap items-center gap-2 border-b border-line bg-surface px-4 py-2.5 sm:px-6"
      *transloco="let t"
    >
      <ng-content />
      @if (showClear()) {
        <button
          type="button"
          (click)="clear.emit()"
          class="ml-auto whitespace-nowrap rounded-sm px-2 py-1 text-[12px] text-ink-dim underline-offset-2 transition-colors hover:text-ink hover:underline"
        >
          {{ t('list.clearFilters') }}
        </button>
      }
    </div>
  `,
})
export class Toolbar {
  readonly showClear = input(false);
  readonly clear = output<void>();
}
