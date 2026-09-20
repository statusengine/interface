import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';
import { ListStore } from '../../core/list/list-store';
import type { Problem } from '../../core/api/types';
import { DataTable } from '../../shared/ui/data-table';
import { FilterToggle } from '../../shared/ui/filter-toggle';
import { PageHeader } from '../../shared/ui/page-header';
import { Pagination } from '../../shared/ui/pagination';
import { RowFlags } from '../../shared/ui/row-flags';
import { SearchInput } from '../../shared/ui/search-input';
import { SortHeader } from '../../shared/ui/sort-header';
import { StateBadge } from '../../shared/ui/state-badge';
import { Toolbar } from '../../shared/ui/toolbar';
import { DurationPipe } from '../../shared/pipes/duration.pipe';
import { SincePipe } from '../../shared/pipes/since.pipe';
import { stateClass } from '../../shared/state/state';

/**
 * The triage list: everything that is not OK, worst first, hosts and
 * services together.
 *
 * Two tables would make an operator reconcile them by hand during the
 * exact minutes they can least afford it. One list, one order.
 */
@Component({
  selector: 'sei-problems-list',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    RouterLink,
    TranslocoDirective,
    PageHeader,
    Toolbar,
    SearchInput,
    FilterToggle,
    DataTable,
    SortHeader,
    Pagination,
    StateBadge,
    RowFlags,
    DurationPipe,
    SincePipe,
  ],
  templateUrl: './problems-list.html',
})
export class ProblemsList {
  readonly stateClass = stateClass;

  readonly store = new ListStore<Problem>({
    path: '/problems',
    defaultSort: 'severity',
    defaultDesc: true,
    filterKeys: [
      'kind',
      'acknowledged',
      'in_downtime',
      'hard_state',
      'flapping',
      'handled',
      'hide_services_of_down_hosts',
    ],
  });

  key(problem: Problem): string {
    return problem.kind + ' :: ' + problem.hostname + ' :: ' + (problem.service_description ?? '');
  }

  duration(problem: Problem, now = Math.floor(Date.now() / 1000)): number {
    return problem.last_state_change ? now - problem.last_state_change : 0;
  }

  /** Cycles the kind filter: everything, hosts only, services only. */
  cycleKind(): void {
    const current = this.store.filters()['kind'];
    this.store.setFilter(
      'kind',
      current === undefined ? 'host' : current === 'host' ? 'service' : undefined,
    );
  }

  kindLabelKey(): string {
    const current = this.store.filters()['kind'];
    return current === 'host'
      ? 'filters.hosts'
      : current === 'service'
        ? 'filters.services'
        : 'filters.all';
  }
}
