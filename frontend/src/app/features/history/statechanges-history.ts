import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';
import { ListStore } from '../../core/list/list-store';
import type { StateChange } from '../../core/api/types';
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

/**
 * When things changed state, and to what.
 *
 * The table defaults to transitions only. The statehistory tables also
 * carry a row for every retry that found the same state again, and
 * "when did this break" is not a question those rows answer - they
 * bury it.
 */
@Component({
  selector: 'sei-statechanges-history',
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
  templateUrl: './statechanges-history.html',
})
export class StateChangesHistory {
  readonly stateClass = stateClass;

  readonly store = new ListStore<StateChange>({
    path: '/history/statechanges',
    defaultSort: 'state_time',
    defaultDesc: true,
    filterKeys: ['host', 'service', 'from', 'to', 'kind', 'state', 'hard_only', 'transitions_only'],
  });

  key(change: StateChange): string {
    return [change.kind, change.hostname, change.service_description ?? '', change.state_time].join(
      ' :: ',
    );
  }

  setHost(host: string | undefined): void {
    this.store.setFilter('host', host);
    if (!host) {
      this.store.setFilter('service', undefined);
    }
  }

  setRange(range: { from: string; to: string }): void {
    this.store.setFilter('from', range.from);
    this.store.setFilter('to', range.to);
  }
}
