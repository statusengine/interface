import { ChangeDetectionStrategy, Component, inject, input, output, signal } from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';
import { Api } from '../../core/api/api.service';
import type { ListResponse } from '../../core/api/types';
import { Icon } from '../../shared/ui/icon';
import { RangePicker } from '../../shared/ui/range-picker';

/**
 * The object and time controls every history page shares.
 *
 * Naming a host is not a nicety here. The history tables are clustered
 * on an object-first key, so a scoped query is one contiguous range and
 * the server allows ninety days; unscoped, it allows six hours. The bar
 * says which of the two is in force rather than letting an operator
 * discover it from a rejected request.
 */
@Component({
  selector: 'sei-history-scope',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective, RangePicker, Icon],
  template: `
    <div class="border-b border-line bg-surface" *transloco="let t">
      <div class="flex flex-wrap items-center gap-2 px-4 py-2.5 sm:px-6">
        <label class="flex items-center gap-1.5 text-[12px] text-ink-dim">
          {{ t('columns.host') }}
          <input
            type="text"
            list="sei-host-names"
            [value]="host() ?? ''"
            (change)="hostChange.emit($any($event.target).value.trim() || undefined)"
            [attr.placeholder]="t('history.anyHost')"
            class="mono w-44 rounded-sm border border-line bg-ground px-2 py-1 text-[12px] text-ink outline-none transition-colors focus:border-accent"
          />
        </label>
        <datalist id="sei-host-names">
          @for (name of hostNames(); track name) {
            <option [value]="name"></option>
          }
        </datalist>

        <label class="flex items-center gap-1.5 text-[12px] text-ink-dim">
          {{ t('columns.service') }}
          <input
            type="text"
            [value]="service() ?? ''"
            [disabled]="!host()"
            (change)="serviceChange.emit($any($event.target).value.trim() || undefined)"
            [attr.placeholder]="host() ? t('history.anyService') : t('history.needsHost')"
            class="mono w-48 rounded-sm border border-line bg-ground px-2 py-1 text-[12px] text-ink outline-none transition-colors focus:border-accent disabled:opacity-50"
          />
        </label>

        <sei-range-picker
          [from]="from()"
          [to]="to()"
          [defaultSeconds]="host() ? 24 * 3600 : 3600"
          [maxSeconds]="host() ? 90 * 24 * 3600 : 6 * 3600"
          (change)="rangeChange.emit($event)"
        />

        <ng-content />
      </div>

      <p
        class="flex items-start gap-1.5 border-t border-line px-4 py-2 text-[12px] sm:px-6"
        [class.text-ink-dim]="!!host()"
        [class.text-warning]="!host()"
      >
        <sei-icon [name]="host() ? 'check-circle' : 'alert'" [size]="13" class="mt-0.5" />
        <span>{{ host() ? t('history.scopedHint') : t('history.unscopedHint') }}</span>
      </p>
    </div>
  `,
})
export class HistoryScope {
  private readonly api = inject(Api);

  readonly host = input<string | undefined>(undefined);
  readonly service = input<string | undefined>(undefined);
  readonly from = input<string | undefined>(undefined);
  readonly to = input<string | undefined>(undefined);

  readonly hostChange = output<string | undefined>();
  readonly serviceChange = output<string | undefined>();
  readonly rangeChange = output<{ from: string; to: string }>();

  readonly hostNames = signal<string[]>([]);

  constructor() {
    void this.loadNames();
  }

  private async loadNames(): Promise<void> {
    try {
      const res = (await this.api.list<string>('/hosts/names')) as ListResponse<string>;
      this.hostNames.set(res.data);
    } catch {
      // A missing datalist costs an operator autocomplete, not the page.
      // Typing the host name still works.
    }
  }
}
