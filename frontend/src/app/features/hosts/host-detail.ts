import {
  ChangeDetectionStrategy,
  Component,
  computed,
  effect,
  inject,
  signal,
  untracked,
} from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { TranslocoDirective, TranslocoService } from '@jsverse/transloco';
import { Api } from '../../core/api/api.service';
import { Commands } from '../../core/commands/commands.service';
import { Live } from '../../core/events/live.service';
import { ApiError } from '../../core/api/api.error';
import { ListStore } from '../../core/list/list-store';
import type { HostDetail as HostDetailData, ServiceStatus } from '../../core/api/types';
import { DataTable } from '../../shared/ui/data-table';
import { FactList, type Fact } from '../../shared/ui/fact-list';
import { Icon } from '../../shared/ui/icon';
import { PluginOutput } from '../../shared/ui/plugin-output';
import { RowFlags } from '../../shared/ui/row-flags';
import { SortHeader } from '../../shared/ui/sort-header';
import { StateBadge } from '../../shared/ui/state-badge';
import { DurationPipe } from '../../shared/pipes/duration.pipe';
import { SincePipe } from '../../shared/pipes/since.pipe';
import { TimestampPipe } from '../../shared/pipes/timestamp.pipe';
import { stateClass, stateTextClass } from '../../shared/state/state';
import { ObjectActions } from '../commands/object-actions';
import { ObjectContext } from '../commands/object-context';
import { ObjectSettings } from '../commands/object-settings';

/** One host: its current state, what the plugin said, and its services. */
@Component({
  selector: 'sei-host-detail',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    RouterLink,
    TranslocoDirective,
    StateBadge,
    PluginOutput,
    FactList,
    RowFlags,
    DataTable,
    SortHeader,
    Icon,
    ObjectActions,
    ObjectContext,
    ObjectSettings,
    DurationPipe,
    SincePipe,
    TimestampPipe,
  ],
  templateUrl: './host-detail.html',
})
export class HostDetail {
  private readonly api = inject(Api);
  private readonly route = inject(ActivatedRoute);
  private readonly transloco = inject(TranslocoService);

  readonly stateClass = stateClass;
  readonly stateTextClass = stateTextClass;

  readonly hostname = this.route.snapshot.paramMap.get('host') ?? '';

  readonly host = signal<HostDetailData | null>(null);
  readonly error = signal<ApiError | null>(null);
  readonly loading = signal(true);

  readonly services = new ListStore<ServiceStatus>({
    path: `/hosts/${encodeURIComponent(this.route.snapshot.paramMap.get('host') ?? '')}/services`,
    defaultSort: 'severity',
    defaultDesc: true,
    filterKeys: [],
  });

  readonly duration = computed(() => {
    const host = this.host();
    if (!host?.last_state_change) {
      return 0;
    }
    return Math.floor(Date.now() / 1000) - host.last_state_change;
  });

  /** The facts an operator checks when a host looks wrong: is it actually
   *  being checked, how often, and by what. */
  readonly facts = computed<Fact[]>(() => {
    const host = this.host();
    if (!host) {
      return [];
    }
    const t = (key: string) => this.transloco.translate('detail.' + key);

    return [
      { label: t('checkCommand'), value: host.check_command || '—', mono: true },
      { label: t('attempt'), value: `${host.current_check_attempt}/${host.max_check_attempts}` },
      { label: t('checkInterval'), value: this.interval(host.normal_check_interval) },
      { label: t('retryInterval'), value: this.interval(host.retry_check_interval) },
      { label: t('checkPeriod'), value: host.check_timeperiod || '—', mono: true },
      { label: t('latency'), value: `${host.latency.toFixed(3)} s` },
      { label: t('executionTime'), value: `${host.execution_time.toFixed(3)} s` },
      { label: t('lastCheckType'), value: host.is_passive_check ? t('passive') : t('active') },
      { label: t('eventHandler'), value: host.event_handler || '—', mono: true },
      { label: t('node'), value: host.node_name || '—', mono: true },
    ];
  });

  private readonly live = inject(Live);
  private readonly commands = inject(Commands);

  constructor() {
    // Refresh on a live change, a polling tick, or a command this
    // session got confirmed - the last one matters because the event
    // stream may deliver its batch before the core has applied the
    // command, leaving the page showing the state from just before.
    effect(() => {
      this.live.tick();
      this.commands.submitted();
      untracked(() => void this.load());
    });
  }

  async load(): Promise<void> {
    this.loading.set(true);
    this.error.set(null);
    try {
      this.host.set(
        await this.api.get<HostDetailData>(`/hosts/${encodeURIComponent(this.hostname)}`),
      );
    } catch (err) {
      this.error.set(ApiError.from(err));
    } finally {
      this.loading.set(false);
    }
  }

  /** Naemon stores intervals in multiples of interval_length, which is
   *  configurable, so this says "units" rather than inventing minutes we
   *  cannot verify from this database. */
  private interval(value: number | undefined): string {
    if (!value) {
      return '—';
    }
    const key = value === 1 ? 'detail.intervalUnit' : 'detail.intervalUnits';
    return this.transloco.translate(key, { count: value });
  }

  serviceDuration(service: ServiceStatus, now = Math.floor(Date.now() / 1000)): number {
    return service.last_state_change ? now - service.last_state_change : 0;
  }
}
