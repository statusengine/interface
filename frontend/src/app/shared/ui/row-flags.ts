import { ChangeDetectionStrategy, Component, computed, input } from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';
import { Icon, type IconName } from './icon';

interface Flag {
  icon: IconName;
  key: string;
  tone: string;
}

/**
 * The handful of per-object states that change how a row should be read:
 * acknowledged, in a downtime, flapping, notifications off, active checks
 * off.
 *
 * They are icons with a title rather than words because a list of five
 * labels would be wider than the output column, and every one of them is
 * the exception rather than the rule - on a healthy row this renders
 * nothing at all.
 */
@Component({
  selector: 'sei-row-flags',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective, Icon],
  template: `
    @if (flags().length) {
      <span class="inline-flex items-center gap-1" *transloco="let t">
        @for (flag of flags(); track flag.key) {
          <span [class]="flag.tone" [title]="t('flags.' + flag.key)">
            <sei-icon [name]="flag.icon" [size]="13" />
            <span class="sr-only">{{ t('flags.' + flag.key) }}</span>
          </span>
        }
      </span>
    }
  `,
})
export class RowFlags {
  readonly acknowledged = input(false);
  readonly inDowntime = input(false);
  readonly flapping = input(false);
  readonly notificationsEnabled = input(true);
  readonly activeChecksEnabled = input(true);

  readonly flags = computed<Flag[]>(() => {
    const out: Flag[] = [];
    if (this.acknowledged()) {
      out.push({ icon: 'check-circle', key: 'acknowledged', tone: 'text-ok' });
    }
    if (this.inDowntime()) {
      out.push({ icon: 'pause', key: 'inDowntime', tone: 'text-accent' });
    }
    if (this.flapping()) {
      out.push({ icon: 'wave', key: 'flapping', tone: 'text-warning' });
    }
    if (!this.notificationsEnabled()) {
      out.push({ icon: 'bell-off', key: 'notificationsDisabled', tone: 'text-ink-dim' });
    }
    if (!this.activeChecksEnabled()) {
      out.push({ icon: 'clock', key: 'activeChecksDisabled', tone: 'text-ink-dim' });
    }
    return out;
  });
}
