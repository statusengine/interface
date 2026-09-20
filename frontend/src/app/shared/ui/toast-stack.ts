import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';
import { Toasts, type ToastTone } from '../../core/toast/toast.service';
import { Icon, type IconName } from './icon';

const TONE: Record<ToastTone, { border: string; text: string; icon: IconName }> = {
  info: { border: 'border-l-accent', text: 'text-accent', icon: 'clock' },
  success: { border: 'border-l-ok', text: 'text-ok', icon: 'check-circle' },
  warning: { border: 'border-l-warning', text: 'text-warning', icon: 'alert' },
  error: { border: 'border-l-critical', text: 'text-critical', icon: 'alert' },
  pending: { border: 'border-l-line-strong', text: 'text-ink-dim', icon: 'refresh' },
};

/**
 * The toast stack.
 *
 * Bottom right and above the content, announced politely so a screen
 * reader hears the outcome of a command without losing the operator's
 * place.
 */
@Component({
  selector: 'sei-toast-stack',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective, Icon],
  template: `
    <div
      class="pointer-events-none fixed inset-x-3 bottom-3 z-50 flex flex-col items-end gap-2 sm:left-auto sm:right-4 sm:w-96"
      role="status"
      aria-live="polite"
      *transloco="let t"
    >
      @for (toast of toasts.items(); track toast.id) {
        <div
          class="pointer-events-auto w-full rounded-md border border-line border-l-2 bg-raised px-3 py-2.5 shadow-[var(--shadow-overlay)]"
          [class]="tone(toast.tone).border"
        >
          <div class="flex items-start gap-2.5">
            <span [class]="tone(toast.tone).text" class="mt-0.5">
              <sei-icon
                [name]="tone(toast.tone).icon"
                [size]="15"
                [class.animate-spin]="toast.tone === 'pending'"
              />
            </span>
            <div class="min-w-0 flex-1">
              <p class="text-[13px] font-medium">{{ toast.title }}</p>
              @if (toast.body) {
                <p class="mt-0.5 text-[12px] leading-snug text-ink-dim">{{ toast.body }}</p>
              }
            </div>
            <button
              type="button"
              (click)="toasts.dismiss(toast.id)"
              class="-mr-1 -mt-1 rounded-sm p-1 text-ink-dim transition-colors hover:bg-sunken hover:text-ink"
              [attr.aria-label]="t('dialog.dismiss')"
            >
              <sei-icon name="close" [size]="13" />
            </button>
          </div>
        </div>
      }
    </div>
  `,
  styles: `
    @keyframes sei-spin {
      to {
        transform: rotate(360deg);
      }
    }
    .animate-spin {
      display: inline-flex;
      animation: sei-spin 1.4s linear infinite;
    }
  `,
})
export class ToastStack {
  readonly toasts = inject(Toasts);

  tone(name: ToastTone) {
    return TONE[name];
  }
}
