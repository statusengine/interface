import { ChangeDetectionStrategy, Component, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';
import { ListStore } from '../../core/list/list-store';
import { COMMAND_ACTIONS, type AuditRecord } from '../../core/api/types';
import { DataTable } from '../../shared/ui/data-table';
import { PageHeader } from '../../shared/ui/page-header';
import { Pagination } from '../../shared/ui/pagination';
import { RangePicker } from '../../shared/ui/range-picker';
import { SearchInput } from '../../shared/ui/search-input';
import { SortHeader } from '../../shared/ui/sort-header';
import { Toolbar } from '../../shared/ui/toolbar';
import { SincePipe } from '../../shared/pipes/since.pipe';
import { TimestampPipe } from '../../shared/pipes/timestamp.pipe';

/**
 * Every external command this interface has submitted.
 *
 * The log is written whatever the outcome, so this is also the page that
 * answers "why did nothing happen when I clicked that". A row is one
 * object: a downtime over forty services reads as forty lines, and the
 * one that was refused is visible among them instead of hidden behind a
 * single summary.
 */
@Component({
  selector: 'sei-audit-list',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    RouterLink,
    TranslocoDirective,
    PageHeader,
    Toolbar,
    SearchInput,
    RangePicker,
    DataTable,
    SortHeader,
    Pagination,
    SincePipe,
    TimestampPipe,
  ],
  templateUrl: './audit-list.html',
})
export class AuditList {
  readonly actions = COMMAND_ACTIONS;

  readonly store = new ListStore<AuditRecord>({
    path: '/commands/audit',
    defaultSort: 'created_at',
    defaultDesc: true,
    filterKeys: ['username', 'action', 'failed', 'from', 'to'],
  });

  /** The broker answers 202 when it has taken the command. Anything else
   *  is a submission that went nowhere. */
  accepted(record: AuditRecord): boolean {
    return record.http_status === 202;
  }

  /** A host name cannot contain a slash; a service description can, so
   *  only the first one separates the two. */
  host(record: AuditRecord): string {
    const slash = record.target.indexOf('/');
    return slash === -1 ? record.target : record.target.slice(0, slash);
  }

  service(record: AuditRecord): string | undefined {
    const slash = record.target.indexOf('/');
    return slash === -1 ? undefined : record.target.slice(slash + 1);
  }

  /** Which payloads are open. Only the row somebody asked about grows a
   *  second line; a log where every row is two lines holds half as much. */
  private readonly open = signal<ReadonlySet<number>>(new Set());

  expanded(record: AuditRecord): boolean {
    return this.open().has(record.id);
  }

  toggle(record: AuditRecord): void {
    const next = new Set(this.open());
    if (!next.delete(record.id)) {
      next.add(record.id);
    }
    this.open.set(next);
  }

  /** Pretty-printed, because a payload is read to find one field in it. */
  payload(record: AuditRecord): string {
    if (record.payload === undefined || record.payload === null) {
      return '';
    }
    return JSON.stringify(record.payload, null, 2);
  }

  setRange(range: { from: string; to: string }): void {
    this.store.setFilter('from', range.from);
    this.store.setFilter('to', range.to);
  }

  /** Clicking a name narrows to that person. The filter is an exact
   *  match on the server, which the free-text search is not. */
  filterByUser(username: string): void {
    this.store.setFilter('username', username);
  }
}
