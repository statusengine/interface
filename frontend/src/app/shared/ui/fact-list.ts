import { ChangeDetectionStrategy, Component, input } from '@angular/core';

export interface Fact {
  /** Translation key under `detail.`. */
  label: string;
  value: string;
  /** Set for machine-emitted values: command lines, timeperiod names. */
  mono?: boolean;
  /** A tone class, for values that carry a state. */
  tone?: string;
}

/**
 * A definition list of facts about one object.
 *
 * Two columns on a phone, more as there is room. A real `dl` rather than
 * a table, because these are label-value pairs and not a grid an operator
 * would sort or scan across.
 */
@Component({
  selector: 'sei-fact-list',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <dl class="grid grid-cols-2 gap-x-6 gap-y-3 sm:grid-cols-3 xl:grid-cols-4">
      @for (fact of facts(); track fact.label) {
        <div class="min-w-0">
          <dt class="text-[11px] text-ink-dim">{{ fact.label }}</dt>
          <dd
            class="truncate text-[13px]"
            [class.mono]="fact.mono"
            [class]="fact.tone ?? 'text-ink'"
            [title]="fact.value"
          >
            {{ fact.value }}
          </dd>
        </div>
      }
    </dl>
  `,
})
export class FactList {
  readonly facts = input.required<Fact[]>();
}
