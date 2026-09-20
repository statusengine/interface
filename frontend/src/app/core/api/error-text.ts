import type { TranslocoService } from '@jsverse/transloco';
import type { ApiError } from './api.error';

/**
 * What to show a person when a request failed.
 *
 * The server sends a stable code and an English sentence. The code is
 * what we translate; the sentence is the fallback for a code this build
 * has no wording for yet, which is better than showing `errors.conflict`
 * to somebody trying to work.
 */
export function apiErrorText(transloco: TranslocoService, error: ApiError | null): string | null {
  if (!error) {
    return null;
  }
  const translated = transloco.translate(error.translationKey);
  return translated === error.translationKey ? error.message : translated;
}
