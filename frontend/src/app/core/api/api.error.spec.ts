import { HttpErrorResponse } from '@angular/common/http';
import { describe, expect, it } from 'vitest';
import { ApiError } from './api.error';

describe('ApiError.from', () => {
  it('reads the standard error envelope', () => {
    const err = ApiError.from(
      new HttpErrorResponse({
        status: 403,
        error: { error: { code: 'forbidden', message: 'not your permission', field: 'role' } },
      }),
    );

    expect(err.status).toBe(403);
    expect(err.code).toBe('forbidden');
    expect(err.message).toBe('not your permission');
    expect(err.field).toBe('role');
    expect(err.isForbidden).toBe(true);
    expect(err.translationKey).toBe('errors.forbidden');
  });

  // A request that never reached the server is a different problem from
  // one the server refused, and saying "unknown error" sends people
  // looking in the wrong place.
  it('distinguishes an unreachable server from a server error', () => {
    const offline = ApiError.from(new HttpErrorResponse({ status: 0 }));

    expect(offline.code).toBe('network_unreachable');
    expect(offline.isTransient).toBe(true);
  });

  it('falls back when the body is not our envelope', () => {
    const err = ApiError.from(
      new HttpErrorResponse({
        status: 502,
        statusText: 'Bad Gateway',
        error: '<html>nginx</html>',
      }),
    );

    expect(err.code).toBe('internal_error');
    expect(err.status).toBe(502);
  });

  it('passes an ApiError through unchanged', () => {
    const original = new ApiError(404, 'not_found', 'gone');
    expect(ApiError.from(original)).toBe(original);
  });

  it('wraps anything else', () => {
    const err = ApiError.from(new TypeError('boom'));
    expect(err.code).toBe('internal_error');
    expect(err.message).toBe('boom');
  });

  it('marks 503 as worth retrying but 400 as not', () => {
    expect(ApiError.from(new HttpErrorResponse({ status: 503 })).isTransient).toBe(true);
    expect(ApiError.from(new HttpErrorResponse({ status: 400 })).isTransient).toBe(false);
  });
});
