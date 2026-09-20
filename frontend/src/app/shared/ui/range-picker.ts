import { ChangeDetectionStrategy, Component, computed, input, output } from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';

/** Presets, in seconds. */
const PRESETS = [
  { key: '1h', seconds: 3600 },
  { key: '6h', seconds: 6 * 3600 },
  { key: '24h', seconds: 24 * 3600 },
  { key: '7d', seconds: 7 * 24 * 3600 },
  { key: '30d', seconds: 30 * 24 * 3600 },
] as const;

/**
 * How far back a history list reaches.
 *
 * Presets rather than two date fields, because the question an operator
 * has is almost always "since when did this start going wrong", and
 * because a window is not optional here: the history tables are large and
 * three of them have no index on their time column, so an unbounded list
 * is a table scan. The control makes the bound visible instead of hiding
 * it in a default.
 */
@Component({
  selector: 'sei-range-picker',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective],
  template: `
    <div class="flex items-center gap-1" role="group" *transloco="let t">
      <span class="mr-1 text-[12px] text-ink-dim">{{ t('range.label') }}</span>
      @for (preset of presets; track preset.key) {
        <button
          type="button"
          (click)="pick(preset.seconds)"
          class="rounded-sm border px-2 py-1 text-[12px] transition-colors"
          [class.border-accent]="active() === preset.seconds"
          [class.text-accent]="active() === preset.seconds"
          [class.bg-accent-wash]="active() === preset.seconds"
          [class.border-line]="active() !== preset.seconds"
          [class.text-ink-dim]="active() !== preset.seconds"
          [attr.aria-pressed]="active() === preset.seconds"
        >
          {{ t('range.' + preset.key) }}
        </button>
      }
    </div>
  `,
})
export class RangePicker {
  readonly presets = PRESETS;

  /** Unix seconds, as they travel in the URL. */
  readonly from = input<string | undefined>(undefined);
  readonly to = input<string | undefined>(undefined);
  readonly defaultSeconds = input(24 * 3600);

  readonly change = output<{ from: string; to: string }>();

  /** Which preset the current window corresponds to, if any. Matching
   *  with a minute of slack, because `to` is a wall-clock moment that has
   *  moved on since the URL was written. */
  readonly active = computed(() => {
    const from = Number(this.from());
    const to = Number(this.to());
    if (!Number.isFinite(from) || !Number.isFinite(to) || from <= 0 || to <= 0) {
      return this.defaultSeconds();
    }
    const span = to - from;
    const match = PRESETS.find((p) => Math.abs(p.seconds - span) < 60);
    return match?.seconds ?? 0;
  });

  pick(seconds: number): void {
    const now = Math.floor(Date.now() / 1000);
    this.change.emit({ from: String(now - seconds), to: String(now) });
  }
}
