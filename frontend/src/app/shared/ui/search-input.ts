import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';
import { Icon } from './icon';

/** A search box. Debouncing is the store's job, not this component's. */
@Component({
  selector: 'sei-search',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective, Icon],
  template: `
    <label class="relative flex min-w-0 flex-1 items-center sm:max-w-72" *transloco="let t">
      <span class="pointer-events-none absolute left-2.5 text-ink-faint">
        <sei-icon name="search" [size]="14" />
      </span>
      <input
        type="search"
        [value]="value()"
        (input)="search.emit($any($event.target).value)"
        [attr.placeholder]="placeholder() || t('list.searchPlaceholder')"
        [attr.aria-label]="placeholder() || t('list.searchPlaceholder')"
        class="mono w-full rounded-sm border border-line bg-ground py-1.5 pl-8 pr-2.5 text-[13px] text-ink outline-none transition-colors focus:border-accent"
      />
    </label>
  `,
})
export class SearchInput {
  readonly value = input('');
  readonly placeholder = input('');
  readonly search = output<string>();
}
