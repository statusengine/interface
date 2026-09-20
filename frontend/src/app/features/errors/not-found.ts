import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';

@Component({
  selector: 'sei-not-found',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [RouterLink, TranslocoDirective],
  template: `
    <div class="px-4 py-12 sm:px-6" *transloco="let t">
      <div class="max-w-prose">
        <h1>{{ t('errors.notFoundTitle') }}</h1>
        <p class="mt-2 text-[13px] text-ink-dim">{{ t('errors.notFoundBody') }}</p>
        <p class="mt-5">
          <a routerLink="/dashboard" class="text-[13px] text-accent hover:underline">
            {{ t('errors.backToDashboard') }}
          </a>
        </p>
      </div>
    </div>
  `,
})
export class NotFound {}
