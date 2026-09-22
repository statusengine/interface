import {
  ChangeDetectionStrategy,
  Component,
  DestroyRef,
  ElementRef,
  computed,
  effect,
  inject,
  input,
  signal,
  viewChild,
} from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';
import uPlot from 'uplot';
import type { Series } from '../../core/api/types';
import { Theme } from '../../core/theme/theme.service';
import {
  axisRange,
  formatAxisValue,
  formatDuration,
  formatValue,
  isDarkSurface,
  washFor,
  withAlpha,
} from './metric-format';

/** The eight categorical slots, in fixed order. Never cycled. */
const SERIES_SLOTS = 8;

interface LegendEntry {
  label: string;
  colour: string;
  /** Value under the cursor, or the last value when the cursor is away. */
  value: string;
  slot: number;
}

/**
 * One metric over time.
 *
 * Every series on this chart shares a unit - the caller groups by unit
 * and renders one chart per group. That is not a stylistic preference:
 * two y-scales on one plot let a reader compare two quantities that were
 * never comparable, and the crossing point is an artefact of the
 * scaling. Two charts say the same thing without the lie.
 *
 * The values are downsampled averages. With a single series the min-max
 * band is drawn behind the line, because averaging a five-minute bucket
 * hides the spike that caused the alert, which is usually the only
 * reason anyone opened the chart. With more than one series the bands
 * would overlap into mud, so the legend says the line is an average and
 * the band is not drawn.
 */
@Component({
  selector: 'sei-metric-chart',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective],
  template: `
    <figure class="m-0" *transloco="let t">
      <figcaption class="mb-2 flex flex-wrap items-baseline justify-between gap-2">
        <h3 class="text-[13px] font-medium">
          {{ title() }}
          @if (unit()) {
            <span class="ml-1 font-normal text-ink-dim">({{ unit() }})</span>
          }
        </h3>
        <span class="text-[11px] text-ink-dim">{{ resolutionLabel() }}</span>
      </figcaption>

      <div #host class="w-full"></div>

      <!-- Always present for two or more series, and the visible values
           are also the relief the palette check asks for on the three
           light-mode hues that fall below 3:1 against white. -->
      <ul class="mt-2 flex flex-wrap gap-x-4 gap-y-1">
        @for (entry of legend(); track entry.label) {
          <li class="flex items-baseline gap-1.5 text-[12px]">
            <span
              class="inline-block h-0.5 w-3 self-center rounded-full"
              [style.background]="entry.colour"
              aria-hidden="true"
            ></span>
            <span class="mono text-ink-dim">{{ entry.label || t('metrics.value') }}</span>
            <span class="mono tabular-nums font-medium text-ink">{{ entry.value }}</span>
          </li>
        }
      </ul>
    </figure>
  `,
  styles: `
    /* uPlot ships its own stylesheet; these are the few rules that make
       it wear this application's tokens instead. */
    :host ::ng-deep .u-wrap {
      position: relative;
    }
    :host ::ng-deep .u-over {
      cursor: crosshair;
    }
    :host ::ng-deep .u-cursor-x,
    :host ::ng-deep .u-cursor-y {
      border-color: var(--line-strong);
    }
    :host ::ng-deep .u-legend {
      display: none;
    }
  `,
})
export class MetricChart {
  private readonly themeService = inject(Theme);
  private readonly destroyRef = inject(DestroyRef);

  readonly host = viewChild.required<ElementRef<HTMLDivElement>>('host');

  readonly title = input.required<string>();
  readonly series = input.required<Series[]>();
  readonly unit = input('');
  readonly bucketSeconds = input(0);

  /** The window the data was asked for, in unix seconds.
   *
   * Without it the axis fits whatever came back, which is wrong twice
   * over. A gap rescales the axis silently, so "the last six hours"
   * quietly becomes "the two hours that happen to have data". And a
   * single point has no range at all: uPlot invents one, and a six-hour
   * window renders as an axis running from 2026 to 2029. That is not a
   * hypothetical - it is what a chart looks like the first time anybody
   * opens it after the monitoring was down.
   */
  readonly from = input(0);
  readonly to = input(0);

  readonly height = input(180);

  /** Index of the point under the cursor, or null. */
  private readonly cursorIndex = signal<number | null>(null);

  private plot?: uPlot;
  private observer?: ResizeObserver;

  readonly resolutionLabel = computed(() => {
    const bucket = this.bucketSeconds();
    if (!bucket) {
      return '';
    }
    return `${formatDuration(bucket)} average`;
  });

  readonly legend = computed<LegendEntry[]>(() => {
    const index = this.cursorIndex();
    return this.series().map((s, i) => {
      const points = s.points ?? [];
      const point = index !== null && index < points.length ? points[index] : points.at(-1);
      return {
        label: s.label,
        colour: `var(--series-${(i % SERIES_SLOTS) + 1})`,
        value: point ? formatValue(point.avg, s.unit || this.unit()) : '—',
        slot: i,
      };
    });
  });

  constructor() {
    // Rebuild on data, size or theme change. uPlot resolves its colours
    // once at construction, so a theme flip needs a new instance rather
    // than a redraw.
    effect(() => {
      this.series();
      this.height();
      this.themeService.preference();
      queueMicrotask(() => this.render());
    });

    this.destroyRef.onDestroy(() => {
      this.observer?.disconnect();
      this.plot?.destroy();
    });
  }

  private xRange(): uPlot.Scale['range'] {
    return (_u, min, max) => axisRange(this.from(), this.to(), min, max, this.bucketSeconds());
  }

  private render(): void {
    const element = this.host().nativeElement;
    const series = this.series();

    this.plot?.destroy();
    this.plot = undefined;
    element.replaceChildren();

    if (series.length === 0 || series.every((s) => (s.points ?? []).length === 0)) {
      return;
    }

    const styles = getComputedStyle(element);
    const token = (name: string) => styles.getPropertyValue(name).trim();

    const ink = token('--ink-dim') || '#888';
    const line = token('--line') || '#ccc';
    const colours = Array.from({ length: SERIES_SLOTS }, (_, i) => token(`--series-${i + 1}`));

    // Every series shares the caller's window, but downsampling can drop
    // a bucket where a series had no sample, so the x axis is the union
    // of every timestamp and each series is aligned onto it. Without
    // this, two series with different gaps would be drawn against each
    // other's timestamps.
    const times = [...new Set(series.flatMap((s) => (s.points ?? []).map((p) => p.t)))].sort(
      (a, b) => a - b,
    );

    const withBand = series.length === 1;
    const data: uPlot.AlignedData = [times];
    const plotSeries: uPlot.Series[] = [{}];

    if (withBand) {
      const byTime = new Map((series[0].points ?? []).map((p) => [p.t, p]));
      data.push(times.map((t) => byTime.get(t)?.max ?? null));
      data.push(times.map((t) => byTime.get(t)?.min ?? null));
      plotSeries.push(
        { stroke: 'transparent', points: { show: false } },
        { stroke: 'transparent', points: { show: false } },
      );
    }

    // How strongly to shade under a line.
    //
    // One series can carry a visible wash; several cannot, because
    // where they overlap the alphas add up and the chart turns to soup.
    // Past three, the lines have to speak for themselves.
    //
    // And a translucent colour carries further on white than on a dark
    // surface: the same alpha that reads as a tint in daylight is a
    // rumour at night. The theme is read from the surface token rather
    // than from matchMedia, so the manual switch counts too.
    const dark = isDarkSurface(token('--surface'));
    const wash = washFor(series.length, dark);

    for (const [i, s] of series.entries()) {
      const byTime = new Map((s.points ?? []).map((p) => [p.t, p]));
      const colour = colours[i % SERIES_SLOTS];
      data.push(times.map((t) => byTime.get(t)?.avg ?? null));
      plotSeries.push({
        label: s.label,
        stroke: colour,
        width: 2,
        ...(wash
          ? {
              fill: verticalWash(colour, wash.top, wash.floor),
              // Only when the whole range sits above zero: then there is
              // no baseline on screen and shading to the bottom edge is
              // what a monitoring chart is read as. When zero is in
              // range, uPlot fills to it, which is the honest reading -
              // and the only one that makes sense for a series that goes
              // negative.
              fillTo: (u: uPlot) => {
                const min = u.scales['y']?.min ?? 0;
                return min > 0 ? min : 0;
              },
            }
          : {}),
        // Markers only once the points are sparse enough to be distinct;
        // a dot on every sample of a dense series is a thick line.
        points: { show: (_u, _i, i0, i1) => i1 - i0 < 40, size: 5 },
      });
    }

    const options: uPlot.Options = {
      width: element.clientWidth || 600,
      height: this.height(),
      padding: [8, 8, 0, 0],
      series: plotSeries,
      legend: { show: false },
      cursor: {
        y: false,
        points: { size: 7 },
      },
      scales: { x: { time: true, range: this.xRange() } },
      axes: [
        {
          stroke: ink,
          grid: { stroke: line, width: 1 },
          ticks: { stroke: line, width: 1 },
          font: '11px "IBM Plex Sans", sans-serif',
        },
        {
          stroke: ink,
          grid: { stroke: line, width: 1 },
          ticks: { stroke: line, width: 1 },
          font: '11px "IBM Plex Mono", monospace',
          size: 56,
          values: (_u, splits) => splits.map((v) => formatAxisValue(v)),
        },
      ],
      bands: withBand ? [{ series: [1, 2], fill: withAlpha(colours[0], 0.16), dir: 1 }] : undefined,
      hooks: {
        setCursor: [
          (u) => {
            this.cursorIndex.set(u.cursor.idx ?? null);
          },
        ],
      },
    };

    this.plot = new uPlot(options, data, element);

    this.observer?.disconnect();
    this.observer = new ResizeObserver(() => {
      if (this.plot && element.clientWidth > 0) {
        this.plot.setSize({ width: element.clientWidth, height: this.height() });
      }
    });
    this.observer.observe(element);
  }
}

/**
 * A wash under a line: the series colour at the top, fading out
 * downwards.
 *
 * A flat fill at this strength reads as a block of colour with a line
 * on top of it - the shape a stock ticker draws, where the area is the
 * point. Here the line is the point and the fill is only there to give
 * it a body, so it fades: strongest where it meets the line, gone by
 * the bottom of the plot.
 */
function verticalWash(colour: string, alpha: number, floor: number) {
  return (u: uPlot, seriesIdx: number): CanvasGradient | string => {
    const bottom = u.bbox.top + u.bbox.height;

    // Anchored to the series, not to the plot box. Anchored to the box,
    // how strong the wash looks depends on where the line happens to
    // sit: a series drawn across the lower third gets the faded end of
    // the gradient and reads as having no fill at all. Starting at the
    // series' own high point means the wash is always at full strength
    // where it meets the line.
    let top = u.bbox.top;
    const values = u.data[seriesIdx] as (number | null)[] | undefined;
    if (values) {
      let highest: number | null = null;
      for (const v of values) {
        if (v !== null && v !== undefined && (highest === null || v > highest)) {
          highest = v;
        }
      }
      if (highest !== null) {
        top = Math.min(bottom - 1, Math.max(u.bbox.top, u.valToPos(highest, 'y', true)));
      }
    }

    const gradient = u.ctx.createLinearGradient(0, top, 0, bottom);
    gradient.addColorStop(0, withAlpha(colour, alpha));
    // Not all the way to nothing: a wash that ends at zero leaves the
    // bottom of a tall plot looking like the line stopped having an
    // underside.
    gradient.addColorStop(1, withAlpha(colour, floor));
    return gradient;
  };
}
