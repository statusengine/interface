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
import type { ServiceStatus } from '../../core/api/types';
import { FactList, type Fact } from '../../shared/ui/fact-list';
import { PluginOutput } from '../../shared/ui/plugin-output';
import { RowFlags } from '../../shared/ui/row-flags';
import { StateBadge } from '../../shared/ui/state-badge';
import { MetricsPanel } from './metrics-panel';
import { ObjectActions } from '../commands/object-actions';
import { DurationPipe } from '../../shared/pipes/duration.pipe';
import { SincePipe } from '../../shared/pipes/since.pipe';
import { TimestampPipe } from '../../shared/pipes/timestamp.pipe';
import { stateClass } from '../../shared/state/state';

/**
 * One service.
 *
 * Identified by query parameters rather than path segments: a Naemon
 * service description is free text, and this installation has one called
 * `C:\ Drive Space`. Percent-encoded slashes in a path are a fight with
 * every proxy in between for no benefit.
 */
@Component({
  selector: 'sei-service-detail',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    RouterLink,
    TranslocoDirective,
    StateBadge,
    PluginOutput,
    FactList,
    RowFlags,
    MetricsPanel,
    ObjectActions,
    DurationPipe,
    SincePipe,
    TimestampPipe,
  ],
  templateUrl: './service-detail.html',
})
export class ServiceDetail {
  private readonly api = inject(Api);
  private readonly route = inject(ActivatedRoute);
  private readonly transloco = inject(TranslocoService);

  readonly stateClass = stateClass;

  readonly hostname = this.route.snapshot.queryParamMap.get('host') ?? '';
  readonly description = this.route.snapshot.queryParamMap.get('service') ?? '';

  readonly service = signal<ServiceStatus | null>(null);
  readonly error = signal<ApiError | null>(null);
  readonly loading = signal(true);

  readonly duration = computed(() => {
    const service = this.service();
    if (!service?.last_state_change) {
      return 0;
    }
    return Math.floor(Date.now() / 1000) - service.last_state_change;
  });

  readonly facts = computed<Fact[]>(() => {
    const service = this.service();
    if (!service) {
      return [];
    }
    const t = (key: string) => this.transloco.translate('detail.' + key);
    const yes = this.transloco.translate('filters.yes');
    const no = this.transloco.translate('filters.no');

    return [
      { label: t('checkCommand'), value: service.check_command || '—', mono: true },
      {
        label: t('attempt'),
        value: `${service.current_check_attempt}/${service.max_check_attempts}`,
      },
      { label: t('checkInterval'), value: this.interval(service.normal_check_interval) },
      { label: t('retryInterval'), value: this.interval(service.retry_check_interval) },
      { label: t('checkPeriod'), value: service.check_timeperiod || '—', mono: true },
      { label: t('latency'), value: `${service.latency.toFixed(3)} s` },
      { label: t('executionTime'), value: `${service.execution_time.toFixed(3)} s` },
      { label: t('lastCheckType'), value: service.is_passive_check ? t('passive') : t('active') },
      { label: t('activeChecks'), value: service.active_checks_enabled ? yes : no },
      { label: t('passiveChecks'), value: service.passive_checks_enabled ? yes : no },
      { label: t('notifications'), value: service.notifications_enabled ? yes : no },
      { label: t('flapDetection'), value: service.flap_detection_enabled ? yes : no },
      { label: t('eventHandler'), value: service.event_handler || '—', mono: true },
      { label: t('node'), value: service.node_name || '—', mono: true },
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
      this.service.set(
        await this.api.get<ServiceStatus>('/service', {
          host: this.hostname,
          service: this.description,
        }),
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
}
