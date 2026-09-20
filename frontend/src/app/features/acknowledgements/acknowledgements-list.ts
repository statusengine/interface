import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';
import { ListStore } from '../../core/list/list-store';
import type { Acknowledgement } from '../../core/api/types';
import { DataTable } from '../../shared/ui/data-table';
import { PageHeader } from '../../shared/ui/page-header';
import { Pagination } from '../../shared/ui/pagination';
import { SearchInput } from '../../shared/ui/search-input';
import { SortHeader } from '../../shared/ui/sort-header';
import { StateBadge } from '../../shared/ui/state-badge';
import { Toolbar } from '../../shared/ui/toolbar';
import { SincePipe } from '../../shared/pipes/since.pipe';
import { TimestampPipe } from '../../shared/pipes/timestamp.pipe';
import { stateClass } from '../../shared/state/state';

/** Who took ownership of what, and what they said about it. */
@Component({
  selector: 'sei-acknowledgements-list',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    RouterLink,
    TranslocoDirective,
    PageHeader,
    Toolbar,
    SearchInput,
    DataTable,
    SortHeader,
    Pagination,
    StateBadge,
    SincePipe,
    TimestampPipe,
  ],
  templateUrl: './acknowledgements-list.html',
})
export class AcknowledgementsList {
  readonly stateClass = stateClass;

  readonly store = new ListStore<Acknowledgement>({
    path: '/acknowledgements',
    defaultSort: 'entry_time',
    defaultDesc: true,
    filterKeys: ['kind', 'host', 'author'],
  });

  key(ack: Acknowledgement): string {
    return [ack.kind, ack.hostname, ack.service_description ?? '', ack.entry_time].join(' :: ');
  }
}
