import { ChangeDetectionStrategy, Component, input } from '@angular/core';
import { RouterLink } from '@angular/router';

/**
 * One number worth reading from across the room.
 *
 * The figure is the tile; the label says what it counts and the line
 * underneath says what it is out of. A percentage with no denominator
 * is how "50% of hosts are down" turns out to mean one host of two.
 *
 * No colour on the number unless something is actually wrong, and even
 * then the rail carries it rather than the digits: a tile that is green
 * when things are fine trains people to read colour instead of numbers.
 */
@Component({
  selector: 'sei-kpi-tile',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [RouterLink],
  template: `
    <div
      class="relative rounded-md border border-line bg-surface px-4 py-3"
      [class.rail]="tone() !== 'neutral'"
      [class.state-critical]="tone() === 'critical'"
      [class.state-warning]="tone() === 'warning'"
    >
      <p class="text-[11px] text-ink-dim">{{ label() }}</p>

      <p class="mt-1 flex items-baseline gap-1.5">
        <span class="text-[26px] font-medium leading-none tabular-nums tracking-tight">{{
          value()
        }}</span>
        @if (unit()) {
          <span class="text-[15px] text-ink-dim">{{ unit() }}</span>
        }
      </p>

      <p class="mt-1.5 text-[12px] text-ink-dim">
        @if (link()) {
          <a
            [routerLink]="link()"
            [queryParams]="linkParams()"
            class="hover:text-accent hover:underline"
          >
            {{ context() }}
          </a>
        } @else {
          {{ context() }}
        }
      </p>
    </div>
  `,
})
export class KpiTile {
  readonly label = input.required<string>();
  readonly value = input.required<string>();
  readonly unit = input('');
  readonly context = input('');

  /** Only 'critical' and 'warning' draw a rail. Good news needs none. */
  readonly tone = input<'neutral' | 'warning' | 'critical'>('neutral');

  readonly link = input<string | undefined>(undefined);
  readonly linkParams = input<Record<string, string>>({});
}
