import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';
import { ListStore } from '../../core/list/list-store';
import type { ServiceStatus } from '../../core/api/types';
import { DataTable } from '../../shared/ui/data-table';
import { FilterToggle } from '../../shared/ui/filter-toggle';
import { PageHeader } from '../../shared/ui/page-header';
import { Pagination } from '../../shared/ui/pagination';
import { RowFlags } from '../../shared/ui/row-flags';
import { SearchInput } from '../../shared/ui/search-input';
import { SortHeader } from '../../shared/ui/sort-header';
import { StateBadge } from '../../shared/ui/state-badge';
import { StateFilter } from '../../shared/ui/state-filter';
import { Toolbar } from '../../shared/ui/toolbar';
import { DurationPipe } from '../../shared/pipes/duration.pipe';
import { SincePipe } from '../../shared/pipes/since.pipe';
import { SERVICE_STATES, stateClass } from '../../shared/state/state';

@Component({
  selector: 'sei-services-list',
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
    StateBadge,
    RowFlags,
    DurationPipe,
    SincePipe,
  ],
  templateUrl: './services-list.html',
})
export class ServicesList {
  readonly serviceStates = SERVICE_STATES;
  readonly stateClass = stateClass;

  readonly store = new ListStore<ServiceStatus>({
    path: '/services',
    defaultSort: 'severity',
    defaultDesc: true,
    filterKeys: [
      'state',
      'host',
      'acknowledged',
      'in_downtime',
      'flapping',
      'notifications_enabled',
      'active_checks_enabled',
      'hard_state',
      'handled',
    ],
  });

  /** A service is identified by host plus description, so @for needs
   *  both. The separator only has to be stable, not unguessable. */
  key(service: ServiceStatus): string {
    return service.hostname + ' :: ' + service.service_description;
  }

  duration(service: ServiceStatus, now = Math.floor(Date.now() / 1000)): number {
    return service.last_state_change ? now - service.last_state_change : 0;
  }
}
