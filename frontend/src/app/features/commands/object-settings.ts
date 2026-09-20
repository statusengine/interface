import { ChangeDetectionStrategy, Component, computed, inject, input, output } from '@angular/core';
import { TranslocoDirective, TranslocoService } from '@jsverse/transloco';
import { Auth } from '../../core/auth/auth.service';
import { Commands } from '../../core/commands/commands.service';
import { ServerInfo } from '../../core/meta/meta.service';
import type { HostStatus, Kind, ServiceStatus } from '../../core/api/types';
import { Icon } from '../../shared/ui/icon';

/** The settings Naemon lets an external command change per object. */
type SwitchName =
  'active_checks' | 'passive_checks' | 'notifications' | 'flap_detection' | 'event_handler';

interface Setting {
  name: SwitchName;
  /** Translation key under `settings.`. */
  label: string;
  on: boolean;
}

/**
 * The five per-object settings, as controls rather than readings.
 *
 * Each one is a real switch in the monitoring core with an
 * `ENABLE_`/`DISABLE_` command behind it, and an operator looking at
 * "Notifications: no" during an incident wants to change it there, not
 * go and find the command.
 *
 * Only the five the status tables report. Naemon has more - freshness
 * checks, obsessing - and a control for a value the page cannot show
 * afterwards is a control nobody can trust.
 */
@Component({
  selector: 'sei-object-settings',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective, Icon],
  template: `
    <section class="rounded-md border border-line bg-surface p-4" *transloco="let t">
      <div class="mb-3 flex flex-wrap items-baseline justify-between gap-2">
        <h2 class="text-[13px] font-medium">{{ t('settings.title') }}</h2>
        @if (!editable()) {
          <p class="flex items-center gap-1.5 text-[12px] text-ink-dim">
            <sei-icon name="alert" [size]="13" />
            {{ commandsEnabled() ? t('settings.needsPermission') : t('commands.disabled') }}
          </p>
        }
      </div>

      <dl class="grid grid-cols-2 gap-x-6 gap-y-3 sm:grid-cols-3 xl:grid-cols-5">
        @for (setting of settings(); track setting.name) {
          <div class="min-w-0">
            <dt class="text-[11px] text-ink-dim" [id]="labelId(setting)">
              {{ t('settings.' + setting.label) }}
            </dt>
            <dd class="mt-0.5">
              @if (editable()) {
                <button
                  type="button"
                  role="switch"
                  [attr.aria-checked]="setting.on"
                  [attr.aria-labelledby]="labelId(setting)"
                  [disabled]="busy()"
                  (click)="toggle(setting)"
                  class="inline-flex items-center gap-1.5 rounded-sm border px-2 py-0.5 text-[13px] transition-colors disabled:cursor-not-allowed disabled:opacity-50"
                  [class.border-accent]="setting.on"
                  [class.text-accent]="setting.on"
                  [class.border-line]="!setting.on"
                  [class.text-ink-dim]="!setting.on"
                >
                  <!-- The word carries the state; the dot and the colour
                       only make it quicker to find. -->
                  <span
                    class="inline-block h-1.5 w-1.5 rounded-full"
                    [class.bg-accent]="setting.on"
                    [class.bg-ink-faint]="!setting.on"
                    aria-hidden="true"
                  ></span>
                  {{ setting.on ? t('settings.on') : t('settings.off') }}
                </button>
              } @else {
                <span
                  class="text-[13px]"
                  [class.text-ink]="setting.on"
                  [class.text-ink-dim]="!setting.on"
                >
                  {{ setting.on ? t('settings.on') : t('settings.off') }}
                </span>
              }
            </dd>
          </div>
        }
      </dl>
    </section>
  `,
})
export class ObjectSettings {
  private readonly auth = inject(Auth);
  private readonly commands = inject(Commands);
  private readonly transloco = inject(TranslocoService);
  private readonly server = inject(ServerInfo);

  readonly kind = input.required<Kind>();
  readonly status = input.required<HostStatus | ServiceStatus>();

  /** Emitted once a change is accepted, so the page can reload. */
  readonly changed = output<void>();

  readonly busy = this.commands.busy;
  readonly commandsEnabled = computed(() => this.server.commandsEnabled);
  readonly editable = computed(() => this.commandsEnabled() && this.auth.can('commands:toggle'));

  readonly settings = computed<Setting[]>(() => {
    const status = this.status();
    return [
      { name: 'active_checks', label: 'activeChecks', on: status.active_checks_enabled },
      { name: 'passive_checks', label: 'passiveChecks', on: status.passive_checks_enabled },
      { name: 'notifications', label: 'notifications', on: status.notifications_enabled },
      { name: 'flap_detection', label: 'flapDetection', on: status.flap_detection_enabled },
      { name: 'event_handler', label: 'eventHandler', on: status.event_handler_enabled },
    ];
  });

  labelId(setting: Setting): string {
    return `setting-${setting.name}`;
  }

  private target() {
    const status = this.status();
    return {
      kind: this.kind(),
      host: status.hostname,
      service:
        this.kind() === 'service' ? (status as ServiceStatus).service_description : undefined,
    };
  }

  private label(): string {
    const target = this.target();
    return target.service ? `${target.host} / ${target.service}` : target.host;
  }

  async toggle(setting: Setting): Promise<void> {
    const enable = !setting.on;
    const name = this.transloco.translate('settings.' + setting.label);

    const ok = await this.commands.run({
      action: 'toggle',
      targets: [this.target()],
      body: { switch: setting.name, enable },
      pending: this.transloco.translate(enable ? 'settings.enabling' : 'settings.disabling', {
        setting: name,
        target: this.label(),
      }),
      success: this.transloco.translate(enable ? 'settings.enabled' : 'settings.disabled', {
        setting: name,
        target: this.label(),
      }),
      // Confirmed against the object's own reading of the same switch,
      // which is the only thing that proves the core applied it.
      verify: (status) => readSwitch(status, setting.name) === enable,
    });
    if (ok) {
      this.changed.emit();
    }
  }
}

function readSwitch(status: HostStatus | ServiceStatus, name: SwitchName): boolean {
  switch (name) {
    case 'active_checks':
      return status.active_checks_enabled;
    case 'passive_checks':
      return status.passive_checks_enabled;
    case 'notifications':
      return status.notifications_enabled;
    case 'flap_detection':
      return status.flap_detection_enabled;
    case 'event_handler':
      return status.event_handler_enabled;
  }
}
