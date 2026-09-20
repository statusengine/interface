import { inject } from '@angular/core';
import { TranslocoService } from '@jsverse/transloco';
import type { TimeWords } from './time-format';

/**
 * The wording and the locale the time pipes need, read fresh each time.
 *
 * Reading the active language per call rather than caching it is what
 * makes a language switch take effect: Transloco re-renders the blocks
 * these pipes are used in, and a pure pipe runs again with whatever this
 * returns now.
 */
export function injectTimeContext(): () => { words: TimeWords; locale: string } {
  const transloco = inject(TranslocoService);
  return () => ({
    locale: transloco.getActiveLang(),
    words: {
      never: transloco.translate('time.never'),
      justNow: transloco.translate('time.justNow'),
      ahead: (duration: string) => transloco.translate('time.ahead', { duration }),
    },
  });
}
