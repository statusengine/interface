import { ChangeDetectionStrategy, Component, computed, input, output } from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';
import { Icon } from './icon';

/**
 * Offset paging. It shows the range and the total rather than page
 * numbers, because "1-50 of 3,418" answers the question people actually
 * have, which is how much is left.
 */
@Component({
  selector: 'sei-pagination',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective, Icon],
  template: `
    @if (total() > 0) {
      <div
        class="flex flex-wrap items-center justify-between gap-3 border-t border-line bg-surface px-4 py-2.5 sm:px-6"
        *transloco="let t"
      >
        <p class="text-[12px] text-ink-dim">
          {{ t('list.range', { from: from(), to: to(), total: total() }) }}
        </p>

        <div class="flex items-center gap-1">
          <label class="mr-2 flex items-center gap-1.5 text-[12px] text-ink-dim">
            {{ t('list.perPage') }}
            <select
              (change)="limitChange.emit(+$any($event.target).value)"
              class="rounded-xs border border-line bg-ground px-1.5 py-1 text-[12px] text-ink"
            >
              @for (size of pageSizes; track size) {
                <option [value]="size" [selected]="size === limit()">{{ size }}</option>
              }
            </select>
          </label>

          <button
            type="button"
            (click)="offsetChange.emit(offset() - limit())"
            [disabled]="offset() === 0"
            class="rounded-sm border border-line p-1.5 text-ink-dim transition-colors hover:border-line-strong hover:text-ink disabled:cursor-not-allowed disabled:opacity-40"
            [attr.aria-label]="t('list.previous')"
          >
            <sei-icon name="chevron-right" [size]="14" class="rotate-180" />
          </button>
          <button
            type="button"
            (click)="offsetChange.emit(offset() + limit())"
            [disabled]="to() >= total()"
            class="rounded-sm border border-line p-1.5 text-ink-dim transition-colors hover:border-line-strong hover:text-ink disabled:cursor-not-allowed disabled:opacity-40"
            [attr.aria-label]="t('list.next')"
          >
            <sei-icon name="chevron-right" [size]="14" />
          </button>
        </div>
      </div>
    }
  `,
})
export class Pagination {
  readonly total = input(0);
  readonly offset = input(0);
  readonly limit = input(50);

  readonly offsetChange = output<number>();
  readonly limitChange = output<number>();

  readonly pageSizes = [25, 50, 100, 250, 500];

  readonly from = computed(() => (this.total() === 0 ? 0 : this.offset() + 1));
  readonly to = computed(() => Math.min(this.offset() + this.limit(), this.total()));
}
