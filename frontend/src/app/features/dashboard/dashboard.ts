import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslocoDirective, TranslocoService } from '@jsverse/transloco';
import { Api } from '../../core/api/api.service';
import { ApiError } from '../../core/api/api.error';
import { apiErrorText } from '../../core/api/error-text';
import type { ListResponse, Problem, Summary } from '../../core/api/types';
import { Icon } from '../../shared/ui/icon';
import { PageHeader } from '../../shared/ui/page-header';
import { StateBadge } from '../../shared/ui/state-badge';
import { StateStrip } from '../../shared/ui/state-strip';
import { DurationPipe } from '../../shared/pipes/duration.pipe';
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
        this.api.get<Summary>('/summary'),
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

  key(problem: Problem): string {
    return problem.kind + ' :: ' + problem.hostname + ' :: ' + (problem.service_description ?? '');
  }

  duration(problem: Problem, now = Math.floor(Date.now() / 1000)): number {
    return problem.last_state_change ? now - problem.last_state_change : 0;
  }
}
