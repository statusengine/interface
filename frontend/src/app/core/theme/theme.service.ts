import { Injectable, effect, signal } from '@angular/core';

export type ThemePreference = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'sei.theme';

/**
 * Light, dark, or whatever the operating system says.
 *
 * Dark is not the default here by taste: this interface spends its life on
 * wall displays and night shifts. It is still a preference, because the
 * same tables get read in a lit office.
 */
@Injectable({ providedIn: 'root' })
export class Theme {
  private readonly _preference = signal<ThemePreference>(read());

  readonly preference = this._preference.asReadonly();

  constructor() {
    effect(() => {
      const pref = this._preference();
      const root = document.documentElement;

      if (pref === 'system') {
        root.removeAttribute('data-theme');
      } else {
        root.setAttribute('data-theme', pref);
      }

      try {
        localStorage.setItem(STORAGE_KEY, pref);
      } catch {
        // Private windows and blocked site data throw here. The theme
        // still applies for this visit; it just will not be remembered.
      }
    });
  }

  set(pref: ThemePreference): void {
    this._preference.set(pref);
  }

  /** Steps light -> dark -> system, for a single toggle button. */
  cycle(): void {
    const order: ThemePreference[] = ['light', 'dark', 'system'];
    const next = order[(order.indexOf(this._preference()) + 1) % order.length];
    this._preference.set(next);
  }

  /** What is actually on screen right now, resolving `system`. */
  resolved(): 'light' | 'dark' {
    const pref = this._preference();
    if (pref !== 'system') {
      return pref;
    }
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
}

function read(): ThemePreference {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === 'light' || stored === 'dark' || stored === 'system') {
      return stored;
    }
  } catch {
    // Unreadable storage is the same as no stored preference.
  }
  return 'system';
}
