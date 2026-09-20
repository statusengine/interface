import { DestroyRef, computed, effect, inject, signal, untracked } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { TranslocoService } from '@jsverse/transloco';
import { ActivatedRoute, Router } from '@angular/router';
import { Api, type ListQuery } from '../api/api.service';
import { ApiError } from '../api/api.error';
import { apiErrorText } from '../api/error-text';
import { Commands } from '../commands/commands.service';
import { Live } from '../events/live.service';
import type { ListMeta } from '../api/types';

/** A value a filter can hold. Undefined means the filter is off. */
export type FilterValue = string | number | boolean | undefined;

export interface ListStoreOptions<T> {
  /** API path, without /api/v1. */
  path: string;
  /** Column the list sorts by when nothing is chosen. */
  defaultSort: string;
  defaultDesc?: boolean;
  defaultLimit?: number;
  /** Filter keys this list understands, so a stray query parameter from
   *  another page is not forwarded to the API and rejected. */
  filterKeys: readonly string[];

  /**
   * How to identify a row, for selection. Without it the list has no
   * selection: not every list has anything worth acting on in bulk.
   */
  keyOf?: (row: T) => string;
}

/**
 * The state behind one list page: filters, sorting, paging and the rows.
 *
 * Everything that decides what is on screen lives in the URL. An operator
 * who has narrowed a list down to the three hosts they care about can
 * paste that link into a ticket, and it still means the same thing an
 * hour later. It also means the back button does what people expect.
 */
export class ListStore<T> {
  private readonly api = inject(Api);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly destroyRef = inject(DestroyRef);
  private readonly live = inject(Live);
  private readonly commands = inject(Commands);
  private readonly transloco = inject(TranslocoService);

  private readonly _rows = signal<T[]>([]);
  private readonly _meta = signal<ListMeta>({ total: 0, limit: 0, offset: 0 });
  private readonly _loading = signal(false);
  private readonly _error = signal<ApiError | null>(null);

  readonly rows = this._rows.asReadonly();
  readonly meta = this._meta.asReadonly();
  readonly loading = this._loading.asReadonly();
  readonly error = this._error.asReadonly();

  /** The failure in the reader's language, falling back to the server's
   *  own sentence when this build has no wording for the code. */
  readonly errorText = computed(() => {
    // Read the active language so switching it re-translates what is
    // already on screen.
    this.language();
    return apiErrorText(this.transloco, this._error());
  });

  private readonly language = toSignal(this.transloco.langChanges$, {
    initialValue: this.transloco.getActiveLang(),
  });

  readonly total = computed(() => this._meta().total);
  readonly isEmpty = computed(
    () => !this._loading() && !this._error() && this._rows().length === 0,
  );

  private readonly _filters = signal<Record<string, FilterValue>>({});
  private readonly _sort = signal('');
  private readonly _desc = signal(false);
  private readonly _offset = signal(0);
  private readonly _limit = signal(50);

  readonly filters = this._filters.asReadonly();
  readonly sort = this._sort.asReadonly();
  readonly desc = this._desc.asReadonly();
  readonly offset = this._offset.asReadonly();
  readonly limit = this._limit.asReadonly();

  /** True when anything narrows the list, so the UI can offer "clear". */
  readonly hasFilters = computed(() =>
    Object.values(this._filters()).some((v) => v !== undefined && v !== ''),
  );

  private searchDebounce?: ReturnType<typeof setTimeout>;
  private requestSeq = 0;

  private readonly _selected = signal<ReadonlySet<string>>(new Set());

  /** The keys of the selected rows. */
  readonly selected = this._selected.asReadonly();
  readonly selectedCount = computed(() => this._selected().size);
  readonly hasSelection = computed(() => this._selected().size > 0);

  /** The selected rows themselves, in the order they appear. */
  readonly selectedRows = computed(() => {
    const keyOf = this.options.keyOf;
    if (!keyOf) {
      return [] as T[];
    }
    const keys = this._selected();
    return this._rows().filter((row) => keys.has(keyOf(row)));
  });

  /** True when every row on this page is selected. */
  readonly allOnPageSelected = computed(() => {
    const keyOf = this.options.keyOf;
    const rows = this._rows();
    if (!keyOf || rows.length === 0) {
      return false;
    }
    const keys = this._selected();
    return rows.every((row) => keys.has(keyOf(row)));
  });

  constructor(private readonly options: ListStoreOptions<T>) {
    this._sort.set(options.defaultSort);
    this._desc.set(options.defaultDesc ?? false);
    this._limit.set(options.defaultLimit ?? 50);

    this.readFromUrl();

    // Refetch whenever anything that shapes the request changes. Reading
    // the signals here is what subscribes this effect to them.
    effect(() => {
      this._filters();
      this._sort();
      this._desc();
      this._offset();
      this._limit();
      untracked(() => void this.fetch());
    });

    // Refresh when the event stream says something moved, when the
    // polling fallback ticks, or when a command this session submitted
    // was confirmed. The list does not care which: all three mean
    // "what you are showing may be out of date".
    effect(() => {
      this.live.tick();
      this.commands.submitted();
      untracked(() => {
        if (!this._loading()) {
          void this.fetch();
        }
      });
    });

    this.destroyRef.onDestroy(() => clearTimeout(this.searchDebounce));
  }

  /** Set a filter. Any change resets to the first page: staying on page
   *  seven of a list that now has two pages shows nothing, which reads
   *  like the filter broke. */
  setFilter(key: string, value: FilterValue): void {
    const next = { ...this._filters() };
    if (value === undefined || value === '' || value === false) {
      delete next[key];
    } else {
      next[key] = value;
    }
    this._filters.set(next);
    this._offset.set(0);
    // A selection was made against what was on screen. Change what is on
    // screen and it no longer means anything.
    this.clearSelection();
    this.writeToUrl();
  }

  /** Set the free-text search, debounced so a fast typist does not
   *  produce one request per keystroke. */
  setSearch(value: string): void {
    clearTimeout(this.searchDebounce);
    this.searchDebounce = setTimeout(() => this.setFilter('q', value.trim() || undefined), 250);
  }

  clearFilters(): void {
    this._filters.set({});
    this._offset.set(0);
    this.clearSelection();
    this.writeToUrl();
  }

  /** Clicking a column header: same column flips direction, a new column
   *  starts ascending. */
  toggleSort(column: string): void {
    if (this._sort() === column) {
      this._desc.update((d) => !d);
    } else {
      this._sort.set(column);
      this._desc.set(false);
    }
    this._offset.set(0);
    this.writeToUrl();
  }

  setPage(offset: number): void {
    this._offset.set(Math.max(0, offset));
    this.clearSelection();
    this.writeToUrl();
  }

  setLimit(limit: number): void {
    this._limit.set(limit);
    this._offset.set(0);
    this.writeToUrl();
  }

  /** Re-run the current request, for a manual refresh or after a command. */
  async reload(): Promise<void> {
    await this.fetch();
  }

  keyFor(row: T): string {
    return this.options.keyOf ? this.options.keyOf(row) : '';
  }

  isSelected(row: T): boolean {
    return this._selected().has(this.keyFor(row));
  }

  toggleRow(row: T): void {
    const key = this.keyFor(row);
    if (!key) {
      return;
    }
    const next = new Set(this._selected());
    if (!next.delete(key)) {
      next.add(key);
    }
    this._selected.set(next);
  }

  /** Selects or clears every row on the current page.
   *
   *  Only this page: an operator who ticks the header box has seen these
   *  fifty rows, not the three thousand behind them, and acting on what
   *  they cannot see is not what they asked for. */
  toggleAllOnPage(): void {
    const keyOf = this.options.keyOf;
    if (!keyOf) {
      return;
    }
    const next = new Set(this._selected());
    if (this.allOnPageSelected()) {
      for (const row of this._rows()) {
        next.delete(keyOf(row));
      }
    } else {
      for (const row of this._rows()) {
        next.add(keyOf(row));
      }
    }
    this._selected.set(next);
  }

  clearSelection(): void {
    this._selected.set(new Set());
  }

  private async fetch(): Promise<void> {
    const seq = ++this.requestSeq;
    this._loading.set(true);
    this._error.set(null);

    const query: ListQuery = {
      ...this._filters(),
      limit: this._limit(),
      offset: this._offset(),
      sort: `${this._sort()}:${this._desc() ? 'desc' : 'asc'}`,
    };

    try {
      const res = await this.api.list<T>(this.options.path, query);
      // A slower earlier request must not overwrite a faster later one.
      // Without this an operator typing quickly sees the results of a
      // prefix of what they typed.
      if (seq !== this.requestSeq) {
        return;
      }
      this._rows.set(res.data);
      this._meta.set(res.meta);
    } catch (err) {
      if (seq !== this.requestSeq) {
        return;
      }
      this._error.set(ApiError.from(err));
      this._rows.set([]);
    } finally {
      if (seq === this.requestSeq) {
        this._loading.set(false);
      }
    }
  }

  private readFromUrl(): void {
    const params = this.route.snapshot.queryParamMap;

    const filters: Record<string, FilterValue> = {};
    for (const key of [...this.options.filterKeys, 'q']) {
      const value = params.get(key);
      if (value !== null && value !== '') {
        filters[key] = value;
      }
    }
    this._filters.set(filters);

    const sort = params.get('sort');
    if (sort) {
      const [column, direction] = sort.split(':');
      this._sort.set(column || this.options.defaultSort);
      this._desc.set(direction === 'desc');
    }

    const offset = Number(params.get('offset'));
    if (Number.isInteger(offset) && offset >= 0) {
      this._offset.set(offset);
    }
    const limit = Number(params.get('limit'));
    if (Number.isInteger(limit) && limit > 0) {
      this._limit.set(limit);
    }
  }

  private writeToUrl(): void {
    const queryParams: Record<string, string | null> = {};
    for (const key of [...this.options.filterKeys, 'q']) {
      const value = this._filters()[key];
      queryParams[key] = value === undefined ? null : String(value);
    }
    queryParams['sort'] =
      this._sort() === this.options.defaultSort &&
      this._desc() === (this.options.defaultDesc ?? false)
        ? null
        : `${this._sort()}:${this._desc() ? 'desc' : 'asc'}`;
    queryParams['offset'] = this._offset() === 0 ? null : String(this._offset());
    queryParams['limit'] =
      this._limit() === (this.options.defaultLimit ?? 50) ? null : String(this._limit());

    // replaceUrl, so paging through a list does not bury the previous
    // page under fifty history entries.
    void this.router.navigate([], {
      relativeTo: this.route,
      queryParams,
      queryParamsHandling: 'merge',
      replaceUrl: true,
    });
  }
}
