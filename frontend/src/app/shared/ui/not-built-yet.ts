import { ChangeDetectionStrategy, Component, input } from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';

/**
 * Stands in for a page that is routed but not implemented yet.
 *
 * It says which phase the page belongs to rather than showing a spinner
 * or a fake table. An empty screen that explains itself is navigable; one
 * that pretends to be loading is not.
 */
@Component({
  selector: 'sei-not-built-yet',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective],
  template: `
    <div class="px-4 py-10 sm:px-6" *transloco="let t">
      <div class="max-w-prose rounded-md border border-dashed border-line bg-surface p-5">
        <h2 class="text-[15px]">{{ t('placeholder.title') }}</h2>
        <p class="mt-1.5 text-[13px] text-ink-dim">
          {{ t('placeholder.body', { phase: phase() }) }}
        </p>
      </div>
    </div>
  `,
})
export class NotBuiltYet {
  readonly phase = input.required<number>();
}
