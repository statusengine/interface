import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';
import { ListStore } from '../../core/list/list-store';
import type { NotificationRecord } from '../../core/api/types';
import { DataTable } from '../../shared/ui/data-table';
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
 * Who was told about what, and when.
 *
 * One row per contact actually notified - the worker writes these from
 * the broker's contactnotificationmethod events, so a notification that
 * reached three people is three rows. That is the useful grain when the
 * question is "did anyone hear about this".
 */
@Component({
  selector: 'sei-notifications-history',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    RouterLink,
    TranslocoDirective,
    PageHeader,
    HistoryNav,
    HistoryScope,
    SearchInput,
    DataTable,
    SortHeader,
    Pagination,
    StateBadge,
    TimestampPipe,
  ],
  templateUrl: './notifications-history.html',
})
export class NotificationsHistory {
  readonly stateClass = stateClass;

  readonly store = new ListStore<NotificationRecord>({
    path: '/history/notifications',
    defaultSort: 'start_time',
    defaultDesc: true,
    filterKeys: ['host', 'service', 'from', 'to', 'kind', 'state'],
  });

  key(n: NotificationRecord): string {
    return [n.kind, n.hostname, n.service_description ?? '', n.start_time, n.contact_name].join(
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
