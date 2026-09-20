import { ChangeDetectionStrategy, Component, computed, input } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';
import type { StateCounts } from '../../core/api/types';
import { stateClass, stateTextClass, type StateKey } from '../state/state';

interface Segment {
  key: StateKey;
  value: number;
  count: number;
  percent: number;
}

/**
 * A population broken down by state, as one proportional bar plus a
 * legend.
 *
 * Four identical cards with big numbers would take four times the space
 * to say less: the thing an operator wants from a dashboard at a glance
 * is the shape - is this a handful of red on a sea of green, or is half
 * the estate down. A bar shows that in one saccade, and the counts are
 * still there to read.
 *
 * Every segment and every legend entry is a link into the matching
 * filtered list, so the dashboard is a way in rather than a dead end.
 */
@Component({
  selector: 'sei-state-strip',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [RouterLink, TranslocoDirective],
  template: `
    <section class="rounded-md border border-line bg-surface p-4" *transloco="let t">
      <div class="mb-3 flex items-baseline justify-between gap-3">
        <h2 class="text-[13px] font-medium">
          <a [routerLink]="link()" class="hover:text-accent hover:underline">{{ title() }}</a>
        </h2>
        <p class="text-[12px] text-ink-dim">
          {{ t('dashboard.totalCount', { count: counts().total }) }}
        </p>
      </div>

      @if (counts().total === 0) {
        <p class="text-[13px] text-ink-dim">{{ t('dashboard.nothingMonitored') }}</p>
      } @else {
        <div
          class="flex h-2 w-full overflow-hidden rounded-full bg-sunken"
          role="img"
          [attr.aria-label]="summaryLabel()"
        >
          @for (segment of segments(); track segment.key) {
            @if (segment.count > 0) {
              <a
                [routerLink]="link()"
                [queryParams]="{ state: segment.value }"
                class="block h-full transition-opacity hover:opacity-80"
                [style.width.%]="segment.percent"
                [class]="barClass(segment.key)"
                [title]="t('states.' + segment.key) + ': ' + segment.count"
              ></a>
            }
          }
        </div>

        <ul class="mt-3 flex flex-wrap gap-x-4 gap-y-1.5">
          @for (segment of segments(); track segment.key) {
            <li>
              <a
                [routerLink]="link()"
                [queryParams]="{ state: segment.value }"
                class="group flex items-baseline gap-1.5 text-[12px]"
              >
                <span
                  class="tabular-nums text-[15px] font-semibold"
                  [class]="segment.count > 0 ? textClass(segment.key) : 'text-ink-faint'"
                  >{{ segment.count }}</span
                >
                <span class="text-ink-dim group-hover:text-ink group-hover:underline">{{
                  t('states.' + segment.key)
                }}</span>
              </a>
            </li>
          }
          @if (counts().pending > 0) {
            <li class="flex items-baseline gap-1.5 text-[12px]">
              <span class="tabular-nums text-[15px] font-semibold text-pending">{{
                counts().pending
              }}</span>
              <span class="text-ink-dim">{{ t('states.pending') }}</span>
            </li>
          }
        </ul>

        @if (counts().problems > 0) {
          <p class="mt-3 border-t border-line pt-2.5 text-[12px] text-ink-dim">
            <a
              [routerLink]="link()"
              [queryParams]="{ problems: 'true', handled: 'false' }"
              class="hover:text-ink hover:underline"
            >
              {{
                t('dashboard.unhandledOf', {
                  unhandled: counts().unhandled,
                  problems: counts().problems,
                })
              }}
            </a>
          </p>
        }
      }
    </section>
  `,
})
export class StateStrip {
  readonly title = input.required<string>();
  readonly counts = input.required<StateCounts>();
  /** State keys in the order they should appear, worst last so the bar
   *  reads from good to bad left to right. */
  readonly states = input.required<{ value: number; key: StateKey }[]>();
  readonly link = input.required<string>();

  readonly segments = computed<Segment[]>(() => {
    const counts = this.counts();
    // The bar is proportional to what has actually been checked; pending
    // objects have no state to colour and are listed separately.
    const checked = Math.max(1, counts.total - counts.pending);
    return this.states().map((state) => {
      const count = counts.by_state[state.key] ?? 0;
      return {
        key: state.key,
        value: state.value,
        count,
        percent: (count / checked) * 100,
      };
    });
  });

  readonly summaryLabel = computed(() =>
    this.segments()
      .filter((s) => s.count > 0)
      .map((s) => `${s.count} ${s.key}`)
      .join(', '),
  );

  barClass(key: StateKey): string {
    // The rail classes set --rail-color; here the same colour fills a bar.
    return {
      'state-ok': 'bg-ok',
      'state-warning': 'bg-warning',
      'state-critical': 'bg-critical',
      'state-unknown': 'bg-unknown',
      'state-unreachable': 'bg-unreachable',
      'state-pending': 'bg-pending',
    }[stateClass(key)]!;
  }

  textClass(key: StateKey): string {
    return stateTextClass(key);
  }
}
