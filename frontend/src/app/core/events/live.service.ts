import { DestroyRef, Injectable, computed, inject, signal } from '@angular/core';
import { ServerInfo } from '../meta/meta.service';

/** One object that changed. */
export interface Change {
  topic: string;
  kind?: 'host' | 'service';
  host?: string;
  service?: string;
}

export interface ChangeBatch {
  changes: Change[];
  topics: string[];
  at: number;
  /** Set when the server coalesced away more than it could carry; the
   *  list is then incomplete and a broad refetch is the right response. */
  dropped?: number;
}

/**
 * How the UI is currently learning about changes.
 *
 * `live` is the event stream. `polling` is the fallback, and it is a
 * working state rather than an error - a deployment with no worker
 * event key never leaves it. `offline` means neither is getting
 * through.
 */
export type LiveMode = 'live' | 'polling' | 'offline' | 'idle';

/** How often to refetch when the stream is not available. */
const POLL_INTERVAL_MS = 30_000;

/** How long without any event before the stream counts as suspect. */
const SILENCE_BEFORE_POLLING_MS = 90_000;

/**
 * Live change notifications, with a fallback that is not a failure.
 *
 * The stream says what changed, never what it changed to: consumers
 * refetch the endpoint they are already rendering. One source of truth,
 * and a pushed payload can never drift from what the REST endpoint
 * returns.
 */
@Injectable({ providedIn: 'root' })
export class Live {
  private readonly server = inject(ServerInfo);
  private readonly destroyRef = inject(DestroyRef);

  private source?: EventSource;
  private pollTimer?: ReturnType<typeof setInterval>;
  private silenceTimer?: ReturnType<typeof setTimeout>;

  private readonly _mode = signal<LiveMode>('idle');
  private readonly _lastBatch = signal<ChangeBatch | null>(null);
  private readonly _tick = signal(0);

  /** How changes are arriving right now. */
  readonly mode = this._mode.asReadonly();

  /** The most recent batch, for a consumer that wants to be selective. */
  readonly lastBatch = this._lastBatch.asReadonly();

  /**
   * Increments whenever anything may have changed - a batch arrived, or
   * the poll interval elapsed. A list can watch this alone and refetch,
   * without caring which of the two happened.
   */
  readonly tick = this._tick.asReadonly();

  readonly isLive = computed(() => this._mode() === 'live');

  constructor() {
    this.destroyRef.onDestroy(() => this.stop());
  }

  /** Starts listening. Safe to call more than once. */
  start(): void {
    if (this.source || this.pollTimer) {
      return;
    }
    if (!this.server.eventsEnabled) {
      // No worker event key on this deployment. Polling is the whole
      // story here, not a degraded mode.
      this.startPolling();
      return;
    }
    this.connect();
  }

  stop(): void {
    this.source?.close();
    this.source = undefined;
    this.stopPolling();
    clearTimeout(this.silenceTimer);
    this._mode.set('idle');
  }

  private connect(): void {
    // withCredentials so the session cookie travels: this is why the
    // stream is SSE and not a browser WebSocket, which cannot carry a
    // header on its handshake.
    const source = new EventSource('/api/v1/events', { withCredentials: true });
    this.source = source;

    source.addEventListener('hello', (event) => {
      const payload = parse<{ connected: boolean }>(event);
      this.stopPolling();
      this._mode.set(payload?.connected ? 'live' : 'polling');
      if (!payload?.connected) {
        // The stream is up but the worker is not feeding it, so nothing
        // will arrive until it comes back.
        this.startPolling();
      }
      this.armSilenceTimer();
    });

    source.addEventListener('upstream', (event) => {
      const payload = parse<{ connected: boolean }>(event);
      if (payload?.connected) {
        this.stopPolling();
        this._mode.set('live');
      } else {
        this._mode.set('polling');
        this.startPolling();
      }
    });

    source.addEventListener('changes', (event) => {
      const batch = parse<ChangeBatch>(event);
      if (!batch) {
        return;
      }
      this._lastBatch.set(batch);
      this._tick.update((n) => n + 1);
      this._mode.set('live');
      this.stopPolling();
      this.armSilenceTimer();
    });

    source.onerror = () => {
      // EventSource reconnects on its own, so this is not the moment to
      // give up - but it is the moment to start polling, so the UI keeps
      // updating while the browser retries.
      this._mode.set(source.readyState === EventSource.CLOSED ? 'offline' : 'polling');
      this.startPolling();
    };
  }

  /**
   * A stream that has been silent for a long time is indistinguishable
   * from a quiet monitoring system - except that one of them is broken.
   * Polling alongside it costs one request a minute and removes the
   * ambiguity.
   */
  private armSilenceTimer(): void {
    clearTimeout(this.silenceTimer);
    this.silenceTimer = setTimeout(() => {
      this._mode.set('polling');
      this.startPolling();
    }, SILENCE_BEFORE_POLLING_MS);
  }

  private startPolling(): void {
    if (this.pollTimer) {
      return;
    }
    if (this._mode() === 'idle') {
      this._mode.set('polling');
    }
    this.pollTimer = setInterval(() => this._tick.update((n) => n + 1), POLL_INTERVAL_MS);
  }

  private stopPolling(): void {
    if (this.pollTimer) {
      clearInterval(this.pollTimer);
      this.pollTimer = undefined;
    }
  }
}

function parse<T>(event: Event): T | null {
  const data = (event as MessageEvent<string>).data;
  if (!data) {
    return null;
  }
  try {
    return JSON.parse(data) as T;
  } catch {
    return null;
  }
}
