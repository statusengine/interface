import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { ActivatedRoute, RouterLink, RouterLinkActive } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';
import { ListStore } from '../../core/list/list-store';
import type { Downtime } from '../../core/api/types';
import { DataTable } from '../../shared/ui/data-table';
import { FilterToggle } from '../../shared/ui/filter-toggle';
import { PageHeader } from '../../shared/ui/page-header';
import { Pagination } from '../../shared/ui/pagination';
import { SearchInput } from '../../shared/ui/search-input';
import { SortHeader } from '../../shared/ui/sort-header';
import { Toolbar } from '../../shared/ui/toolbar';
import { DurationPipe } from '../../shared/pipes/duration.pipe';
import { TimestampPipe } from '../../shared/pipes/timestamp.pipe';

/**
 * Maintenance windows, current or finished.
 *
 * One component for both, because the columns are the same and the only
 * real difference is whether a row can still end. Which one this is comes
 * from the route's data rather than an input: the store is built in a
 * field initializer, and a bound input is not set yet at that point.
 */
@Component({
  selector: 'sei-downtimes-list',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    RouterLink,
    RouterLinkActive,
    TranslocoDirective,
    PageHeader,
    Toolbar,
    SearchInput,
    FilterToggle,
    DataTable,
    SortHeader,
    Pagination,
    DurationPipe,
    TimestampPipe,
  ],
  templateUrl: './downtimes-list.html',
})
export class DowntimesList {
  readonly history: boolean = inject(ActivatedRoute).snapshot.data['history'] === true;

  readonly store = new ListStore<Downtime>({
    path: this.history ? '/downtimes/history' : '/downtimes',
    defaultSort: 'scheduled_start_time',
    defaultDesc: true,
    filterKeys: ['kind', 'host', 'running'],
  });

  readonly titleKey = this.history ? 'downtimeHistory' : 'downtimes';

  key(downtime: Downtime): string {
    return [
      downtime.kind,
      downtime.hostname,
      downtime.service_description ?? '',
      downtime.internal_id,
    ].join(' :: ');
  }

  /** How long the window lasts, from its scheduled bounds. */
  span(downtime: Downtime): number {
    return Math.max(0, downtime.scheduled_end_time - downtime.scheduled_start_time);
  }

  /** Running right now, as opposed to scheduled for later or finished. */
  isRunning(downtime: Downtime, now = Math.floor(Date.now() / 1000)): boolean {
    return (
      downtime.was_started &&
      downtime.scheduled_start_time <= now &&
      now < downtime.scheduled_end_time
    );
  }
}
