import { ChangeDetectionStrategy, Component, HostListener, signal } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';
import { Sidebar } from '../sidebar/sidebar';
import { Topbar } from '../topbar/topbar';

const COLLAPSE_KEY = 'sei.rail.collapsed';

/**
 * The application frame: a persistent rail on a wide screen, an overlay
 * drawer on a narrow one, and the routed page between them.
 */
@Component({
  selector: 'sei-shell',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [RouterOutlet, Sidebar, Topbar, TranslocoDirective],
  templateUrl: './shell.html',
})
export class Shell {
  readonly drawerOpen = signal(false);
  readonly railCollapsed = signal(readCollapsed());

  openDrawer(): void {
    this.drawerOpen.set(true);
  }

  closeDrawer(): void {
    this.drawerOpen.set(false);
  }

  toggleRail(): void {
    this.railCollapsed.update((c) => !c);
    try {
      localStorage.setItem(COLLAPSE_KEY, String(this.railCollapsed()));
    } catch {
      // Applied for this visit, just not remembered.
    }
  }

  @HostListener('document:keydown.escape')
  onEscape(): void {
    this.closeDrawer();
  }
}

function readCollapsed(): boolean {
  try {
    return localStorage.getItem(COLLAPSE_KEY) === 'true';
  } catch {
    return false;
  }
}
