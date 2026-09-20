import {
  ChangeDetectionStrategy,
  Component,
  computed,
  inject,
  input,
  output,
  signal,
} from '@angular/core';
import { TranslocoDirective, TranslocoService } from '@jsverse/transloco';
import { Auth } from '../../core/auth/auth.service';
import { Commands, type CommandTarget } from '../../core/commands/commands.service';
import { ServerInfo } from '../../core/meta/meta.service';
import { Dialog } from '../../shared/ui/dialog';
import { Icon } from '../../shared/ui/icon';
import { AcknowledgeFields, DowntimeFields } from './command-fields';

type Mode = 'none' | 'acknowledge' | 'downtime';

/**
 * Acting on everything an operator has ticked.
 *
 * Three actions, not the seven the single-object bar offers. Toggling
 * notifications across a selection where some are on and some off has
 * no obvious meaning, and a passive result is a statement about one
 * check. Downtime, acknowledge and check-now are the three that read
 * the same whether they apply to one object or fifty.
 */
@Component({
  selector: 'sei-bulk-actions',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective, Dialog, Icon, AcknowledgeFields, DowntimeFields],
  template: `
    <ng-container *transloco="let t">
      @if (targets().length > 0 && anyAction()) {
        <div
          class="sticky bottom-0 z-20 flex flex-wrap items-center gap-2 border-t border-line bg-raised px-4 py-2.5 shadow-[var(--shadow-overlay)] sm:px-6"
          role="region"
          [attr.aria-label]="t('bulk.region')"
        >
          <span class="text-[13px] font-medium">
            {{ t('bulk.selected', { count: targets().length }) }}
          </span>

          @if (!commandsEnabled()) {
            <span class="flex items-center gap-1.5 text-[12px] text-ink-dim">
              <sei-icon name="alert" [size]="13" />
              {{ t('commands.disabled') }}
            </span>
          } @else {
            @if (canReschedule()) {
              <button type="button" (click)="checkNow()" [disabled]="busy()" class="sei-action">
                <sei-icon name="refresh" [size]="14" />
                {{ t('commands.checkNow') }}
              </button>
            }
            @if (canAcknowledge()) {
              <button
                type="button"
                (click)="open('acknowledge')"
                [disabled]="busy()"
                class="sei-action"
              >
                <sei-icon name="acknowledge" [size]="14" />
                {{ t('commands.acknowledge') }}
              </button>
            }
            @if (canDowntime()) {
              <button
                type="button"
                (click)="open('downtime')"
                [disabled]="busy()"
                class="sei-action"
              >
                <sei-icon name="downtime" [size]="14" />
                {{ t('commands.downtime') }}
              </button>
            }
          }

          <button type="button" (click)="clear.emit()" class="ml-auto sei-action">
            {{ t('bulk.clear') }}
          </button>
        </div>
      }

      <sei-dialog
        [open]="mode() !== 'none'"
        [title]="mode() === 'none' ? '' : t('commands.' + mode() + 'Title')"
        [subtitle]="subtitle()"
        (closed)="close()"
      >
        @switch (mode()) {
          @case ('acknowledge') {
            <sei-acknowledge-fields
              id="bulk-ack"
              [(comment)]="comment"
              [(sticky)]="sticky"
              [(notify)]="notifyContacts"
              [(persistent)]="persistent"
            />
          }
          @case ('downtime') {
            <sei-downtime-fields
              id="bulk-downtime"
              [(minutes)]="downtimeMinutes"
              [(comment)]="comment"
              [(allServices)]="downtimeAllServices"
              [offerAllServices]="hasHostTarget()"
            />
          }
        }

        <button footer type="button" (click)="close()" class="sei-action">
          {{ t('commands.cancel') }}
        </button>
        @switch (mode()) {
          @case ('acknowledge') {
            <button
              footer
              type="button"
              (click)="acknowledge()"
              [disabled]="busy() || !comment().trim()"
              class="sei-action-primary"
            >
              {{ t('bulk.applyTo', { count: targets().length }) }}
            </button>
          }
          @case ('downtime') {
            <button
              footer
              type="button"
              (click)="scheduleDowntime()"
              [disabled]="busy() || !comment().trim()"
              class="sei-action-primary"
            >
              {{ t('bulk.applyTo', { count: targets().length }) }}
            </button>
          }
        }
      </sei-dialog>
    </ng-container>
  `,
})
export class BulkActions {
  private readonly auth = inject(Auth);
  private readonly commands = inject(Commands);
  private readonly transloco = inject(TranslocoService);
  private readonly server = inject(ServerInfo);

  readonly targets = input.required<CommandTarget[]>();

  /** Emitted once a command has been accepted, so the list can drop the
   *  selection it was made against. */
  readonly applied = output<void>();
  readonly clear = output<void>();

  readonly mode = signal<Mode>('none');
  readonly busy = this.commands.busy;

  readonly comment = signal('');
  readonly sticky = signal(true);
  readonly notifyContacts = signal(false);
  readonly persistent = signal(false);
  readonly downtimeMinutes = signal(60);
  readonly downtimeAllServices = signal(false);

  readonly commandsEnabled = computed(() => this.server.commandsEnabled);
  readonly canAcknowledge = computed(() => this.auth.can('commands:acknowledge'));
  readonly canDowntime = computed(() => this.auth.can('commands:downtime'));
  readonly canReschedule = computed(() => this.auth.can('commands:reschedule'));
  readonly anyAction = computed(
    () => this.canAcknowledge() || this.canDowntime() || this.canReschedule(),
  );

  readonly hasHostTarget = computed(() => this.targets().some((t) => t.kind === 'host'));

  /** Names the objects while they fit, then counts them. */
  readonly subtitle = computed(() => {
    const targets = this.targets();
    const names = targets.map((t) => (t.service ? `${t.host} / ${t.service}` : t.host));
    if (names.length <= 3) {
      return names.join(', ');
    }
    return this.transloco.translate('bulk.andMore', {
      names: names.slice(0, 2).join(', '),
      count: names.length - 2,
    });
  });

  open(mode: Mode): void {
    this.comment.set('');
    this.sticky.set(true);
    this.notifyContacts.set(false);
    this.persistent.set(false);
    this.downtimeMinutes.set(60);
    this.downtimeAllServices.set(false);
    this.mode.set(mode);
  }

  close(): void {
    this.mode.set('none');
  }

  private t(key: string, params?: Record<string, unknown>): string {
    return this.transloco.translate(key, params);
  }

  /** Bulk gets its own wording. Feeding "3 objects" into a message
   *  written for one object produces "3 objects is in a downtime". */
  private count(): Record<string, unknown> {
    return { count: this.targets().length };
  }

  async checkNow(): Promise<void> {
    const ok = await this.commands.run({
      action: 'reschedule',
      targets: this.targets(),
      body: { forced: true },
      pending: this.t('bulk.rescheduling', this.count()),
      success: this.t('bulk.rescheduled', this.count()),
    });
    if (ok) {
      this.applied.emit();
    }
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
      pending: this.t('bulk.acknowledging', this.count()),
      success: this.t('bulk.acknowledged', this.count()),
      verify: (status) => status.acknowledged,
    });
    if (ok) {
      this.close();
      this.applied.emit();
    }
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
        // The server applies this only to host targets, so a mixed
        // selection does not need splitting here.
        all_services: this.hasHostTarget() ? this.downtimeAllServices() : undefined,
      },
      pending: this.t('bulk.schedulingDowntime', this.count()),
      success: this.t('bulk.downtimeScheduled', this.count()),
      verify: (status) => status.in_downtime,
    });
    if (ok) {
      this.close();
      this.applied.emit();
    }
  }
}
