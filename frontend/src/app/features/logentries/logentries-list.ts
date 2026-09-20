import { ChangeDetectionStrategy, Component } from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';
import { ListStore } from '../../core/list/list-store';
import type { LogEntry } from '../../core/api/types';
import { DataTable } from '../../shared/ui/data-table';
import { PageHeader } from '../../shared/ui/page-header';
import { Pagination } from '../../shared/ui/pagination';
import { RangePicker } from '../../shared/ui/range-picker';
import { SearchInput } from '../../shared/ui/search-input';
import { SortHeader } from '../../shared/ui/sort-header';
import { Toolbar } from '../../shared/ui/toolbar';
import { TimestampPipe } from '../../shared/pipes/timestamp.pipe';

/**
 * Raw log lines from the monitoring core.
 *
 * The lines are already self-describing ("SERVICE ALERT: host;svc;..."),
 * so this does not try to parse them into columns. It colours the leading
 * keyword, which is the part an eye scans for, and leaves the rest
 * verbatim - a log view that reformats its input is a log view you cannot
 * trust against the file on disk.
 */
@Component({
  selector: 'sei-logentries-list',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    TranslocoDirective,
    PageHeader,
    Toolbar,
    SearchInput,
    RangePicker,
    DataTable,
    SortHeader,
    Pagination,
    TimestampPipe,
  ],
  templateUrl: './logentries-list.html',
})
export class LogEntriesList {
  readonly store = new ListStore<LogEntry>({
    path: '/logentries',
    defaultSort: 'entry_time',
    defaultDesc: true,
    filterKeys: ['from', 'to', 'node'],
  });

  setRange(range: { from: string; to: string }): void {
    this.store.setFilter('from', range.from);
    this.store.setFilter('to', range.to);
  }

  /** The tone for a line, from the keyword it opens with. Colour is a
   *  hint on top of the text, never the only signal. */
  tone(line: string): string {
    if (line.includes('CRITICAL') || line.includes('DOWN') || line.includes('UNREACHABLE')) {
      return 'text-critical';
    }
    if (line.includes('WARNING')) {
      return 'text-warning';
    }
    if (line.includes(' OK') || line.startsWith('OK') || line.includes('UP;')) {
      return 'text-ok';
    }
    if (line.includes('UNKNOWN')) {
      return 'text-unknown';
    }
    return 'text-ink-dim';
  }
}
