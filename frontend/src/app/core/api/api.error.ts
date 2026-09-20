import { HttpErrorResponse } from '@angular/common/http';
import type { ApiErrorBody } from './types';

/**
 * A failed request, normalised. The UI needs three things from an error:
 * a stable code to branch on, a translation key, and something to show
 * when no translation exists.
 */
export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    override readonly message: string,
    readonly field?: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }

  /** Key into the `errors` section of the translation files. */
  get translationKey(): string {
    return `errors.${this.code}`;
  }

  get isUnauthorized(): boolean {
    return this.status === 401;
  }

  get isForbidden(): boolean {
    return this.status === 403;
  }

  /** True when retrying the same request could plausibly work. */
  get isTransient(): boolean {
    return this.status === 0 || this.status === 503 || this.status === 504;
  }

  static from(err: unknown): ApiError {
    if (err instanceof ApiError) {
      return err;
    }
    if (err instanceof HttpErrorResponse) {
      // status 0 means the request never reached the server: offline,
      // DNS, a dropped tunnel. Saying "unknown error" there sends people
      // looking in the wrong place.
      if (err.status === 0) {
        return new ApiError(0, 'network_unreachable', 'Could not reach the server.');
      }
      const body = err.error as Partial<ApiErrorBody> | string | null;
      if (body && typeof body === 'object' && body.error?.code) {
        return new ApiError(err.status, body.error.code, body.error.message, body.error.field);
      }
      return new ApiError(err.status, 'internal_error', err.statusText || 'Request failed.');
    }
    return new ApiError(0, 'internal_error', err instanceof Error ? err.message : String(err));
  }
}
