import { ChangeDetectionStrategy, Component, computed, inject, input, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TranslocoDirective, TranslocoService } from '@jsverse/transloco';
import { Auth } from '../../core/auth/auth.service';
import { Commands, type CommandTarget } from '../../core/commands/commands.service';
import { ServerInfo } from '../../core/meta/meta.service';
import type { HostStatus, Kind, ServiceStatus } from '../../core/api/types';
import { Dialog } from '../../shared/ui/dialog';
import { Icon } from '../../shared/ui/icon';
import { AcknowledgeFields, DOWNTIME_PRESETS, DowntimeFields } from './command-fields';

/** Which form the one dialog is currently showing. */
type Mode = 'none' | 'acknowledge' | 'downtime' | 'result' | 'notify';

/**
 * What an operator can do to one host or service.
 *
 * Every button is gated on the permission the server enforces anyway.
 * Hiding a control the caller may not use is a courtesy; the refusal in
 * the API is the control.
 */
@Component({
  selector: 'sei-object-actions',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormsModule, TranslocoDirective, Dialog, Icon, AcknowledgeFields, DowntimeFields],
  templateUrl: './object-actions.html',
})
export class ObjectActions {
  private readonly auth = inject(Auth);
  private readonly commands = inject(Commands);
  private readonly transloco = inject(TranslocoService);
  readonly server = inject(ServerInfo);

  readonly kind = input.required<Kind>();
  readonly status = input.required<HostStatus | ServiceStatus>();

  readonly mode = signal<Mode>('none');
  readonly busy = this.commands.busy;

  readonly downtimePresets = DOWNTIME_PRESETS;

  // Form state. Reset each time a dialog opens, so a dismissed comment
  // does not reappear on the next object.
  readonly comment = signal('');
  readonly sticky = signal(true);
  readonly notifyContacts = signal(false);
  readonly persistent = signal(false);
  readonly downtimeMinutes = signal(60);
  readonly downtimeAllServices = signal(false);
  readonly resultCode = signal(0);
  readonly resultOutput = signal('');
  readonly resultPerfdata = signal('');
  readonly notifyForced = signal(false);
  readonly notifyBroadcast = signal(false);

  readonly host = computed(() => this.status().hostname);
  readonly service = computed(() =>
    this.kind() === 'service' ? (this.status() as ServiceStatus).service_description : undefined,
  );
  readonly label = computed(() => {
    const service = this.service();
    return service ? `${this.host()} / ${service}` : this.host();
  });

  readonly commandsEnabled = computed(() => this.server.commandsEnabled);
  readonly canAcknowledge = computed(() => this.auth.can('commands:acknowledge'));
  readonly canDowntime = computed(() => this.auth.can('commands:downtime'));
  readonly canReschedule = computed(() => this.auth.can('commands:reschedule'));
  readonly canSubmitResult = computed(() => this.auth.can('commands:passiveresult'));
  readonly canNotify = computed(() => this.auth.can('commands:notification'));
  readonly canToggle = computed(() => this.auth.can('commands:toggle'));

  readonly anyAction = computed(
    () =>
      this.canAcknowledge() ||
      this.canDowntime() ||
      this.canReschedule() ||
      this.canSubmitResult() ||
      this.canNotify() ||
      this.canToggle(),
  );

  /** A host can be UP, DOWN or UNREACHABLE; a service adds UNKNOWN. */
  readonly resultCodes = computed(() =>
    this.kind() === 'host'
      ? [
          { code: 0, key: 'up' },
          { code: 1, key: 'down' },
          { code: 2, key: 'unreachable' },
        ]
      : [
          { code: 0, key: 'ok' },
          { code: 1, key: 'warning' },
          { code: 2, key: 'critical' },
          { code: 3, key: 'unknown' },
        ],
  );

  /** Acknowledging something that is already OK does nothing. */
  readonly canBeAcknowledged = computed(
    () => this.status().state !== 0 && !this.status().acknowledged,
  );

  /** This object, as the single-element list every command takes. */
  private targets(): CommandTarget[] {
    return [{ kind: this.kind(), host: this.host(), service: this.service() }];
  }

  open(mode: Mode): void {
    this.comment.set('');
    this.sticky.set(true);
    this.notifyContacts.set(false);
    this.persistent.set(false);
    this.downtimeMinutes.set(60);
    this.downtimeAllServices.set(false);
    this.resultCode.set(this.status().state);
    this.resultOutput.set('');
    this.resultPerfdata.set('');
    this.notifyForced.set(false);
    this.notifyBroadcast.set(false);
    this.mode.set(mode);
  }

  close(): void {
    this.mode.set('none');
  }

  private t(key: string, params?: Record<string, unknown>): string {
    return this.transloco.translate(key, params);
  }

  async checkNow(): Promise<void> {
    const before = this.status().last_check;
    await this.commands.run({
      action: 'reschedule',
      targets: this.targets(),
      body: { forced: true },
      pending: this.t('commands.rescheduling', { target: this.label() }),
      success: this.t('commands.rescheduled', { target: this.label() }),
      // A forced check has landed once the core records a newer one.
      verify: (status) => status.last_check > before,
    });
  }

  async acknowledge(): Promise<void> {
    const ok = await this.commands.run({
      action: 'acknowledge',
      targets: this.targets(),
      body: {
        comment: this.comment(),
        sticky: this.sticky(),
        notify: this.notifyContacts(),
        persistent: this.persistent(),
      },
      pending: this.t('commands.acknowledging', { target: this.label() }),
      success: this.t('commands.acknowledged', { target: this.label() }),
      verify: (status) => status.acknowledged,
    });
    if (ok) {
      this.close();
    }
  }

  async removeAcknowledgement(): Promise<void> {
    await this.commands.run({
      action: 'remove-acknowledgement',
      targets: this.targets(),
      body: {},
      pending: this.t('commands.removingAck', { target: this.label() }),
      success: this.t('commands.removedAck', { target: this.label() }),
      verify: (status) => !status.acknowledged,
    });
  }

  async scheduleDowntime(): Promise<void> {
    const start = Math.floor(Date.now() / 1000);
    const end = start + this.downtimeMinutes() * 60;

    const ok = await this.commands.run({
      action: 'downtime',
      targets: this.targets(),
      body: {
        start,
        end,
        fixed: true,
        comment: this.comment(),
        all_services: this.kind() === 'host' ? this.downtimeAllServices() : undefined,
      },
      pending: this.t('commands.schedulingDowntime', { target: this.label() }),
      success: this.t('commands.downtimeScheduled', { target: this.label() }),
      verify: (status) => status.in_downtime,
    });
    if (ok) {
      this.close();
    }
  }

  async submitResult(): Promise<void> {
    const before = this.status().last_check;
    const ok = await this.commands.run({
      action: 'submit-result',
      targets: this.targets(),
      body: {
        return_code: this.resultCode(),
        output: this.resultOutput(),
        perf_data: this.resultPerfdata() || undefined,
      },
      pending: this.t('commands.submittingResult', { target: this.label() }),
      success: this.t('commands.resultSubmitted', { target: this.label() }),
      verify: (status) => status.last_check > before,
    });
    if (ok) {
      this.close();
    }
  }

  async sendNotification(): Promise<void> {
    const ok = await this.commands.run({
      action: 'notify',
      targets: this.targets(),
      body: {
        comment: this.comment(),
        forced: this.notifyForced(),
        broadcast: this.notifyBroadcast(),
      },
      pending: this.t('commands.notifying', { target: this.label() }),
      // Nothing observable changes on the object, so this settles at
      // "submitted" rather than claiming a delivery it cannot see.
      success: this.t('commands.notified', { target: this.label() }),
    });
    if (ok) {
      this.close();
    }
  }

  async toggleNotifications(): Promise<void> {
    const enable = !this.status().notifications_enabled;
    await this.commands.run({
      action: 'toggle-notifications',
      targets: this.targets(),
      body: { enable },
      pending: this.t(
        enable ? 'commands.enablingNotifications' : 'commands.disablingNotifications',
        {
          target: this.label(),
        },
      ),
      success: this.t(enable ? 'commands.notificationsEnabled' : 'commands.notificationsDisabled', {
        target: this.label(),
      }),
      verify: (status) => status.notifications_enabled === enable,
    });
  }

  async toggleActiveChecks(): Promise<void> {
    const enable = !this.status().active_checks_enabled;
    await this.commands.run({
      action: 'toggle-active-checks',
      targets: this.targets(),
      body: { enable },
      pending: this.t(enable ? 'commands.enablingChecks' : 'commands.disablingChecks', {
        target: this.label(),
      }),
      success: this.t(enable ? 'commands.checksEnabled' : 'commands.checksDisabled', {
        target: this.label(),
      }),
      verify: (status) => status.active_checks_enabled === enable,
    });
  }
}
