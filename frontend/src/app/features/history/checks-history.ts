import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';
import { ListStore } from '../../core/list/list-store';
import type { CheckResult } from '../../core/api/types';
import { DataTable } from '../../shared/ui/data-table';
import { FilterToggle } from '../../shared/ui/filter-toggle';
import { PageHeader } from '../../shared/ui/page-header';
import { Pagination } from '../../shared/ui/pagination';
import { SearchInput } from '../../shared/ui/search-input';
import { SortHeader } from '../../shared/ui/sort-header';
import { StateBadge } from '../../shared/ui/state-badge';
import { TimestampPipe } from '../../shared/pipes/timestamp.pipe';
import { stateClass } from '../../shared/state/state';
import { HistoryNav } from './history-nav';
import { HistoryScope } from './history-scope';

/** Every check the monitoring core has executed, in a window. */
@Component({
  selector: 'sei-checks-history',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    RouterLink,
    TranslocoDirective,
    PageHeader,
    HistoryNav,
    HistoryScope,
    SearchInput,
    FilterToggle,
    DataTable,
    SortHeader,
    Pagination,
    StateBadge,
    TimestampPipe,
  ],
  templateUrl: './checks-history.html',
})
export class ChecksHistory {
  readonly stateClass = stateClass;

  readonly store = new ListStore<CheckResult>({
    path: '/history/checks',
    defaultSort: 'start_time',
    defaultDesc: true,
    filterKeys: ['host', 'service', 'from', 'to', 'kind', 'state', 'hard_only'],
  });

  key(check: CheckResult): string {
    return [check.kind, check.hostname, check.service_description ?? '', check.start_time].join(
      ' :: ',
    );
  }

  /** How long the check itself took, in milliseconds - the number that
   *  explains a check that keeps timing out. */
  durationMs(check: CheckResult): number {
    return Math.round(check.execution_time * 1000);
  }

  setHost(host: string | undefined): void {
    this.store.setFilter('host', host);
    // A service filter without its host is meaningless and the API
    // rejects it, so clearing the host clears both.
    if (!host) {
      this.store.setFilter('service', undefined);
    }
  }

  setRange(range: { from: string; to: string }): void {
    this.store.setFilter('from', range.from);
    this.store.setFilter('to', range.to);
  }
}
