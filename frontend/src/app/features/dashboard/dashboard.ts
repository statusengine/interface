import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslocoDirective, TranslocoService } from '@jsverse/transloco';
import { Api } from '../../core/api/api.service';
import { ApiError } from '../../core/api/api.error';
import { apiErrorText } from '../../core/api/error-text';
import type { HourBucket, ListResponse, Problem, Summary } from '../../core/api/types';
import { Icon } from '../../shared/ui/icon';
import { PageHeader } from '../../shared/ui/page-header';
import { StateBadge } from '../../shared/ui/state-badge';
import { StateStrip } from '../../shared/ui/state-strip';
import { HourBars } from '../../shared/ui/hour-bars';
import { KpiTile } from './kpi-tile';
import { DurationPipe } from '../../shared/pipes/duration.pipe';
import { formatHour } from '../../shared/pipes/time-format';
import { percentChange, share } from './dashboard-math';
import { SincePipe } from '../../shared/pipes/since.pipe';
import { HOST_STATES, SERVICE_STATES, stateClass } from '../../shared/state/state';

/**
 * What needs attention right now.
 *
 * The dashboard opens with the unhandled problems, not with counters. In
 * this domain the most characteristic thing is the list of what is wrong,
 * and a page of big numbers makes an operator take a second step to reach
 * it. The two strips above give the shape of the estate; the list below
 * is the work.
 */
@Component({
  selector: 'sei-dashboard',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    RouterLink,
    TranslocoDirective,
    PageHeader,
    StateStrip,
    HourBars,
    KpiTile,
    StateBadge,
    Icon,
    DurationPipe,
    SincePipe,
  ],
  templateUrl: './dashboard.html',
})
export class Dashboard {
  private readonly api = inject(Api);

  readonly hostStates = HOST_STATES;
  readonly serviceStates = SERVICE_STATES;
  readonly stateClass = stateClass;

  readonly summary = signal<Summary | null>(null);
  private readonly transloco = inject(TranslocoService);

  readonly summaryError = signal<ApiError | null>(null);

  /** The failure in the reader's language; the server's sentence is the
   *  fallback for a code with no wording yet. */
  readonly summaryErrorText = computed(() => apiErrorText(this.transloco, this.summaryError()));
  readonly loading = signal(true);

  readonly problems = signal<Problem[]>([]);
  readonly problemTotal = signal(0);

  /** How far back the recent figures reach. The two the server makes
   *  cheap: both read an indexed range, neither scans a history table. */
  readonly windows = [24, 168] as const;
  readonly hours = signal<number>(24);

  /** Share of the estate in its good state, per population. The
   *  denominator travels with it everywhere it is shown: "50% of hosts
   *  are down" has meant one host of two more than once. */
  readonly hostAvailability = computed(() =>
    share(this.summary()?.hosts.by_state['up'], this.summary()?.hosts.total),
  );
  readonly serviceAvailability = computed(() =>
    share(this.summary()?.services.by_state['ok'], this.summary()?.services.total),
  );

  readonly problemsTotal = computed(
    () => (this.summary()?.hosts.problems ?? 0) + (this.summary()?.services.problems ?? 0),
  );
  readonly unhandledTotal = computed(
    () => (this.summary()?.hosts.unhandled ?? 0) + (this.summary()?.services.unhandled ?? 0),
  );
  readonly unhandledShare = computed(() => share(this.unhandledTotal(), this.problemsTotal()));

  readonly objectsTotal = computed(
    () => (this.summary()?.hosts.total ?? 0) + (this.summary()?.services.total ?? 0),
  );
  readonly changedTotal = computed(
    () =>
      (this.summary()?.window.hosts_changed ?? 0) + (this.summary()?.window.services_changed ?? 0),
  );
  /** The inverse of what changed: what a manager wants is the good
   *  number, and "96% unchanged" is the same fact as "4% moved". */
  readonly steadyShare = computed(() => {
    const total = this.objectsTotal();
    return total === 0 ? null : share(total - this.changedTotal(), total);
  });

  readonly inDowntime = computed(
    () => (this.summary()?.hosts.in_downtime ?? 0) + (this.summary()?.services.in_downtime ?? 0),
  );

  readonly notifications = computed(() => this.summary()?.window.notifications ?? 0);

  /** How this window compares with the one before it. The chart already
   *  shows the busiest hour; what it cannot show is whether today is a
   *  bad day or an ordinary one. */
  readonly alertsChange = computed(() => {
    const window = this.summary()?.window;
    if (!window) {
      return null;
    }
    return percentChange(window.notifications, window.notifications_previous);
  });

  /** The busiest hour, which is what "19 alerts" leaves out: nineteen
   *  spread over a day is background noise, nineteen in one hour is an
   *  incident. */
  readonly peak = computed(() => {
    const buckets = this.summary()?.window.notifications_by_hour ?? [];
    const busiest = buckets.reduce<HourBucket | null>(
      (most, b) => (most === null || b.count > most.count ? b : most),
      null,
    );
    if (!busiest || busiest.count === 0) {
      return null;
    }
    return {
      count: busiest.count,
      at: formatHour(busiest.t, this.transloco.getActiveLang(), buckets.length),
    };
  });
  readonly trend = computed(() => this.summary()?.window.notifications_by_hour ?? []);
  readonly oldest = computed(() => this.summary()?.window.oldest_problem);

  /** How long the oldest unhandled problem has been in its state. */
  readonly oldestAge = computed(() => {
    const since = this.oldest()?.since ?? 0;
    return since ? Math.max(0, Math.floor(Date.now() / 1000) - since) : 0;
  });

  /** How stale the numbers are. The worker writes status_update_time on
   *  every update, so a dashboard that cannot say how current it is is
   *  worse than one that admits the data is an hour old. */
  readonly staleness = computed(() => {
    const last = this.summary()?.last_update ?? 0;
    return last ? Math.max(0, Math.floor(Date.now() / 1000) - last) : 0;
  });

  /** Over five minutes without an update usually means the worker or the
   *  core stopped, not that nothing happened. */
  readonly isStale = computed(() => this.staleness() > 300);

  constructor() {
    void this.load();
  }

  async load(): Promise<void> {
    this.loading.set(true);
    this.summaryError.set(null);
    try {
      const [summary, problems] = await Promise.all([
        this.api.get<Summary>('/summary', { hours: this.hours() }),
        this.api.list<Problem>('/problems', {
          handled: 'false',
          limit: 10,
          sort: 'severity:desc',
        }) as Promise<ListResponse<Problem>>,
      ]);
      this.summary.set(summary);
      this.problems.set(problems.data);
      this.problemTotal.set(problems.meta.total);
    } catch (err) {
      this.summaryError.set(ApiError.from(err));
    } finally {
      this.loading.set(false);
    }
  }

  setHours(hours: number): void {
    this.hours.set(hours);
    void this.load();
  }

  key(problem: Problem): string {
    return problem.kind + ' :: ' + problem.hostname + ' :: ' + (problem.service_description ?? '');
  }

  duration(problem: Problem, now = Math.floor(Date.now() / 1000)): number {
    return problem.last_state_change ? now - problem.last_state_change : 0;
  }
}
