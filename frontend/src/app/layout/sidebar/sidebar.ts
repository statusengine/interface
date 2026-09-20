import { ChangeDetectionStrategy, Component, computed, inject, input, output } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';
import { Auth } from '../../core/auth/auth.service';
import { ServerInfo } from '../../core/meta/meta.service';
import { Icon } from '../../shared/ui/icon';
import { NAVIGATION, type NavGroup } from '../navigation';

@Component({
  selector: 'sei-sidebar',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [RouterLink, RouterLinkActive, TranslocoDirective, Icon],
  templateUrl: './sidebar.html',
})
export class Sidebar {
  private readonly auth = inject(Auth);
  private readonly server = inject(ServerInfo);

  /** Icons only, for operators who want the table wider. */
  readonly collapsed = input(false);
  /** True while the rail is an overlay drawer on a narrow screen. */
  readonly drawer = input(false);

  readonly navigate = output<void>();
  readonly toggleCollapsed = output<void>();

  readonly version = computed(() => this.server.version);

  /** Only the entries this role can actually open. A rail full of links
   *  that 403 is a rail that teaches people to ignore it. */
  readonly groups = computed<NavGroup[]>(() => {
    // Read the identity signal so this recomputes on sign-in.
    this.auth.identity();
    return NAVIGATION.map((group) => ({
      ...group,
      items: group.items.filter((item) => !item.needs || this.auth.can(item.needs)),
    })).filter((group) => group.items.length > 0);
  });
}
