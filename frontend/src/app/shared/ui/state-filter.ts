import { ChangeDetectionStrategy, Component, computed, input, output } from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';
import { stateTextClass, type StateKey } from '../state/state';

/**
 * Pick which states a list shows. Each option carries its own colour, so
 * the control teaches the same vocabulary the table uses.
 *
 * With nothing selected the list shows everything, which is what an
 * operator who has just cleared the last option means - not "show me
 * nothing".
 */
@Component({
  selector: 'sei-state-filter',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective],
  template: `
    <div class="flex flex-wrap items-center gap-1" role="group" *transloco="let t">
      @for (option of options(); track option.value) {
        <button
          type="button"
          (click)="toggle(option.value)"
          class="rounded-sm border px-2 py-1 text-[12px] font-medium transition-colors"
          [class]="
            selected().includes(option.value) ? activeClass(option.key) : 'border-line text-ink-dim'
          "
          [attr.aria-pressed]="selected().includes(option.value)"
        >
          {{ t('states.' + option.key) }}
        </button>
      }
    </div>
  `,
})
export class StateFilter {
  readonly options = input.required<{ value: number; key: StateKey }[]>();
  /** Comma-separated, as it travels in the URL. */
  readonly value = input<string | undefined>(undefined);

  readonly valueChange = output<string | undefined>();

  readonly selected = computed(() => {
    const raw = this.value();
    if (!raw) {
      return [] as number[];
    }
    return raw
      .split(',')
      .map((s) => Number(s.trim()))
      .filter((n) => Number.isInteger(n));
  });

  activeClass(key: StateKey): string {
    return `border-current ${stateTextClass(key)}`;
  }

  toggle(value: number): void {
    const current = this.selected();
    const next = current.includes(value)
      ? current.filter((v) => v !== value)
      : [...current, value].sort((a, b) => a - b);
    this.valueChange.emit(next.length ? next.join(',') : undefined);
  }
}
