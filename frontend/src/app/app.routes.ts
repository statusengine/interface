import type { Routes } from '@angular/router';
import { anonymousGuard, authGuard, permissionGuard } from './core/auth/auth.guard';
import { PlaceholderPage } from './features/placeholder-page';
import { Shell } from './layout/shell/shell';
import type { Permission } from './core/api/types';

/**
 * Pages scheduled for a later phase are routed now, to the placeholder,
 * so the navigation rail is complete from the first build and nothing in
 * it dead-ends.
 */
function placeholder(path: string, titleKey: string, phase: number, needs?: Permission) {
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

      {
        path: 'dashboard',
        loadComponent: () => import('./features/dashboard/dashboard').then((m) => m.Dashboard),
      },
      {
        path: 'problems',
        canActivate: [permissionGuard('problems:read')],
        loadComponent: () =>
          import('./features/problems/problems-list').then((m) => m.ProblemsList),
      },
      {
        path: 'hosts',
        canActivate: [permissionGuard('hosts:read')],
        loadComponent: () => import('./features/hosts/hosts-list').then((m) => m.HostsList),
      },
      {
        path: 'hosts/:host',
        canActivate: [permissionGuard('hosts:read')],
        loadComponent: () => import('./features/hosts/host-detail').then((m) => m.HostDetail),
      },
      {
        path: 'services',
        canActivate: [permissionGuard('services:read')],
        loadComponent: () =>
          import('./features/services/services-list').then((m) => m.ServicesList),
      },
      {
        // Singular, and identified by query parameters: a Naemon service
        // description is free text and routinely contains slashes.
        path: 'service',
        canActivate: [permissionGuard('services:read')],
        loadComponent: () =>
          import('./features/services/service-detail').then((m) => m.ServiceDetail),
      },
      {
        path: 'downtimes',
        canActivate: [permissionGuard('downtimes:read')],
        data: { history: false },
        loadComponent: () =>
          import('./features/downtimes/downtimes-list').then((m) => m.DowntimesList),
      },
      {
        path: 'downtimes/history',
        canActivate: [permissionGuard('downtimes:read')],
        data: { history: true },
        loadComponent: () =>
          import('./features/downtimes/downtimes-list').then((m) => m.DowntimesList),
      },
      {
        path: 'acknowledgements',
        canActivate: [permissionGuard('acknowledgements:read')],
        loadComponent: () =>
          import('./features/acknowledgements/acknowledgements-list').then(
            (m) => m.AcknowledgementsList,
          ),
      },
      {
        path: 'logentries',
        canActivate: [permissionGuard('logentries:read')],
        loadComponent: () =>
          import('./features/logentries/logentries-list').then((m) => m.LogEntriesList),
      },

      {
        path: 'history/checks',
        canActivate: [permissionGuard('history:read')],
        loadComponent: () =>
          import('./features/history/checks-history').then((m) => m.ChecksHistory),
      },
      {
        path: 'history/statechanges',
        canActivate: [permissionGuard('history:read')],
        loadComponent: () =>
          import('./features/history/statechanges-history').then((m) => m.StateChangesHistory),
      },
      {
        path: 'history/notifications',
        canActivate: [permissionGuard('history:read')],
        loadComponent: () =>
          import('./features/history/notifications-history').then((m) => m.NotificationsHistory),
      },

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
