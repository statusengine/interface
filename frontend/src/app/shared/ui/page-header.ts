import { ChangeDetectionStrategy, Component, input } from '@angular/core';

/**
 * The band at the top of every page. Title on the left, page-level actions
 * on the right via content projection.
 */
@Component({
  selector: 'sei-page-header',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="border-b border-line bg-surface px-4 py-3.5 sm:px-6">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div class="min-w-0">
          <h1 class="truncate">{{ title() }}</h1>
          @if (description()) {
            <p class="mt-0.5 max-w-prose text-[13px] text-ink-dim">{{ description() }}</p>
          }
        </div>
        <div class="flex shrink-0 items-center gap-2">
          <ng-content select="[actions]" />
        </div>
      </div>
      <ng-content />
    </div>
  `,
})
export class PageHeader {
  readonly title = input.required<string>();
  readonly description = input<string>('');
}
