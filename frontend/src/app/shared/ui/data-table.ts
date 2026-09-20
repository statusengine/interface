import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';
import { Icon } from './icon';

/**
 * The frame every list page renders into: a dense table with a sticky
 * header, and the three states a remote list can be in.
 *
 * On a narrow screen the table scrolls sideways rather than reshaping
 * into cards. Each page marks its secondary columns so they drop away
 * below a breakpoint, which leaves the two that matter - what is broken
 * and what it said - readable on a phone. One DOM, one set of keyboard
 * and screen-reader semantics, and a table that is still a table when
 * someone turns their phone sideways.
 */
@Component({
  selector: 'sei-data-table',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective, Icon],
  template: `
    <ng-container *transloco="let t">
      @if (error()) {
        <div class="px-4 py-10 sm:px-6">
          <div class="max-w-prose rounded-md border-l-2 border-critical bg-surface p-4">
            <h2 class="text-[15px]">{{ t('list.errorTitle') }}</h2>
            <p class="mt-1.5 text-[13px] text-ink-dim">{{ error() }}</p>
            <button
              type="button"
              (click)="retry.emit()"
              class="mt-3 inline-flex items-center gap-1.5 rounded-sm border border-line px-2.5 py-1.5 text-[13px] transition-colors hover:border-line-strong"
            >
              <sei-icon name="refresh" [size]="14" />
              {{ t('list.retry') }}
            </button>
          </div>
        </div>
      } @else if (empty()) {
        <div class="px-4 py-14 sm:px-6">
          <div class="max-w-prose">
            <h2 class="text-[15px]">{{ emptyTitle() }}</h2>
            <p class="mt-1.5 text-[13px] text-ink-dim">{{ emptyBody() }}</p>
            @if (showClear()) {
              <button
                type="button"
                (click)="clear.emit()"
                class="mt-3 rounded-sm border border-line px-2.5 py-1.5 text-[13px] transition-colors hover:border-line-strong"
              >
                {{ t('list.clearFilters') }}
              </button>
            }
          </div>
        </div>
      } @else {
        <div class="relative overflow-x-auto">
          <!-- A thin bar rather than a spinner over the table: the rows
               on screen are still the right answer to the previous
               question, and blanking them makes paging feel broken. -->
          @if (loading()) {
            <div
              class="absolute inset-x-0 top-0 z-20 h-0.5 overflow-hidden bg-accent-wash"
              role="status"
              [attr.aria-label]="t('list.loading')"
            >
              <div
                class="h-full w-1/3 animate-[sei-sweep_1.1s_ease-in-out_infinite] bg-accent"
              ></div>
            </div>
          }
          <table class="w-full border-collapse text-[13px]">
            <thead class="sticky top-0 z-10 bg-surface">
              <ng-content select="[head]" />
            </thead>
            <tbody>
              <ng-content select="[body]" />
            </tbody>
          </table>
        </div>
      }
    </ng-container>
  `,
  styles: `
    @keyframes sei-sweep {
      0% {
        transform: translateX(-100%);
      }
      100% {
        transform: translateX(400%);
      }
    }
  `,
})
export class DataTable {
  readonly loading = input(false);
  readonly empty = input(false);
  readonly error = input<string | null>(null);
  readonly emptyTitle = input('Nothing here');
  readonly emptyBody = input('');
  readonly showClear = input(false);

  readonly retry = output<void>();
  readonly clear = output<void>();
}
