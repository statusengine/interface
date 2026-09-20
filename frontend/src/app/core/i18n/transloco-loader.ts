import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import type { Translation, TranslocoLoader } from '@jsverse/transloco';

/** Loads a language file from the bundle. Nothing is fetched from a CDN:
 *  this app has to work on a network with no route out. */
@Injectable({ providedIn: 'root' })
export class HttpTranslocoLoader implements TranslocoLoader {
  private readonly http = inject(HttpClient);

  /**
   * Asks the browser to revalidate rather than serve from cache.
   *
   * These files keep their names across builds, and an earlier version
   * of the server sent them with `immutable`. A browser that loaded the
   * app back then will not ask again for a year on its own, and every
   * page added since renders its bare translation keys. The request
   * header is what gets those browsers back in step; the server-side
   * fix only helps the ones that have not cached it yet.
   */
  getTranslation(lang: string) {
    return this.http.get<Translation>(`/i18n/${lang}.json`, {
      headers: new HttpHeaders({ 'Cache-Control': 'no-cache' }),
    });
  }
}
