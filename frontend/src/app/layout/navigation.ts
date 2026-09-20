import type { IconName } from '../shared/ui/icon';
import type { Permission } from '../core/api/types';

export interface NavItem {
  /** Router path, absolute. */
  path: string;
  /** Translation key under `nav.`. */
  label: string;
  icon: IconName;
  /** Hidden when the signed-in role does not hold this. */
  needs?: Permission;
}

export interface NavGroup {
  /** Translation key under `nav.groups.`, or null for the top block. */
  label: string | null;
  items: NavItem[];
}

/**
 * The navigation, grouped the way an operator's day is shaped rather than
 * the way the database is: what is wrong now, what exists, what we did
 * about it, what happened.
 */
export const NAVIGATION: NavGroup[] = [
  {
    label: null,
    items: [
      { path: '/dashboard', label: 'dashboard', icon: 'dashboard' },
      { path: '/problems', label: 'problems', icon: 'problems', needs: 'problems:read' },
    ],
  },
  {
    label: 'monitoring',
    items: [
      { path: '/hosts', label: 'hosts', icon: 'host', needs: 'hosts:read' },
      { path: '/services', label: 'services', icon: 'service', needs: 'services:read' },
    ],
  },
  {
    label: 'operations',
    items: [
      { path: '/downtimes', label: 'downtimes', icon: 'downtime', needs: 'downtimes:read' },
      {
        path: '/acknowledgements',
        label: 'acknowledgements',
        icon: 'acknowledge',
        needs: 'acknowledgements:read',
      },
    ],
  },
  {
    label: 'records',
    items: [
      { path: '/logentries', label: 'logentries', icon: 'log', needs: 'logentries:read' },
      { path: '/history/checks', label: 'historyChecks', icon: 'history', needs: 'history:read' },
      {
        path: '/history/statechanges',
        label: 'historyStateChanges',
        icon: 'history',
        needs: 'history:read',
      },
      {
        path: '/history/notifications',
        label: 'historyNotifications',
        icon: 'history',
        needs: 'history:read',
      },
    ],
  },
];
