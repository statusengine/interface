import { ChangeDetectionStrategy, Component, computed, input } from '@angular/core';

/**
 * Inline SVG icons. No icon font and no sprite fetched from anywhere:
 * this interface has to render completely on a host with no route to the
 * internet, and a missing glyph in a navigation rail is a dead end.
 *
 * The set is deliberately small. Every icon here appears next to a text
 * label, so it is a shape to aim at rather than the only thing carrying
 * the meaning.
 */
export type IconName =
  | 'dashboard'
  | 'problems'
  | 'host'
  | 'service'
  | 'downtime'
  | 'acknowledge'
  | 'log'
  | 'history'
  | 'chart'
  | 'menu'
  | 'close'
  | 'chevron-down'
  | 'chevron-right'
  | 'search'
  | 'sun'
  | 'moon'
  | 'monitor'
  | 'logout'
  | 'user'
  | 'globe'
  | 'collapse'
  | 'expand';

const PATHS: Record<IconName, string> = {
  dashboard: 'M3 3h7v7H3zM14 3h7v4h-7zM14 10h7v11h-7zM3 13h7v8H3z',
  problems: 'M12 3 2.5 20h19zM12 10v4M12 17.2v.05',
  host: 'M3 4.5h18v9H3zM7 18h10M9 13.5v4.5M15 13.5v4.5M6.5 8h4',
  service: 'M4 6h16M4 12h16M4 18h9M18.5 16.5l1.6 1.6 2.4-2.6',
  downtime: 'M12 3a9 9 0 1 0 9 9M12 7v5l3 2M15.5 3.5h6',
  acknowledge: 'M4 12.5l5 5L20 6.5',
  log: 'M5 3h9l5 5v13H5zM14 3v5h5M8 13h8M8 17h5',
  history: 'M3.5 12a8.5 8.5 0 1 0 2.6-6.1M3.5 4.5V9h4.5M12 7.5V12l3 2',
  chart: 'M3 20h18M6 20V11M11 20V5M16 20v-6',
  menu: 'M4 6h16M4 12h16M4 18h16',
  close: 'M6 6l12 12M18 6L6 18',
  'chevron-down': 'M5 8.5l7 7 7-7',
  'chevron-right': 'M8.5 5l7 7-7 7',
  search: 'M10.5 3a7.5 7.5 0 1 0 0 15 7.5 7.5 0 0 0 0-15zM16 16l5 5',
  sun: 'M12 6.5a5.5 5.5 0 1 0 0 11 5.5 5.5 0 0 0 0-11zM12 1.5v2.5M12 20v2.5M3.9 3.9l1.8 1.8M18.3 18.3l1.8 1.8M1.5 12H4M20 12h2.5M3.9 20.1l1.8-1.8M18.3 5.7l1.8-1.8',
  moon: 'M20 14.5A8.5 8.5 0 0 1 9.5 4a8.5 8.5 0 1 0 10.5 10.5z',
  monitor: 'M3 4.5h18v12H3zM8.5 20.5h7M12 16.5v4',
  logout: 'M14 4.5h5v15h-5M11 15.5l3.5-3.5L11 8.5M3 12h11.5',
  user: 'M12 3.5a4 4 0 1 0 0 8 4 4 0 0 0 0-8zM4.5 20.5a7.5 7.5 0 0 1 15 0',
  globe:
    'M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18zM3 12h18M12 3c2.5 2.6 3.8 5.6 3.8 9S14.5 18.4 12 21c-2.5-2.6-3.8-5.6-3.8-9S9.5 5.6 12 3z',
  collapse: 'M15.5 5l-7 7 7 7M20 5v14',
  expand: 'M8.5 5l7 7-7 7M4 5v14',
};

@Component({
  selector: 'sei-icon',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <svg
      [attr.width]="size()"
      [attr.height]="size()"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      [attr.stroke-width]="strokeWidth()"
      stroke-linecap="round"
      stroke-linejoin="round"
      aria-hidden="true"
      focusable="false"
    >
      <path [attr.d]="path()" />
    </svg>
  `,
  styles: `
    :host {
      display: inline-flex;
      flex: none;
      align-items: center;
      justify-content: center;
    }
  `,
})
export class Icon {
  readonly name = input.required<IconName>();
  readonly size = input(16);

  /** Thinner strokes at larger sizes, so an icon keeps the same weight on
   *  screen instead of turning into a slab. */
  readonly strokeWidth = computed(() => (this.size() >= 24 ? 1.5 : 1.7));

  readonly path = computed(() => PATHS[this.name()]);
}
