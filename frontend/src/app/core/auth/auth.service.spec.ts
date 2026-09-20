import { TestBed } from '@angular/core/testing';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { Api } from '../api/api.service';
import { ApiError } from '../api/api.error';
import type { Identity } from '../api/types';
import { Auth } from './auth.service';

const operator: Identity = {
  username: 'ops',
  display_name: 'Ops',
  email: '',
  role: 'operator',
  permissions: ['hosts:read', 'commands:acknowledge'],
  is_demo: false,
  read_only: false,
  expires_at: 0,
};

const guest: Identity = {
  ...operator,
  username: 'guest',
  display_name: 'Demo',
  role: 'guest',
  permissions: ['hosts:read'],
  is_demo: true,
  read_only: true,
};

function setup(api: Partial<Api>) {
  TestBed.configureTestingModule({
    providers: [Auth, { provide: Api, useValue: api }],
  });
  return TestBed.inject(Auth);
}

describe('Auth', () => {
  beforeEach(() => TestBed.resetTestingModule());

  it('starts unresolved and signed out', () => {
    const auth = setup({ get: vi.fn() });

    expect(auth.resolved()).toBe(false);
    expect(auth.isAuthenticated()).toBe(false);
    // Nothing is known yet, so the safe answer is read-only.
    expect(auth.isReadOnly()).toBe(true);
  });

  it('restores a session', async () => {
    const auth = setup({ get: vi.fn().mockResolvedValue(operator) });

    await auth.restore();

    expect(auth.resolved()).toBe(true);
    expect(auth.isAuthenticated()).toBe(true);
    expect(auth.displayName()).toBe('Ops');
    expect(auth.can('commands:acknowledge')).toBe(true);
    expect(auth.can('users:manage')).toBe(false);
  });

  // A 401 from /auth/me is the normal answer for "nobody is signed in",
  // not a failure worth surfacing.
  it('treats a 401 on restore as signed out and still resolves', async () => {
    const auth = setup({
      get: vi.fn().mockRejectedValue(new ApiError(401, 'unauthorized', 'nope')),
    });

    await auth.restore();

    expect(auth.resolved()).toBe(true);
    expect(auth.isAuthenticated()).toBe(false);
  });

  it('resolves even when the server errors, so a guard cannot hang', async () => {
    const auth = setup({
      get: vi.fn().mockRejectedValue(new ApiError(500, 'internal_error', 'boom')),
    });

    await auth.restore();

    expect(auth.resolved()).toBe(true);
    expect(auth.isAuthenticated()).toBe(false);
  });

  it('reports the demo account as read-only', async () => {
    const auth = setup({ post: vi.fn().mockResolvedValue(guest) });

    await auth.loginDemo();

    expect(auth.isDemo()).toBe(true);
    expect(auth.isReadOnly()).toBe(true);
    expect(auth.can('commands:acknowledge')).toBe(false);
  });

  // Whatever the server says on logout, the client must end up signed out;
  // otherwise a network blip leaves a UI that thinks it still has rights.
  it('signs out locally even if the request fails', async () => {
    const auth = setup({
      get: vi.fn().mockResolvedValue(operator),
      postVoid: vi.fn().mockRejectedValue(new ApiError(0, 'network_unreachable', 'offline')),
    });
    await auth.restore();
    expect(auth.isAuthenticated()).toBe(true);

    await expect(auth.logout()).rejects.toThrow();

    expect(auth.isAuthenticated()).toBe(false);
  });
});

describe('Auth.restore during an outage', () => {
  beforeEach(() => TestBed.resetTestingModule());

  // The regression: a database outage used to look exactly like being
  // signed out, which bounced everyone to a login page that could not
  // work either.
  it('does not report being signed out when it could not check', async () => {
    const auth = setup({
      get: vi.fn().mockRejectedValue(new ApiError(503, 'unavailable', 'dependency is down')),
    });

    await auth.restore();

    expect(auth.resolved()).toBe(true);
    expect(auth.unavailable()).toBe(true);
  });

  it('reports a 401 as plainly signed out', async () => {
    const auth = setup({
      get: vi.fn().mockRejectedValue(new ApiError(401, 'unauthorized', 'nope')),
    });

    await auth.restore();

    expect(auth.isAuthenticated()).toBe(false);
    expect(auth.unavailable()).toBe(false);
  });

  it('clears the outage flag once the server answers again', async () => {
    const get = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(503, 'unavailable', 'down'))
      .mockResolvedValue(operator);
    const auth = setup({ get });

    await auth.restore();
    expect(auth.unavailable()).toBe(true);

    await auth.restore();
    expect(auth.unavailable()).toBe(false);
    expect(auth.isAuthenticated()).toBe(true);
  });
});
