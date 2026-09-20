import { TestBed } from '@angular/core/testing';
import { TranslocoService } from '@jsverse/transloco';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { Api } from '../api/api.service';
import { ApiError } from '../api/api.error';
import { Toasts } from '../toast/toast.service';
import { Commands, type CommandTarget } from './commands.service';

const host: CommandTarget = { kind: 'host', host: 'db01' };

const ack = {
  status: 'submitted',
  action: 'acknowledge',
  submitted: 1,
  commands: 1,
  accepted: 1,
  targets: ['db01'],
  verify: [host],
  note: 'reached the broker',
};

function setup(api: Partial<Api>) {
  TestBed.resetTestingModule();
  TestBed.configureTestingModule({
    providers: [
      Commands,
      Toasts,
      { provide: Api, useValue: api },
      // The real service needs its loader; the keys are echoed back,
      // which is all these assertions need.
      { provide: TranslocoService, useValue: { translate: (key: string) => key } },
    ],
  });
  return { commands: TestBed.inject(Commands), toasts: TestBed.inject(Toasts) };
}

/** Lets the pending promises settle under fake timers. */
async function settle() {
  await vi.advanceTimersByTimeAsync(0);
}

describe('Commands', () => {
  beforeEach(() => vi.useFakeTimers());
  afterEach(() => vi.useRealTimers());

  it('reports a submission the server refused', async () => {
    const post = vi.fn().mockRejectedValue(new ApiError(403, 'forbidden', 'not your permission'));
    const { commands, toasts } = setup({ post });

    const ok = await commands.run({
      targets: [host],
      action: 'acknowledge',
      body: {},
      pending: 'p',
      success: 's',
    });

    expect(ok).toBe(false);
    expect(toasts.items()[0]).toMatchObject({ tone: 'error', body: 'not your permission' });
    expect(commands.busy()).toBe(false);
  });

  // Nothing observable changes on the object for a custom notification,
  // so claiming it was delivered would be a guess.
  it('settles at submitted when there is nothing to verify', async () => {
    const post = vi.fn().mockResolvedValue(ack);
    const { commands, toasts } = setup({ post });

    const ok = await commands.run({
      targets: [host],
      action: 'notify',
      body: {},
      pending: 'p',
      success: 'sent',
    });

    expect(ok).toBe(true);
    expect(toasts.items()[0]).toMatchObject({ tone: 'success', title: 'sent' });
  });

  it('confirms once the object has actually changed', async () => {
    const post = vi.fn().mockResolvedValue(ack);
    const get = vi
      .fn()
      .mockResolvedValueOnce({ acknowledged: false })
      .mockResolvedValue({ acknowledged: true });
    const { commands, toasts } = setup({ post, get });

    await commands.run({
      targets: [host],
      action: 'acknowledge',
      body: {},
      pending: 'p',
      success: 'acknowledged',
      verify: (status) => (status as { acknowledged: boolean }).acknowledged,
    });

    // While the poll is still running the toast says so, rather than
    // claiming a success the interface has not seen.
    expect(toasts.items()[0]).toMatchObject({ tone: 'pending' });

    await vi.advanceTimersByTimeAsync(2000);
    await settle();
    expect(toasts.items()[0]).toMatchObject({ tone: 'pending' });

    await vi.advanceTimersByTimeAsync(2000);
    await settle();
    expect(toasts.items()[0]).toMatchObject({ tone: 'success', title: 'acknowledged' });
  });

  // The command reached the broker; the core has not acted on it within
  // the window. That is not a failure and not a success.
  it('says so when confirmation never arrives', async () => {
    const post = vi.fn().mockResolvedValue(ack);
    const get = vi.fn().mockResolvedValue({ acknowledged: false });
    const { commands, toasts } = setup({ post, get });

    await commands.run({
      targets: [host],
      action: 'acknowledge',
      body: {},
      pending: 'p',
      success: 'acknowledged',
      verify: (status) => (status as { acknowledged: boolean }).acknowledged,
    });

    // Six polls two seconds apart, and then check before the warning's
    // own lifetime runs out.
    await vi.advanceTimersByTimeAsync(6 * 2000 + 100);
    await settle();

    expect(toasts.items()[0]).toMatchObject({ tone: 'warning', title: 'commands.notConfirmed' });
  });

  it('keeps watching when a poll itself fails', async () => {
    const post = vi.fn().mockResolvedValue(ack);
    const get = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(0, 'network_unreachable', 'offline'))
      .mockResolvedValue({ acknowledged: true });
    const { commands, toasts } = setup({ post, get });

    await commands.run({
      targets: [host],
      action: 'acknowledge',
      body: {},
      pending: 'p',
      success: 'acknowledged',
      verify: (status) => (status as { acknowledged: boolean }).acknowledged,
    });

    await vi.advanceTimersByTimeAsync(4000);
    await settle();

    expect(toasts.items()[0]).toMatchObject({ tone: 'success' });
  });

  it('polls the service endpoint for a service target', async () => {
    const post = vi.fn().mockResolvedValue({
      ...ack,
      verify: [{ kind: 'service' as const, host: 'db01', service: 'C:\\ Drive Space' }],
    });
    const get = vi.fn().mockResolvedValue({ acknowledged: true });
    const { commands } = setup({ post, get });

    await commands.run({
      targets: [host],
      action: 'acknowledge',
      body: {},
      pending: 'p',
      success: 's',
      verify: () => true,
    });
    await vi.advanceTimersByTimeAsync(2000);
    await settle();

    expect(get).toHaveBeenCalledWith('/service', { host: 'db01', service: 'C:\\ Drive Space' });
  });

  it('bumps the change counter so open lists refetch', async () => {
    const post = vi.fn().mockResolvedValue(ack);
    const { commands } = setup({ post });

    const before = commands.submitted();
    await commands.run({ targets: [host], action: 'notify', body: {}, pending: 'p', success: 's' });

    expect(commands.submitted()).toBeGreaterThan(before);
  });
});
