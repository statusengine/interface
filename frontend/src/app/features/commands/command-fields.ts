import { ChangeDetectionStrategy, Component, input, model } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TranslocoDirective } from '@jsverse/transloco';

/** Minutes offered for a quick downtime. */
export const DOWNTIME_PRESETS = [30, 60, 120, 240, 480, 1440];

/**
 * The acknowledgement form.
 *
 * Its own component because the same fields are filled in for one
 * object and for fifty, and a second copy of them is a second place for
 * the wording and the defaults to drift.
 */
@Component({
  selector: 'sei-acknowledge-fields',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormsModule, TranslocoDirective],
  template: `
    <ng-container *transloco="let t">
      <p class="mb-3 text-[13px] text-ink-dim">{{ t('commands.acknowledgeExplainer') }}</p>

      <label class="mb-1.5 block text-[13px] font-medium" [attr.for]="id() + '-comment'">
        {{ t('commands.comment') }}
      </label>
      <textarea
        [id]="id() + '-comment'"
        rows="3"
        required
        [ngModel]="comment()"
        (ngModelChange)="comment.set($event)"
        [attr.placeholder]="t('commands.commentPlaceholder')"
        class="sei-input"
      ></textarea>
      <p class="mt-1 text-[11px] text-ink-dim">{{ t('commands.noSemicolons') }}</p>

      <div class="mt-3 space-y-2">
        <label class="flex items-start gap-2 text-[13px]">
          <input
            type="checkbox"
            [ngModel]="sticky()"
            (ngModelChange)="sticky.set($event)"
            class="mt-0.5"
          />
          <span>
            {{ t('commands.sticky') }}
            <span class="block text-[12px] text-ink-dim">{{ t('commands.stickyHelp') }}</span>
          </span>
        </label>
        <label class="flex items-start gap-2 text-[13px]">
          <input
            type="checkbox"
            [ngModel]="notify()"
            (ngModelChange)="notify.set($event)"
            class="mt-0.5"
          />
          <span>
            {{ t('commands.notifyContacts') }}
            <span class="block text-[12px] text-ink-dim">{{
              t('commands.notifyContactsHelp')
            }}</span>
          </span>
        </label>
        <label class="flex items-start gap-2 text-[13px]">
          <input
            type="checkbox"
            [ngModel]="persistent()"
            (ngModelChange)="persistent.set($event)"
            class="mt-0.5"
          />
          <span>
            {{ t('commands.persistent') }}
            <span class="block text-[12px] text-ink-dim">{{ t('commands.persistentHelp') }}</span>
          </span>
        </label>
      </div>
    </ng-container>
  `,
})
export class AcknowledgeFields {
  /** Prefixes the field ids, so two of these can never collide. */
  readonly id = input('ack');

  readonly comment = model('');
  readonly sticky = model(true);
  readonly notify = model(false);
  readonly persistent = model(false);
}

/** The downtime form. */
@Component({
  selector: 'sei-downtime-fields',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormsModule, TranslocoDirective],
  template: `
    <ng-container *transloco="let t">
      <p class="mb-3 text-[13px] text-ink-dim">{{ t('commands.downtimeExplainer') }}</p>

      <span class="mb-1.5 block text-[13px] font-medium">{{ t('commands.duration') }}</span>
      <div class="flex flex-wrap gap-1">
        @for (preset of presets; track preset) {
          <button
            type="button"
            (click)="minutes.set(preset)"
            class="rounded-sm border px-2.5 py-1 text-[12px] transition-colors"
            [class.border-accent]="minutes() === preset"
            [class.text-accent]="minutes() === preset"
            [class.bg-accent-wash]="minutes() === preset"
            [class.border-line]="minutes() !== preset"
            [class.text-ink-dim]="minutes() !== preset"
            [attr.aria-pressed]="minutes() === preset"
          >
            {{ preset < 60 ? preset + 'm' : preset / 60 + 'h' }}
          </button>
        }
      </div>

      <label class="mb-1.5 mt-3 block text-[13px] font-medium" [attr.for]="id() + '-comment'">
        {{ t('commands.comment') }}
      </label>
      <textarea
        [id]="id() + '-comment'"
        rows="2"
        required
        [ngModel]="comment()"
        (ngModelChange)="comment.set($event)"
        [attr.placeholder]="t('commands.downtimeCommentPlaceholder')"
        class="sei-input"
      ></textarea>

      @if (offerAllServices()) {
        <label class="mt-3 flex items-start gap-2 text-[13px]">
          <input
            type="checkbox"
            [ngModel]="allServices()"
            (ngModelChange)="allServices.set($event)"
            class="mt-0.5"
          />
          <span>
            {{ t('commands.allServices') }}
            <span class="block text-[12px] text-ink-dim">{{ t('commands.allServicesHelp') }}</span>
          </span>
        </label>
      }
    </ng-container>
  `,
})
export class DowntimeFields {
  readonly id = input('downtime');

  /** Only meaningful when at least one target is a host. */
  readonly offerAllServices = input(false);

  readonly presets = DOWNTIME_PRESETS;
  readonly minutes = model(60);
  readonly comment = model('');
  readonly allServices = model(false);
}
