import {
  ApplicationConfig,
  inject,
  provideAppInitializer,
  provideBrowserGlobalErrorListeners,
  provideZonelessChangeDetection,
  isDevMode,
} from '@angular/core';
import { provideHttpClient, withFetch, withInterceptors } from '@angular/common/http';
import { provideRouter, withComponentInputBinding, withInMemoryScrolling } from '@angular/router';
import { provideTransloco } from '@jsverse/transloco';

import { routes } from './app.routes';
import { authInterceptor } from './core/auth/auth.interceptor';
import { Auth } from './core/auth/auth.service';
import { HttpTranslocoLoader } from './core/i18n/transloco-loader';
import { LanguageService, SUPPORTED_LANGUAGES } from './core/i18n/language.service';
import { ServerInfo } from './core/meta/meta.service';
import { Theme } from './core/theme/theme.service';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideZonelessChangeDetection(),

    provideRouter(
      routes,
      withComponentInputBinding(),
      // A page's scroll position belongs to that page; coming back to a
      // list should land where it was left, not at the top.
      withInMemoryScrolling({ scrollPositionRestoration: 'enabled', anchorScrolling: 'enabled' }),
    ),

    provideHttpClient(withFetch(), withInterceptors([authInterceptor])),

    provideTransloco({
      config: {
        availableLangs: [...SUPPORTED_LANGUAGES],
        defaultLang: 'en',
        fallbackLang: 'en',
        reRenderOnLangChange: true,
        prodMode: !isDevMode(),
        missingHandler: {
          // In development a missing key should be loud; in production it
          // should render the key rather than an empty gap, so the page
          // stays usable and the gap is obvious in a screenshot.
          logMissingKey: isDevMode(),
          useFallbackTranslation: true,
        },
      },
      loader: HttpTranslocoLoader,
    }),

    // Everything the first paint depends on, resolved before the router
    // runs: the theme (so there is no flash), the language, the server's
    // capabilities and the session.
    provideAppInitializer(() => {
      inject(Theme);
      inject(LanguageService);
      const server = inject(ServerInfo);
      const auth = inject(Auth);
      return Promise.all([server.load(), auth.restore()]);
    }),
  ],
};
