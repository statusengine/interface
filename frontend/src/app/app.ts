import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';

@Component({
  selector: 'app-root',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [RouterOutlet, TranslocoDirective],
  template: `
    <ng-container *transloco="let t">
      <!-- First stop for a keyboard user: past the rail, into the table. -->
      <a
        href="#main"
        class="sr-only focus:not-sr-only focus:fixed focus:left-3 focus:top-3 focus:z-50 focus:rounded-sm focus:bg-accent focus:px-3 focus:py-2 focus:text-[13px] focus:text-accent-ink"
      >
        {{ t('a11y.skipToContent') }}
      </a>
    </ng-container>
    <router-outlet />
  `,
})
export class App {}
