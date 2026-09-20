import { inject } from '@angular/core';
import { Router, type CanActivateFn, type UrlTree } from '@angular/router';
import { Auth } from './auth.service';
import type { Permission } from '../api/types';

/**
 * Lets a route through only for a signed-in user. On a cold load the
 * session has not been resolved yet, so this waits for it rather than
 * bouncing someone who is in fact logged in.
 */
export const authGuard: CanActivateFn = async (_route, state): Promise<boolean | UrlTree> => {
  const auth = inject(Auth);
  const router = inject(Router);

  if (!auth.resolved()) {
    await auth.restore();
  }
  if (auth.isAuthenticated()) {
    return true;
  }
  // Remember where they were going, so the login lands them there.
  return router.createUrlTree(['/login'], { queryParams: { next: state.url } });
};

/** Guards a route behind one permission. */
export function permissionGuard(permission: Permission): CanActivateFn {
  return async (route, state): Promise<boolean | UrlTree> => {
    const auth = inject(Auth);
    const router = inject(Router);

    const allowed = await (authGuard(route, state) as Promise<boolean | UrlTree>);
    if (allowed !== true) {
      return allowed;
    }
    if (auth.can(permission)) {
      return true;
    }
    return router.createUrlTree(['/forbidden'], { queryParams: { need: permission } });
  };
}

/** Keeps a signed-in user off the login page. */
export const anonymousGuard: CanActivateFn = async (): Promise<boolean | UrlTree> => {
  const auth = inject(Auth);
  const router = inject(Router);

  if (!auth.resolved()) {
    await auth.restore();
  }
  return auth.isAuthenticated() ? router.createUrlTree(['/']) : true;
};
