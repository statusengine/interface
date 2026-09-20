import { Injectable, inject, signal } from '@angular/core';
import { TranslocoService } from '@jsverse/transloco';

export const SUPPORTED_LANGUAGES = ['en', 'de'] as const;
export type Language = (typeof SUPPORTED_LANGUAGES)[number];

const STORAGE_KEY = 'sei.lang';

/** Remembers the chosen language and keeps <html lang> honest, which is
 *  what a screen reader reads to pick a voice. */
@Injectable({ providedIn: 'root' })
export class LanguageService {
  private readonly transloco = inject(TranslocoService);
  private readonly _active = signal<Language>(initial());

  readonly active = this._active.asReadonly();
  readonly available = SUPPORTED_LANGUAGES;

  constructor() {
    this.apply(this._active());
  }

  set(lang: Language): void {
    this._active.set(lang);
    this.apply(lang);
    try {
      localStorage.setItem(STORAGE_KEY, lang);
    } catch {
      // Not remembered, still applied.
    }
  }

  private apply(lang: Language): void {
    this.transloco.setActiveLang(lang);
    document.documentElement.lang = lang;
  }
}

function initial(): Language {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (isSupported(stored)) {
      return stored;
    }
  } catch {
    // fall through to the browser's preference
  }
  const browser = navigator.language?.slice(0, 2);
  return isSupported(browser) ? browser : 'en';
}

function isSupported(value: string | null | undefined): value is Language {
  return !!value && (SUPPORTED_LANGUAGES as readonly string[]).includes(value);
}
