import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';

/**
 * Reached when a route's permission guard turns someone away. It names
 * the permission that was missing, because "access denied" leaves an
 * operator with nothing to ask their administrator for.
 */
@Component({
  selector: 'sei-forbidden',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [RouterLink, TranslocoDirective],
  template: `
    <div class="px-4 py-12 sm:px-6" *transloco="let t">
      <div class="max-w-prose">
        <h1>{{ t('errors.forbiddenTitle') }}</h1>
        <p class="mt-2 text-[13px] text-ink-dim">{{ t('errors.forbiddenBody') }}</p>
        @if (needed) {
          <p
            class="mono mt-3 inline-block rounded-sm border border-line bg-surface px-2 py-1 text-[12px]"
          >
            {{ needed }}
          </p>
        }
        <p class="mt-5">
          <a routerLink="/dashboard" class="text-[13px] text-accent hover:underline">
            {{ t('errors.backToDashboard') }}
          </a>
        </p>
      </div>
    </div>
  `,
})
export class Forbidden {
  readonly needed = inject(ActivatedRoute).snapshot.queryParamMap.get('need');
}
