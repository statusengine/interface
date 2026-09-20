import { Injectable, signal } from '@angular/core';

export type ToastTone = 'info' | 'success' | 'warning' | 'error' | 'pending';

export interface Toast {
  id: number;
  tone: ToastTone;
  title: string;
  body?: string;
  /** Milliseconds before it dismisses itself; 0 stays until dismissed. */
  ttl: number;
}

const DEFAULT_TTL: Record<ToastTone, number> = {
  info: 5000,
  success: 5000,
  warning: 8000,
  // An error stays. It is the one the operator has to read, and it is
  // the one most likely to appear while they are looking elsewhere.
  error: 0,
  // A pending toast is replaced by its outcome rather than timing out.
  pending: 0,
};

/**
 * Short-lived messages about things the operator did.
 *
 * A submitted command produces a `pending` toast that is later replaced
 * in place by its outcome, so the sequence reads as one event rather
 * than two unrelated ones stacking up.
 */
@Injectable({ providedIn: 'root' })
export class Toasts {
  private readonly _items = signal<Toast[]>([]);
  private nextId = 1;
  private readonly timers = new Map<number, ReturnType<typeof setTimeout>>();

  readonly items = this._items.asReadonly();

  /** Shows a toast and returns its id, for replacing it later. */
  show(tone: ToastTone, title: string, body?: string, ttl = DEFAULT_TTL[tone]): number {
    const id = this.nextId++;
    this._items.update((items) => [...items, { id, tone, title, body, ttl }]);
    this.arm(id, ttl);
    return id;
  }

  /** Replaces a toast in place, keeping its position in the stack. */
  replace(
    id: number,
    tone: ToastTone,
    title: string,
    body?: string,
    ttl = DEFAULT_TTL[tone],
  ): void {
    const existing = this._items().some((toast) => toast.id === id);
    if (!existing) {
      this.show(tone, title, body, ttl);
      return;
    }
    this._items.update((items) =>
      items.map((toast) => (toast.id === id ? { ...toast, tone, title, body, ttl } : toast)),
    );
    this.arm(id, ttl);
  }

  dismiss(id: number): void {
    this.clearTimer(id);
    this._items.update((items) => items.filter((toast) => toast.id !== id));
  }

  dismissAll(): void {
    for (const id of this.timers.keys()) {
      this.clearTimer(id);
    }
    this._items.set([]);
  }

  private arm(id: number, ttl: number): void {
    this.clearTimer(id);
    if (ttl > 0) {
      this.timers.set(
        id,
        setTimeout(() => this.dismiss(id), ttl),
      );
    }
  }

  private clearTimer(id: number): void {
    const timer = this.timers.get(id);
    if (timer !== undefined) {
      clearTimeout(timer);
      this.timers.delete(id);
    }
  }
}
