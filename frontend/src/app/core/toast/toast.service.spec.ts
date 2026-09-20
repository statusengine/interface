import { TestBed } from '@angular/core/testing';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { Toasts } from './toast.service';

describe('Toasts', () => {
  let toasts: Toasts;

  beforeEach(() => {
    vi.useFakeTimers();
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({ providers: [Toasts] });
    toasts = TestBed.inject(Toasts);
  });

  afterEach(() => vi.useRealTimers());

  it('dismisses a transient toast by itself', () => {
    toasts.show('success', 'done');
    expect(toasts.items()).toHaveLength(1);

    vi.advanceTimersByTime(5000);
    expect(toasts.items()).toHaveLength(0);
  });

  // An error is the one the operator has to read, and the one most
  // likely to appear while they are looking elsewhere.
  it('keeps an error until it is dismissed', () => {
    const id = toasts.show('error', 'that did not work');

    vi.advanceTimersByTime(60_000);
    expect(toasts.items()).toHaveLength(1);

    toasts.dismiss(id);
    expect(toasts.items()).toHaveLength(0);
  });

  it('keeps a pending toast until its outcome replaces it', () => {
    toasts.show('pending', 'submitting');
    vi.advanceTimersByTime(60_000);
    expect(toasts.items()).toHaveLength(1);
  });

  // The sequence should read as one event, not as two unrelated ones
  // stacking up.
  it('replaces in place rather than adding a second toast', () => {
    const id = toasts.show('pending', 'submitting');
    toasts.replace(id, 'success', 'done');

    expect(toasts.items()).toHaveLength(1);
    expect(toasts.items()[0]).toMatchObject({ id, tone: 'success', title: 'done' });
  });

  it('keeps its position in the stack when replaced', () => {
    const first = toasts.show('pending', 'first');
    toasts.show('info', 'second');

    toasts.replace(first, 'success', 'first done');

    expect(toasts.items().map((t) => t.title)).toEqual(['first done', 'second']);
  });

  it('re-arms the timer when a pending toast becomes transient', () => {
    const id = toasts.show('pending', 'submitting');
    toasts.replace(id, 'success', 'done');

    vi.advanceTimersByTime(5000);
    expect(toasts.items()).toHaveLength(0);
  });

  // A toast that timed out before its outcome arrived must not make the
  // outcome disappear with it.
  it('shows a replacement for a toast that is already gone', () => {
    const id = toasts.show('success', 'done');
    vi.advanceTimersByTime(5000);
    expect(toasts.items()).toHaveLength(0);

    toasts.replace(id, 'warning', 'actually, not confirmed');
    expect(toasts.items()).toHaveLength(1);
    expect(toasts.items()[0].title).toBe('actually, not confirmed');
  });

  it('clears everything at once', () => {
    toasts.show('info', 'a');
    toasts.show('error', 'b');
    toasts.dismissAll();
    expect(toasts.items()).toHaveLength(0);
  });
});
