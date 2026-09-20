import { describe, expect, it } from 'vitest';
import { percentChange, share } from './dashboard-math';

describe('share', () => {
  it('rounds to whole points', () => {
    expect(share(7, 22)).toBe(32);
    expect(share(1, 4)).toBe(25);
  });

  it('has nothing to say when there is nothing to divide by', () => {
    // An installation with no services is not 0% healthy, it is a
    // question with no answer.
    expect(share(0, 0)).toBeNull();
    expect(share(5, undefined)).toBeNull();
  });

  it('treats a missing part as none of the whole', () => {
    expect(share(undefined, 10)).toBe(0);
  });
});

describe('percentChange', () => {
  it('reads both directions', () => {
    expect(percentChange(4, 3)).toBe(33);
    expect(percentChange(3, 4)).toBe(-25);
    expect(percentChange(4, 4)).toBe(0);
  });

  // The first alert after a quiet period is not a 100% rise, and
  // "infinitely more than zero" is not a number to print.
  it('refuses to compare against nothing', () => {
    expect(percentChange(19, 0)).toBeNull();
    expect(percentChange(0, 0)).toBeNull();
  });
});
