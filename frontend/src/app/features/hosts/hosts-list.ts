import { ChangeDetectionStrategy, Component, computed } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';
import { ListStore } from '../../core/list/list-store';
import type { CommandTarget } from '../../core/commands/commands.service';
import { BulkActions } from '../commands/bulk-actions';
import type { HostStatus } from '../../core/api/types';
import { DataTable } from '../../shared/ui/data-table';
import { FilterToggle } from '../../shared/ui/filter-toggle';
import { PageHeader } from '../../shared/ui/page-header';
import { Pagination } from '../../shared/ui/pagination';
import { SearchInput } from '../../shared/ui/search-input';
import { SortHeader } from '../../shared/ui/sort-header';
import { StateBadge } from '../../shared/ui/state-badge';
import { StateFilter } from '../../shared/ui/state-filter';
import { Toolbar } from '../../shared/ui/toolbar';
import { DurationPipe } from '../../shared/pipes/duration.pipe';
import { SincePipe } from '../../shared/pipes/since.pipe';
import { HOST_STATES, stateClass } from '../../shared/state/state';
import { RowFlags } from '../../shared/ui/row-flags';

@Component({
  selector: 'sei-hosts-list',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    RouterLink,
    TranslocoDirective,
    PageHeader,
    Toolbar,
    SearchInput,
    StateFilter,
    FilterToggle,
    DataTable,
    SortHeader,
    Pagination,
    BulkActions,
    StateBadge,
    RowFlags,
    DurationPipe,
    SincePipe,
  ],
  templateUrl: './hosts-list.html',
})
export class HostsList {
  readonly hostStates = HOST_STATES;
  readonly stateClass = stateClass;

  // Constructed in a field initializer, which runs inside the
  // component's injection context - that is what lets ListStore reach
  // for Api, Router and ActivatedRoute itself.
  readonly store = new ListStore<HostStatus>({
    path: '/hosts',
    defaultSort: 'severity',
    defaultDesc: true,
    filterKeys: [
      'state',
      'acknowledged',
      'in_downtime',
      'flapping',
      'notifications_enabled',
      'active_checks_enabled',
      'hard_state',
      'handled',
    ],
    keyOf: (host) => host.hostname,
  });

  /** Seconds the host has been in its current state. */
  /** What the bulk bar acts on: whatever is ticked, right now. */
  readonly selectedTargets = computed<CommandTarget[]>(() =>
    this.store.selectedRows().map((row) => ({ kind: 'host' as const, host: row.hostname })),
  );

  duration(host: HostStatus, now = Math.floor(Date.now() / 1000)): number {
    return host.last_state_change ? now - host.last_state_change : 0;
  }
}
