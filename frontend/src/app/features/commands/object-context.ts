import { ChangeDetectionStrategy, Component, computed, inject, input, output } from '@angular/core';
import { TranslocoDirective, TranslocoService } from '@jsverse/transloco';
import { Auth } from '../../core/auth/auth.service';
import { Commands } from '../../core/commands/commands.service';
import { ServerInfo } from '../../core/meta/meta.service';
import type { Acknowledgement, Downtime, Kind } from '../../core/api/types';
import { Icon } from '../../shared/ui/icon';
import { DurationPipe } from '../../shared/pipes/duration.pipe';
import { SincePipe } from '../../shared/pipes/since.pipe';
import { TimestampPipe } from '../../shared/pipes/timestamp.pipe';

/**
 * Why this object is quiet.
 *
 * A downtime and an acknowledgement both suppress notifications, and
 * both are decisions somebody made for a reason. Reduced to an icon in
 * the header, that reason is invisible - and the reason is the part an
 * operator needs: who took it, what they said, and how long it lasts.
 * So each gets a panel with the record behind it and the way to undo it.
 */
@Component({
  selector: 'sei-object-context',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective, Icon, DurationPipe, SincePipe, TimestampPipe],
  templateUrl: './object-context.html',
})
export class ObjectContext {
  private readonly auth = inject(Auth);
  private readonly commands = inject(Commands);
  private readonly transloco = inject(TranslocoService);
  private readonly server = inject(ServerInfo);

  readonly kind = input.required<Kind>();
  readonly host = input.required<string>();
  readonly service = input<string | undefined>(undefined);
  readonly downtimes = input<Downtime[]>([]);
  /**
   * Whether the core currently considers this acknowledged. Separate
   * from the record below: the worker can miss an acknowledgement event
   * while the status still says one applies, and the way to undo it
   * must not disappear along with the record.
   */
  readonly acknowledged = input(false);
  readonly acknowledgement = input<Acknowledgement | undefined>(undefined);

  /** Emitted after a command is accepted, so the page can reload. */
  readonly changed = output<void>();

  readonly busy = this.commands.busy;
  readonly commandsEnabled = computed(() => this.server.commandsEnabled);
  readonly canDowntime = computed(() => this.auth.can('commands:downtime'));
  readonly canAcknowledge = computed(() => this.auth.can('commands:acknowledge'));

  readonly hasAnything = computed(() => this.downtimes().length > 0 || this.acknowledged());

  /** Seconds until a window ends, or 0 once it has. */
  remaining(downtime: Downtime, now = Math.floor(Date.now() / 1000)): number {
    return Math.max(0, downtime.scheduled_end_time - now);
  }

  /** A window that has not begun yet is scheduled, not running. */
  isRunning(downtime: Downtime, now = Math.floor(Date.now() / 1000)): boolean {
    return (
      downtime.was_started &&
      downtime.scheduled_start_time <= now &&
      now < downtime.scheduled_end_time
    );
  }

  private target() {
    return { kind: this.kind(), host: this.host(), service: this.service() };
  }

  private label(): string {
    const service = this.service();
    return service ? `${this.host()} / ${service}` : this.host();
  }

  async cancelDowntime(downtime: Downtime): Promise<void> {
    // With several windows overlapping, cancelling one leaves the object
    // in a downtime, so there is nothing observable to confirm against.
    // Only the last one produces a visible change.
    const last = this.downtimes().length === 1;

    const ok = await this.commands.run({
      action: 'downtime/delete',
      targets: [this.target()],
      body: { internal_id: downtime.internal_id },
      pending: this.transloco.translate('context.cancellingDowntime', { target: this.label() }),
      success: this.transloco.translate('context.downtimeCancelled', { target: this.label() }),
      verify: last ? (status) => !status.in_downtime : undefined,
    });
    if (ok) {
      this.changed.emit();
    }
  }

  async removeAcknowledgement(): Promise<void> {
    const ok = await this.commands.run({
      action: 'remove-acknowledgement',
      targets: [this.target()],
      body: {},
      pending: this.transloco.translate('commands.removingAck', { target: this.label() }),
      success: this.transloco.translate('commands.removedAck', { target: this.label() }),
      verify: (status) => !status.acknowledged,
    });
    if (ok) {
      this.changed.emit();
    }
  }
}
