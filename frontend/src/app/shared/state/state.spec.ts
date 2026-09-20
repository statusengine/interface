import { describe, expect, it } from 'vitest';
import { HOST_STATES, SERVICE_STATES, severity, stateClass, stateTextClass } from './state';

describe('state colours', () => {
  it('gives a host and a service the same colour for the same meaning', () => {
    // UP and OK are the same news; DOWN and CRITICAL are the same news.
    // An operator scanning a mixed problem list should not have to
    // translate between two palettes.
    expect(stateClass('up')).toBe(stateClass('ok'));
    expect(stateClass('down')).toBe(stateClass('critical'));
    expect(stateTextClass('up')).toBe(stateTextClass('ok'));
  });

  // Nagios uses orange for UNKNOWN next to a yellow WARNING, which is the
  // most misread pair in the classic interface - and the two mean very
  // different things.
  it('keeps unknown clearly apart from warning', () => {
    expect(stateClass('unknown')).not.toBe(stateClass('warning'));
    expect(stateTextClass('unknown')).not.toBe(stateTextClass('warning'));
  });

  it('keeps unreachable apart from down', () => {
    expect(stateClass('unreachable')).not.toBe(stateClass('down'));
  });

  it('falls back to pending for anything it does not know', () => {
    expect(stateClass('something-new')).toBe('state-pending');
    expect(stateClass('')).toBe('state-pending');
  });
});

describe('severity', () => {
  // The client-side scale has to agree with the backend's CASE
  // expressions, or a locally sorted list disagrees with a server-sorted
  // one for the same data.
  it('ranks a host that is down above everything else', () => {
    expect(severity('host', 1)).toBeGreaterThan(severity('service', 2));
    expect(severity('host', 1)).toBeGreaterThan(severity('host', 2));
  });

  it('ranks unreachable below a critical service', () => {
    // An unreachable host is usually a consequence of some other host
    // being down; a critical service is its own problem.
    expect(severity('host', 2)).toBeLessThan(severity('service', 2));
  });

  it('ranks service states the way people triage, not by state number', () => {
    const critical = severity('service', 2);
    const warning = severity('service', 1);
    const unknown = severity('service', 3);
    const ok = severity('service', 0);

    expect(critical).toBeGreaterThan(warning);
    expect(warning).toBeGreaterThan(unknown);
    expect(unknown).toBeGreaterThan(ok);
  });

  it('gives an OK state no severity at all', () => {
    expect(severity('host', 0)).toBe(0);
    expect(severity('service', 0)).toBe(0);
  });
});

describe('state filter options', () => {
  it('covers every state the core can report', () => {
    expect(HOST_STATES.map((s) => s.value)).toEqual([0, 1, 2]);
    expect(SERVICE_STATES.map((s) => s.value)).toEqual([0, 1, 2, 3]);
  });

  it('gives every option a colour', () => {
    for (const option of [...HOST_STATES, ...SERVICE_STATES]) {
      expect(stateClass(option.key)).not.toBe('state-pending');
    }
  });
});
