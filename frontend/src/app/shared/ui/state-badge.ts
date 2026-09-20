import { ChangeDetectionStrategy, Component, computed, input } from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';
import { stateTextClass } from '../state/state';

/**
 * A state, named and coloured.
 *
 * The word carries the meaning and the colour only speeds it up, which is
 * what keeps this readable for a colour-blind operator and in a
 * screenshot pasted into a black-and-white ticket.
 *
 * A soft state is marked, because it is the difference between "act now"
 * and "the core is still deciding". Every other Nagios interface hides
 * that in a "2/3" column halfway across the screen.
 */
@Component({
  selector: 'sei-state-badge',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective],
  template: `
    <span class="inline-flex items-baseline gap-1.5 whitespace-nowrap" *transloco="let t">
      <span class="font-medium" [class]="colour()">{{ t('states.' + state()) }}</span>
      @if (!hard() && state() !== 'pending') {
        <span class="text-[11px] text-ink-dim" [title]="t('states.softTitle')">
          {{ t('states.soft') }}
        </span>
      }
    </span>
  `,
})
export class StateBadge {
  readonly state = input.required<string>();
  readonly hard = input(true);

  readonly colour = computed(() => stateTextClass(this.state()));
}
