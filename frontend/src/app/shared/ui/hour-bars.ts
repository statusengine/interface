import { ChangeDetectionStrategy, Component, computed, inject, input } from '@angular/core';
import { TranslocoService } from '@jsverse/transloco';
import type { HourBucket } from '../../core/api/types';
import { formatHour } from '../pipes/time-format';

interface Bar {
  t: number;
  count: number;
  /** Height as a percentage of the tallest bar. */
  height: number;
  when: string;
  peak: boolean;
}

/**
 * Counts per hour, as bars.
 *
 * Bars rather than a line: these are counts in separate buckets, not a
 * continuous measurement, and a line between two hours implies values
 * in between that nobody recorded.
 *
 * An hour with nothing in it keeps a one-pixel foot on the baseline.
 * Drawing it as blank space would make "we measured, and nothing
 * happened" look the same as "we have no data for this hour".
 */
@Component({
  selector: 'sei-hour-bars',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <figure class="m-0">
      <div class="flex items-baseline justify-between gap-3">
        <figcaption class="text-[11px] text-ink-dim">{{ label() }}</figcaption>
        <span class="text-[11px] tabular-nums text-ink-faint">{{ scaleLabel() }}</span>
      </div>

      <div class="mt-2 flex h-20 items-end gap-0.5" role="img" [attr.aria-label]="summary()">
        @for (bar of bars(); track bar.t) {
          <!-- One colour for every bar: the tallest one is already the
               tallest, and a second colour saying so would be the only
               thing in the chart carried by colour alone. -->
          <div
            class="min-h-px flex-1 rounded-t-[3px] bg-accent"
            [style.height.%]="bar.height"
            [title]="bar.when + ' · ' + bar.count"
          ></div>
        }
      </div>

      <div class="mt-1 flex justify-between text-[11px] tabular-nums text-ink-faint">
        <span>{{ first() }}</span>
        <span>{{ last() }}</span>
      </div>
    </figure>
  `,
})
export class HourBars {
  private readonly transloco = inject(TranslocoService);

  readonly buckets = input.required<HourBucket[]>();
  /** Names the series. With one series there is no legend to name it. */
  readonly label = input('');

  private readonly max = computed(() =>
    this.buckets().reduce((most, b) => Math.max(most, b.count), 0),
  );

  readonly bars = computed<Bar[]>(() => {
    const max = this.max();
    return this.buckets().map((b) => ({
      t: b.t,
      count: b.count,
      // Against the tallest bar, so a quiet day still shows its shape.
      // The floor keeps an empty hour visible as a measured zero.
      height: max === 0 ? 0 : Math.max(1.5, (b.count / max) * 100),
      when: this.when(b.t),
      peak: b.count > 0 && b.count === max,
    }));
  });

  readonly scaleLabel = computed(() => (this.max() === 0 ? '' : `max ${this.max()}`));
  readonly first = computed(() => this.when(this.buckets()[0]?.t));
  readonly last = computed(() => this.when(this.buckets().at(-1)?.t));

  /** What a screen reader gets instead of the bars. */
  readonly summary = computed(() => {
    const total = this.buckets().reduce((sum, b) => sum + b.count, 0);
    const peak = this.bars().find((b) => b.peak);
    return this.transloco.translate('dashboard.trendSummary', {
      total,
      hours: this.buckets().length,
      peak: peak ? peak.count : 0,
      at: peak ? peak.when : '—',
    });
  });

  private when(t: number | undefined): string {
    if (!t) {
      return '';
    }
    return formatHour(t, this.transloco.getActiveLang(), this.buckets().length);
  }
}
