import { ChangeDetectionStrategy, Component, computed, input } from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';

/**
 * What the check plugin said, verbatim.
 *
 * Monospace and preserved whitespace, because this is machine output and
 * plugins align things with spaces. Long output follows the first line
 * the way Naemon separates them, joined here in TypeScript rather than in
 * the template - a newline inside an interpolation is the kind of thing a
 * formatter reflows into something that no longer means what it did.
 *
 * Performance data is kept apart from the prose: it is a different kind
 * of statement about the same check, and running the two together is how
 * the classic interface produces lines nobody can read.
 */
@Component({
  selector: 'sei-plugin-output',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective],
  template: `
    <div *transloco="let t">
      @if (text()) {
        <pre
          class="mono overflow-x-auto whitespace-pre-wrap break-words rounded-sm bg-sunken p-3 text-[12.5px] leading-relaxed text-ink"
          >{{ text() }}</pre>
      } @else {
        <p class="text-[13px] text-ink-dim">{{ t('detail.noOutput') }}</p>
      }

      @if (perfdata()) {
        <details class="mt-2">
          <summary class="cursor-pointer text-[12px] text-ink-dim hover:text-ink">
            {{ t('detail.performanceData') }}
          </summary>
          <pre
            class="mono mt-1.5 overflow-x-auto whitespace-pre-wrap break-all rounded-sm bg-sunken p-3 text-[12px] text-ink-dim"
            >{{ perfdata() }}</pre>
        </details>
      }
    </div>
  `,
})
export class PluginOutput {
  readonly output = input('');
  readonly longOutput = input('');
  readonly perfdata = input('');

  readonly text = computed(() =>
    [this.output(), this.longOutput()].filter((part) => part.length > 0).join('\n'),
  );

  readonly hasAny = computed(() => !!(this.text() || this.perfdata()));
}
