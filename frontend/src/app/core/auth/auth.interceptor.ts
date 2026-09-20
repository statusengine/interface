import { inject } from '@angular/core';
import { HttpErrorResponse, type HttpInterceptorFn } from '@angular/common/http';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';
import { Auth } from './auth.service';

/**
 * Sends the session cookie on every API call, and turns a server-side
 * "your session is gone" into the client-side equivalent so the two cannot
 * disagree about who is signed in.
 */
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(Auth);
  const router = inject(Router);

  const withCookies = req.clone({ withCredentials: true });

  return next(withCookies).pipe(
    catchError((err: unknown) => {
      const is401 = err instanceof HttpErrorResponse && err.status === 401;
      // The login and session-probe endpoints answer 401 as their normal
      // negative result; redirecting on those would be a loop.
      const isAuthProbe = req.url.includes('/auth/login') || req.url.includes('/auth/me');

      if (is401 && !isAuthProbe) {
        auth.forget();
        void router.navigate(['/login'], {
          queryParams: { next: router.url, reason: 'expired' },
        });
      }
      return throwError(() => err);
    }),
  );
};
