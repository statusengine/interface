import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';

/**
 * The application root is only a router outlet.
 *
 * The skip link lives in the shell, not here: the login page has no
 * main region, so a link to `#main` on every route pointed at nothing
 * from the one page a keyboard user reaches first.
 */
@Component({
  selector: 'app-root',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [RouterOutlet],
  template: `<router-outlet />`,
})
export class App {}
