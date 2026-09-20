import { ChangeDetectionStrategy, Component, computed, input, output } from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';

/**
 * A filter with three positions: off, yes and no.
 *
 * A checkbox can only say "yes" or "did not ask", so filtering for the
 * absence of something - services with notifications turned off, problems
 * nobody has acknowledged - needs a third position. Clicking cycles
 * through them and the label says which one is showing, so the state is
 * never carried by colour alone.
 */
@Component({
  selector: 'sei-filter-toggle',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective],
  template: `
    <button
      type="button"
      (click)="cycle()"
      class="inline-flex items-center gap-1.5 whitespace-nowrap rounded-sm border px-2 py-1 text-[12px] transition-colors"
      [class.border-accent]="isSet()"
      [class.text-accent]="isSet()"
      [class.bg-accent-wash]="isSet()"
      [class.border-line]="!isSet()"
      [class.text-ink-dim]="!isSet()"
      [attr.aria-pressed]="isSet()"
      *transloco="let t"
    >
      {{ label() }}
      @if (isSet()) {
        <span class="font-medium">{{
          value() === 'true' ? t('filters.yes') : t('filters.no')
        }}</span>
      }
    </button>
  `,
})
export class FilterToggle {
  readonly label = input.required<string>();
  /** 'true', 'false' or undefined. Strings because that is what a URL
   *  query parameter holds, and the URL is the source of truth. */
  readonly value = input<string | undefined>(undefined);

  readonly valueChange = output<string | undefined>();

  readonly isSet = computed(() => this.value() !== undefined);

  cycle(): void {
    switch (this.value()) {
      case undefined:
        this.valueChange.emit('true');
        break;
      case 'true':
        this.valueChange.emit('false');
        break;
      default:
        this.valueChange.emit(undefined);
    }
  }
}
