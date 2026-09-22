import {
  ChangeDetectionStrategy,
  Component,
  computed,
  effect,
  inject,
  input,
  signal,
  untracked,
} from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';
import { Api } from '../../core/api/api.service';
import { ApiError } from '../../core/api/api.error';
import type { ListResponse, MetricMeta, MetricResult, Series } from '../../core/api/types';
import { Icon } from '../../shared/ui/icon';
import { MetricChart } from '../../shared/ui/metric-chart';
import { RangePicker } from '../../shared/ui/range-picker';
import { SincePipe } from '../../shared/pipes/since.pipe';

/** One chart's worth of series: everything that shares a unit. */
interface UnitGroup {
  unit: string;
  title: string;
  series: Series[];
}

/**
 * The performance data for one service.
 *
 * Series are grouped by unit and each group gets its own chart. A
 * service that reports both a percentage and a byte count has two
 * quantities that were never comparable; putting them on one plot with
 * two scales would let a reader compare them anyway, and the crossing
 * point would be an artefact of the scaling.
 */
@Component({
  selector: 'sei-metrics-panel',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective, MetricChart, RangePicker, Icon, SincePipe],
  template: `
    <section class="rounded-md border border-line bg-surface" *transloco="let t">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-line px-4 py-3">
        <h2 class="text-[13px] font-medium">{{ t('metrics.title') }}</h2>
        <sei-range-picker
          [from]="from()"
          [to]="to()"
          [defaultSeconds]="6 * 3600"
          (change)="setRange($event)"
        />
      </div>

      <div class="p-4">
        @if (error(); as err) {
          <div class="max-w-prose">
            <p class="text-[13px] text-ink">{{ t('metrics.errorTitle') }}</p>
            <p class="mt-1 text-[13px] text-ink-dim">{{ err.message }}</p>
            <button
              type="button"
              (click)="load()"
              class="mt-3 inline-flex items-center gap-1.5 rounded-sm border border-line px-2.5 py-1.5 text-[13px] transition-colors hover:border-line-strong"
            >
              <sei-icon name="refresh" [size]="14" />
              {{ t('list.retry') }}
            </button>
          </div>
        } @else if (loading() && groups().length === 0) {
          <p class="text-[13px] text-ink-dim">{{ t('metrics.loading') }}</p>
        } @else if (labels().length === 0) {
          <p class="max-w-prose text-[13px] text-ink-dim">{{ t('metrics.noneAtAll') }}</p>
        } @else if (groups().length === 0) {
          <div class="max-w-prose">
            <p class="text-[13px] text-ink">{{ t('metrics.emptyWindowTitle') }}</p>
            <p class="mt-1 text-[13px] text-ink-dim">
              {{ t('metrics.emptyWindowBody', { since: oldest() | seiSince }) }}
            </p>
          </div>
        } @else {
          <div class="grid gap-6" [class.xl:grid-cols-2]="groups().length > 1">
            @for (group of groups(); track group.unit) {
              <sei-metric-chart
                [title]="group.title"
                [series]="group.series"
                [unit]="group.unit"
                [bucketSeconds]="bucketSeconds()"
                [from]="windowFrom()"
                [to]="windowTo()"
              />
            }
          </div>
        }
      </div>
    </section>
  `,
})
export class MetricsPanel {
  private readonly api = inject(Api);

  readonly hostname = input.required<string>();
  readonly description = input.required<string>();

  readonly labels = signal<MetricMeta[]>([]);
  readonly result = signal<MetricResult | null>(null);
  readonly error = signal<ApiError | null>(null);
  readonly loading = signal(false);

  private readonly window = signal<{ from: string; to: string } | null>(null);

  readonly from = computed(() => this.window()?.from);
  readonly to = computed(() => this.window()?.to);
  readonly bucketSeconds = computed(() => this.result()?.bucket_seconds ?? 0);

  // The window the server answered for, so every chart draws the span
  // that was asked for rather than the span that happened to have data.
  readonly windowFrom = computed(() => this.result()?.from ?? 0);
  readonly windowTo = computed(() => this.result()?.to ?? 0);

  /** The earliest sample the service has, so an empty window can say how
   *  far back there is anything to see. */
  readonly oldest = computed(() => {
    const first = this.labels()
      .map((l) => l.first_seen)
      .filter((t) => t > 0);
    return first.length ? Math.min(...first) : 0;
  });

  readonly groups = computed<UnitGroup[]>(() => {
    const series = (this.result()?.series ?? []).filter((s) => (s.points ?? []).length > 0);
    const byUnit = new Map<string, Series[]>();
    for (const s of series) {
      const unit = s.unit ?? '';
      byUnit.set(unit, [...(byUnit.get(unit) ?? []), s]);
    }
    return [...byUnit.entries()]
      .map(([unit, members]) => ({
        unit,
        // With one unit group the service name is already the heading
        // above; naming the group after its only series is more useful.
        title:
          members.length === 1
            ? members[0].label || this.description()
            : unit || this.description(),
        series: members,
      }))
      .sort((a, b) => a.unit.localeCompare(b.unit));
  });

  constructor() {
    // Not in the constructor body: a required input has no value until
    // Angular sets it, and reading one before that is NG0950. An effect
    // also covers navigating straight from one service to another,
    // where the component is reused and only the inputs change.
    effect(() => {
      this.hostname();
      this.description();
      this.window();
      untracked(() => void this.load());
    });
  }

  setRange(range: { from: string; to: string }): void {
    this.window.set(range);
  }

  async load(): Promise<void> {
    this.loading.set(true);
    this.error.set(null);

    const window = this.window();
    const query: Record<string, string | number> = {
      host: this.hostname(),
      service: this.description(),
    };
    if (window) {
      query['from'] = window.from;
      query['to'] = window.to;
    }

    try {
      const [labels, result] = await Promise.all([
        this.api.list<MetricMeta>('/metrics/labels', {
          host: this.hostname(),
          service: this.description(),
        }) as Promise<ListResponse<MetricMeta>>,
        this.api.get<MetricResult>('/metrics/series', query),
      ]);
      this.labels.set(labels.data);
      this.result.set(result);
    } catch (err) {
      this.error.set(ApiError.from(err));
    } finally {
      this.loading.set(false);
    }
  }
}
