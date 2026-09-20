import { ChangeDetectionStrategy, Component, computed, input, output } from '@angular/core';
import { Icon } from './icon';

/**
 * A sortable column header. Used as an attribute on a `th` so the table
 * keeps its native semantics, and carries `aria-sort` so a screen reader
 * announces the current order rather than just "button".
 */
@Component({
  selector: 'th[sei-sort]',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Icon],
  host: {
    class:
      'w-px border-b border-line px-3 py-2 text-left align-middle font-medium text-ink-dim whitespace-nowrap',
    '[attr.aria-sort]': 'ariaSort()',
  },
  template: `
    <button
      type="button"
      (click)="sort.emit(column())"
      class="group -mx-1 inline-flex items-center gap-1 rounded-xs px-1 py-0.5 transition-colors hover:text-ink"
      [class.text-accent]="isActive()"
    >
      <ng-content />
      <span [class.opacity-0]="!isActive()" [class.group-hover:opacity-40]="!isActive()">
        <sei-icon [name]="isActive() && desc() ? 'arrow-down' : 'arrow-up'" [size]="12" />
      </span>
    </button>
  `,
})
export class SortHeader {
  readonly column = input.required<string>();
  readonly active = input<string>('');
  readonly desc = input(false);

  readonly sort = output<string>();

  readonly isActive = computed(() => this.active() === this.column());
  readonly ariaSort = computed(() =>
    this.isActive() ? (this.desc() ? 'descending' : 'ascending') : 'none',
  );
}
