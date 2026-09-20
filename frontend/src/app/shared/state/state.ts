/**
 * The state vocabulary, in one place.
 *
 * The API sends both the raw number and a `state_text`, so nothing here
 * has to map numbers. What it does own is the visual: which rail colour a
 * state gets, and in which order states are worked through.
 */

export type StateKey =
  'ok' | 'up' | 'warning' | 'critical' | 'down' | 'unknown' | 'unreachable' | 'pending';

/** The CSS class that sets `--rail-color`. See styles.css. */
export function stateClass(text: string): string {
  switch (text) {
    case 'ok':
    case 'up':
      return 'state-ok';
    case 'warning':
      return 'state-warning';
    case 'critical':
    case 'down':
      return 'state-critical';
    case 'unknown':
      return 'state-unknown';
    case 'unreachable':
      return 'state-unreachable';
    default:
      return 'state-pending';
  }
}

/** Text colour for a state label. */
export function stateTextClass(text: string): string {
  switch (text) {
    case 'ok':
    case 'up':
      return 'text-ok';
    case 'warning':
      return 'text-warning';
    case 'critical':
    case 'down':
      return 'text-critical';
    case 'unknown':
      return 'text-unknown';
    case 'unreachable':
      return 'text-unreachable';
    default:
      return 'text-pending';
  }
}

/**
 * How bad a state is, on one scale across hosts and services. Matches the
 * backend's severity sort so a client-side sort cannot disagree with a
 * server-side one.
 */
export function severity(kind: 'host' | 'service', state: number): number {
  if (kind === 'host') {
    return state === 1 ? 5 : state === 2 ? 3 : 0;
  }
  return state === 2 ? 4 : state === 1 ? 2 : state === 3 ? 1 : 0;
}

/** The state numbers a host can be in, for filter controls. */
export const HOST_STATES: { value: number; key: StateKey }[] = [
  { value: 0, key: 'up' },
  { value: 1, key: 'down' },
  { value: 2, key: 'unreachable' },
];

/** The state numbers a service can be in, for filter controls. */
export const SERVICE_STATES: { value: number; key: StateKey }[] = [
  { value: 0, key: 'ok' },
  { value: 1, key: 'warning' },
  { value: 2, key: 'critical' },
  { value: 3, key: 'unknown' },
];
