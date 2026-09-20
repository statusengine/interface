import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';

/** The switch between the three history views. */
@Component({
  selector: 'sei-history-nav',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [RouterLink, RouterLinkActive, TranslocoDirective],
  template: `
    <nav
      class="mt-3 flex flex-wrap gap-1"
      *transloco="let t"
      [attr.aria-label]="t('history.views')"
    >
      @for (view of views; track view.path) {
        <a
          [routerLink]="view.path"
          routerLinkActive="border-accent text-accent bg-accent-wash"
          queryParamsHandling="preserve"
          class="rounded-sm border border-line px-2.5 py-1 text-[12px] text-ink-dim transition-colors hover:text-ink"
        >
          {{ t('nav.' + view.label) }}
        </a>
      }
    </nav>
  `,
})
export class HistoryNav {
  readonly views = [
    { path: '/history/checks', label: 'historyChecks' },
    { path: '/history/statechanges', label: 'historyStateChanges' },
    { path: '/history/notifications', label: 'historyNotifications' },
  ];
}
