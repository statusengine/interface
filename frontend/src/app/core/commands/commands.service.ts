import { Injectable, inject, signal } from '@angular/core';
import { TranslocoService } from '@jsverse/transloco';
import { Api } from '../api/api.service';
import { ApiError } from '../api/api.error';
import { Toasts } from '../toast/toast.service';
import type { HostStatus, Kind, ServiceStatus } from '../api/types';

/** One object a command applies to. */
export interface CommandTarget {
  kind: Kind;
  host: string;
  service?: string;
}

/** What the server answers to a submission. */
export interface CommandAck {
  status: string;
  action: string;
  /** How many objects the command was aimed at. */
  submitted: number;
  /** How many Naemon commands that became - a host downtime covering
   *  its services is two. */
  commands: number;
  accepted: number;
  targets: string[];
  /** The objects worth polling. Empty when there are too many for that
   *  to be cheaper than waiting for the list to refresh. */
  verify: CommandTarget[];
  note: string;
}

/** A predicate over the object a command was aimed at. */
export type Verifier = (status: HostStatus | ServiceStatus) => boolean;

export interface RunOptions {
  /** Endpoint under /commands. */
  action: string;
  /** The objects to act on. One is a list of one. */
  targets: CommandTarget[];
  /** Everything else the command needs. */
  body: Record<string, unknown>;
  /** Shown while the command is in flight and while it is being
   *  confirmed. */
  pending: string;
  /** Shown once the monitoring core has visibly applied it. */
  success: string;
  /**
   * How to tell the command took. Without one, the toast settles at
   * "submitted" - which is the honest ceiling for a command whose
   * effect this interface cannot observe, such as a custom
   * notification.
   */
  verify?: Verifier;
}

/** How long to watch for confirmation, and how often. */
const VERIFY_ATTEMPTS = 6;
const VERIFY_INTERVAL_MS = 2000;

/**
 * Submitting external commands, and finding out whether they took.
 *
 * A 202 from the worker means the command reached the message broker.
 * It does not mean Naemon ran it: the queue acknowledges the publish,
 * the broker module has no reply path, and it logs nothing for a
 * command it does not recognise. So the UI says "submitted", then
 * watches the object for a few seconds and only then says "confirmed".
 * If confirmation does not arrive, it says that too, rather than
 * claiming a success it cannot see.
 */
@Injectable({ providedIn: 'root' })
export class Commands {
  private readonly api = inject(Api);
  private readonly toasts = inject(Toasts);
  private readonly transloco = inject(TranslocoService);

  /** Bumped whenever a command is accepted, so open lists can refetch. */
  private readonly _submitted = signal(0);
  readonly submitted = this._submitted.asReadonly();

  private readonly _busy = signal(false);
  readonly busy = this._busy.asReadonly();

  /**
   * Runs a command. Resolves true when the server accepted it, which is
   * what a dialog needs to know in order to close; confirmation
   * continues in the background and lands in the toast.
   */
  async run(options: RunOptions): Promise<boolean> {
    const toastId = this.toasts.show('pending', options.pending);
    this._busy.set(true);

    let ack: CommandAck;
    try {
      ack = await this.api.post<CommandAck>(`/commands/${options.action}`, {
        ...options.body,
        targets: options.targets,
      });
    } catch (err) {
      const error = ApiError.from(err);
      this.toasts.replace(
        toastId,
        'error',
        this.transloco.translate('commands.failed'),
        error.message,
      );
      return false;
    } finally {
      this._busy.set(false);
    }

    this._submitted.update((n) => n + 1);

    if (!options.verify || ack.verify.length === 0) {
      this.toasts.replace(
        toastId,
        'success',
        options.success,
        this.transloco.translate('commands.submittedNote'),
      );
      return true;
    }

    this.toasts.replace(toastId, 'pending', this.transloco.translate('commands.confirming'));
    void this.confirm(toastId, ack, options);
    return true;
  }

  private async confirm(toastId: number, ack: CommandAck, options: RunOptions): Promise<void> {
    // Every object has to show the change, not just the first: a bulk
    // that took on three of four is not done.
    const pending = [...ack.verify];

    for (let attempt = 0; attempt < VERIFY_ATTEMPTS; attempt++) {
      await sleep(VERIFY_INTERVAL_MS);
      try {
        const results = await Promise.all(
          pending.map(async (target) => ({
            target,
            done: options.verify!(await this.fetchTarget(target)),
          })),
        );
        for (const { target, done } of results) {
          if (done) {
            pending.splice(pending.indexOf(target), 1);
          }
        }
        if (pending.length === 0) {
          this.toasts.replace(toastId, 'success', options.success);
          this._submitted.update((n) => n + 1);
          return;
        }
      } catch {
        // A failed poll is not a failed command. Keep watching; the
        // timeout message below covers giving up.
      }
    }

    // Not a failure. The command reached the broker; the core has not
    // visibly acted on it within the window we were willing to watch,
    // and saying so is more useful than a green tick that might be
    // wrong.
    this.toasts.replace(
      toastId,
      'warning',
      this.transloco.translate('commands.notConfirmed'),
      this.transloco.translate('commands.notConfirmedNote', {
        target: pending.map((t) => (t.service ? `${t.host}/${t.service}` : t.host)).join(', '),
      }),
    );
  }

  private fetchTarget(target: CommandTarget): Promise<HostStatus | ServiceStatus> {
    if (target.kind === 'service' && target.service) {
      return this.api.get<ServiceStatus>('/service', {
        host: target.host,
        service: target.service,
      });
    }
    return this.api.get<HostStatus>(`/hosts/${encodeURIComponent(target.host)}`);
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
