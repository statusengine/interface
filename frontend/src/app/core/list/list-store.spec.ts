import { Injector, runInInjectionContext, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router } from '@angular/router';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { Api } from '../api/api.service';
import { ApiError } from '../api/api.error';
import { Commands } from '../commands/commands.service';
import { Live } from '../events/live.service';
import { ListStore } from './list-store';

interface Row {
  hostname: string;
}

function setup(options: { queryParams?: Record<string, string>; list?: ReturnType<typeof vi.fn> }) {
  const list =
    options.list ??
    vi
      .fn()
      .mockResolvedValue({ data: [{ hostname: 'a' }], meta: { total: 1, limit: 50, offset: 0 } });
  const navigate = vi.fn().mockResolvedValue(true);

  const params = new Map(Object.entries(options.queryParams ?? {}));

  // The store refreshes itself on a live tick and on a confirmed
  // command; both are stubbed at a fixed value so they never fire here.
  const tick = signal(0);
  const submitted = signal(0);

  TestBed.configureTestingModule({
    providers: [
      { provide: Api, useValue: { list } },
      { provide: Router, useValue: { navigate } },
      { provide: Live, useValue: { tick: tick.asReadonly() } },
      { provide: Commands, useValue: { submitted: submitted.asReadonly() } },
      {
        provide: ActivatedRoute,
        useValue: {
          snapshot: {
            queryParamMap: {
              get: (key: string) => params.get(key) ?? null,
            },
          },
        },
      },
    ],
  });

  const injector = TestBed.inject(Injector);
  const store = runInInjectionContext(
    injector,
    () =>
      new ListStore<Row>({
        path: '/hosts',
        defaultSort: 'severity',
        defaultDesc: true,
        filterKeys: ['state', 'acknowledged'],
        keyOf: (row) => row.hostname,
      }),
  );

  return { store, list, navigate, tick };
}

/** Lets the effect and its fetch settle. */
const settle = () => new Promise((resolve) => setTimeout(resolve, 0));

describe('ListStore', () => {
  beforeEach(() => TestBed.resetTestingModule());

  it('fetches with its defaults', async () => {
    const { store, list } = setup({});
    await settle();

    expect(list).toHaveBeenCalledWith('/hosts', {
      limit: 50,
      offset: 0,
      sort: 'severity:desc',
    });
    expect(store.rows()).toHaveLength(1);
    expect(store.total()).toBe(1);
    expect(store.loading()).toBe(false);
  });

  // Everything that decides what is on screen lives in the URL, so a
  // pasted link still means the same thing an hour later.
  it('restores filters, sorting and paging from the URL', async () => {
    const { store, list } = setup({
      queryParams: {
        state: '1,2',
        acknowledged: 'false',
        q: 'db',
        sort: 'hostname:asc',
        offset: '100',
        limit: '25',
      },
    });
    await settle();

    expect(store.filters()).toEqual({ state: '1,2', acknowledged: 'false', q: 'db' });
    expect(store.sort()).toBe('hostname');
    expect(store.desc()).toBe(false);
    expect(store.offset()).toBe(100);
    expect(list).toHaveBeenCalledWith('/hosts', {
      state: '1,2',
      acknowledged: 'false',
      q: 'db',
      limit: 25,
      offset: 100,
      sort: 'hostname:asc',
    });
  });

  // A query parameter belonging to another page must not be forwarded to
  // the API, which would reject it as an unknown filter.
  it('ignores query parameters it does not know', async () => {
    const { store } = setup({ queryParams: { running: 'true', nonsense: '1' } });
    await settle();

    expect(store.filters()).toEqual({});
  });

  it('resets to the first page when a filter changes', async () => {
    // Staying on page seven of a list that now has two pages shows
    // nothing, which reads like the filter broke.
    const { store } = setup({ queryParams: { offset: '100' } });
    await settle();
    expect(store.offset()).toBe(100);

    store.setFilter('acknowledged', 'true');
    await settle();

    expect(store.offset()).toBe(0);
  });

  it('clears a filter set to an empty value', async () => {
    const { store } = setup({ queryParams: { acknowledged: 'true' } });
    await settle();

    store.setFilter('acknowledged', undefined);
    await settle();

    expect(store.filters()['acknowledged']).toBeUndefined();
    expect(store.hasFilters()).toBe(false);
  });

  it('flips direction on the same column and starts fresh on a new one', async () => {
    const { store } = setup({});
    await settle();

    store.toggleSort('severity');
    expect(store.desc()).toBe(false);

    store.toggleSort('hostname');
    expect(store.sort()).toBe('hostname');
    expect(store.desc()).toBe(false);
  });

  // Without a sequence guard a slow earlier request overwrites a faster
  // later one, and an operator typing quickly sees the results for a
  // prefix of what they typed.
  it('ignores a response that arrives after a newer one', async () => {
    let resolveSlow: (value: unknown) => void = () => {};
    const slow = new Promise((resolve) => {
      resolveSlow = resolve;
    });

    const list = vi
      .fn()
      .mockReturnValueOnce(slow)
      .mockResolvedValue({
        data: [{ hostname: 'fresh' }],
        meta: { total: 1, limit: 50, offset: 0 },
      });

    const { store } = setup({ list });
    // Let the first request actually start before a second one overtakes
    // it; the effect that issues it runs on a microtask.
    await settle();
    expect(store.loading()).toBe(true);

    store.setFilter('acknowledged', 'true');
    await settle();

    expect(store.rows()[0]?.hostname).toBe('fresh');

    resolveSlow({ data: [{ hostname: 'stale' }], meta: { total: 9, limit: 50, offset: 0 } });
    await settle();

    expect(store.rows()[0]?.hostname).toBe('fresh');
    expect(store.total()).toBe(1);
  });

  it('surfaces an error and empties the rows', async () => {
    const list = vi.fn().mockRejectedValue(new ApiError(500, 'internal_error', 'boom'));
    const { store } = setup({ list });
    await settle();

    expect(store.error()?.code).toBe('internal_error');
    expect(store.rows()).toEqual([]);
    expect(store.loading()).toBe(false);
    // An error is not an empty list; the two need different messages.
    expect(store.isEmpty()).toBe(false);
  });

  it('reports an empty result as empty rather than as an error', async () => {
    const list = vi.fn().mockResolvedValue({ data: [], meta: { total: 0, limit: 50, offset: 0 } });
    const { store } = setup({ list });
    await settle();

    expect(store.isEmpty()).toBe(true);
    expect(store.error()).toBeNull();
  });

  // Paging should not bury the previous page under fifty history entries.
  it('replaces the URL rather than pushing onto history', async () => {
    const { store, navigate } = setup({});
    await settle();

    store.setPage(50);
    expect(navigate).toHaveBeenCalled();
    const options = navigate.mock.calls.at(-1)?.[1];
    expect(options.replaceUrl).toBe(true);
    expect(options.queryParams.offset).toBe('50');
  });

  it('refetches when the live stream reports a change', async () => {
    const { store, list, tick } = setup({});
    await settle();
    const before = list.mock.calls.length;

    tick.update((n) => n + 1);
    await settle();

    expect(list.mock.calls.length).toBeGreaterThan(before);
    expect(store.offset()).toBe(0);
  });

  it('omits defaults from the URL so a plain link stays plain', async () => {
    const { store, navigate } = setup({});
    await settle();

    store.setPage(50);
    store.setPage(0);

    const options = navigate.mock.calls.at(-1)?.[1];
    expect(options.queryParams.offset).toBeNull();
    expect(options.queryParams.sort).toBeNull();
  });

  describe('selection', () => {
    const threeRows = {
      data: [{ hostname: 'a' }, { hostname: 'b' }, { hostname: 'c' }],
      meta: { total: 3, limit: 50, offset: 0 },
    };

    it('ticks and unticks one row', async () => {
      const { store } = setup({ list: vi.fn().mockResolvedValue(threeRows) });
      await settle();

      store.toggleRow({ hostname: 'b' });
      expect(store.selectedCount()).toBe(1);
      expect(store.isSelected({ hostname: 'b' })).toBe(true);
      expect(store.selectedRows()).toEqual([{ hostname: 'b' }]);

      store.toggleRow({ hostname: 'b' });
      expect(store.hasSelection()).toBe(false);
    });

    // An operator who ticks the header box has seen these rows, not the
    // three thousand behind them.
    it('selects only the rows on this page', async () => {
      const { store } = setup({ list: vi.fn().mockResolvedValue(threeRows) });
      await settle();

      store.toggleAllOnPage();
      expect(store.selectedCount()).toBe(3);
      expect(store.allOnPageSelected()).toBe(true);

      store.toggleAllOnPage();
      expect(store.selectedCount()).toBe(0);
    });

    // A selection was made against what was on screen. Change what is
    // on screen and it no longer means anything - acting on a stale
    // selection is how the wrong host ends up in a downtime.
    const viewChanges: [string, (s: ListStore<Row>) => void][] = [
      ['a filter', (store) => store.setFilter('acknowledged', 'true')],
      ['a page', (store) => store.setPage(50)],
      ['clearing the filters', (store) => store.clearFilters()],
    ];

    for (const [what, change] of viewChanges) {
      it(`drops the selection when ${what} changes`, async () => {
        const { store } = setup({ list: vi.fn().mockResolvedValue(threeRows) });
        await settle();
        store.toggleAllOnPage();
        expect(store.selectedCount()).toBe(3);

        change(store);
        expect(store.selectedCount()).toBe(0);
      });
    }

    // Not every list has anything worth acting on in bulk, and one
    // without a key must not pretend to have a selection.
    it('has no selection without a key', async () => {
      TestBed.resetTestingModule();
      TestBed.configureTestingModule({
        providers: [
          { provide: Api, useValue: { list: vi.fn().mockResolvedValue(threeRows) } },
          { provide: Router, useValue: { navigate: vi.fn().mockResolvedValue(true) } },
          {
            provide: ActivatedRoute,
            useValue: { snapshot: { queryParamMap: { get: () => null } } },
          },
          { provide: Live, useValue: { tick: signal(0).asReadonly() } },
          { provide: Commands, useValue: { submitted: signal(0).asReadonly() } },
        ],
      });
      const store = runInInjectionContext(
        TestBed.inject(Injector),
        () =>
          new ListStore<Row>({
            path: '/logentries',
            defaultSort: 'entry_time',
            filterKeys: [],
          }),
      );
      await settle();

      store.toggleRow({ hostname: 'a' });
      store.toggleAllOnPage();

      expect(store.selectedCount()).toBe(0);
      expect(store.allOnPageSelected()).toBe(false);
      expect(store.selectedRows()).toEqual([]);
    });
  });
});
