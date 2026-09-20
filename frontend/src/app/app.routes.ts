import type { Routes } from '@angular/router';
import { anonymousGuard, authGuard, permissionGuard } from './core/auth/auth.guard';
import { PlaceholderPage } from './features/placeholder-page';
import { Shell } from './layout/shell/shell';

/**
 * Pages scheduled for a later phase are routed now, to the placeholder,
 * so the navigation rail is complete from the first build and nothing in
 * it dead-ends.
 */
function placeholder(
  path: string,
  titleKey: string,
  phase: number,
  needs?: Parameters<typeof permissionGuard>[0],
) {
  return {
    path,
    component: PlaceholderPage,
    data: { titleKey, phase },
    ...(needs ? { canActivate: [permissionGuard(needs)] } : {}),
  };
}

export const routes: Routes = [
  {
    path: 'login',
    canActivate: [anonymousGuard],
    loadComponent: () => import('./features/auth/login/login').then((m) => m.Login),
  },
  {
    path: '',
    component: Shell,
    canActivate: [authGuard],
    children: [
      { path: '', pathMatch: 'full', redirectTo: 'dashboard' },

      placeholder('dashboard', 'dashboard', 2),
      placeholder('problems', 'problems', 2, 'problems:read'),
      placeholder('hosts', 'hosts', 2, 'hosts:read'),
      placeholder('services', 'services', 2, 'services:read'),
      placeholder('downtimes', 'downtimes', 2, 'downtimes:read'),
      placeholder('acknowledgements', 'acknowledgements', 2, 'acknowledgements:read'),
      placeholder('logentries', 'logentries', 2, 'logentries:read'),
      placeholder('history/checks', 'historyChecks', 3, 'history:read'),
      placeholder('history/statechanges', 'historyStateChanges', 3, 'history:read'),
      placeholder('history/notifications', 'historyNotifications', 3, 'history:read'),

      {
        path: 'forbidden',
        loadComponent: () => import('./features/errors/forbidden').then((m) => m.Forbidden),
      },
      {
        path: '**',
        loadComponent: () => import('./features/errors/not-found').then((m) => m.NotFound),
      },
    ],
  },
];
