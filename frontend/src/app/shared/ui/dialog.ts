import {
  ChangeDetectionStrategy,
  Component,
  ElementRef,
  effect,
  input,
  output,
  viewChild,
} from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';
import { Icon } from './icon';

/**
 * A modal dialog.
 *
 * Built on the native `<dialog>` element with `showModal()`, which
 * brings the focus trap, the Escape handling, the inert background and
 * the top-layer stacking for free. Every hand-rolled modal reimplements
 * those four things and most of them get the focus trap wrong.
 *
 * `m-auto` is not decoration: a modal dialog centres itself through its
 * automatic margins, and Tailwind's preflight sets `margin: 0` on
 * everything, which pins it to the top left instead.
 */
@Component({
  selector: 'sei-dialog',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective, Icon],
  template: `
    <dialog
      #dialog
      (close)="closed.emit()"
      (cancel)="closed.emit()"
      class="m-auto w-[min(32rem,calc(100vw-2rem))] rounded-lg border border-line bg-surface p-0 text-ink shadow-[var(--shadow-overlay)] backdrop:bg-black/60"
      *transloco="let t"
    >
      <form method="dialog" class="contents">
        <header class="flex items-start justify-between gap-3 border-b border-line px-4 py-3">
          <div class="min-w-0">
            <h2 class="text-[15px]">{{ title() }}</h2>
            @if (subtitle()) {
              <p class="mono mt-0.5 truncate text-[12px] text-ink-dim">{{ subtitle() }}</p>
            }
          </div>
          <button
            type="submit"
            value="dismiss"
            class="-mr-1 -mt-1 rounded-sm p-1.5 text-ink-dim transition-colors hover:bg-sunken hover:text-ink"
            [attr.aria-label]="t('dialog.close')"
          >
            <sei-icon name="close" [size]="16" />
          </button>
        </header>
      </form>

      <div class="px-4 py-4">
        <ng-content />
      </div>

      <footer class="flex flex-wrap justify-end gap-2 border-t border-line px-4 py-3">
        <ng-content select="[footer]" />
      </footer>
    </dialog>
  `,
})
export class Dialog {
  readonly open = input(false);
  readonly title = input.required<string>();
  readonly subtitle = input('');

  readonly closed = output<void>();

  // Not `required`: the effect below is registered in the constructor
  // and runs before the view exists, and reading a required query that
  // early is NG0951. The signal is reactive, so the effect re-runs once
  // the element is there.
  private readonly dialog = viewChild<ElementRef<HTMLDialogElement>>('dialog');

  constructor() {
    effect(() => {
      const element = this.dialog()?.nativeElement;
      if (!element) {
        return;
      }
      if (this.open()) {
        if (!element.open) {
          element.showModal();
        }
      } else if (element.open) {
        element.close();
      }
    });
  }
}
