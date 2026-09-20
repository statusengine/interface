import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { firstValueFrom } from 'rxjs';
import { ApiError } from './api.error';
import type { ListResponse } from './types';

/** Query parameters a list endpoint accepts. */
export interface ListQuery {
  limit?: number;
  offset?: number;
  sort?: string;
  q?: string;
  [key: string]: string | number | boolean | undefined | null;
}

/**
 * The single door to the backend. Every request goes through here so the
 * base path, the credential mode and the error shape are decided once.
 */
@Injectable({ providedIn: 'root' })
export class Api {
  private readonly http = inject(HttpClient);
  private readonly base = '/api/v1';

  async get<T>(path: string, query?: ListQuery): Promise<T> {
    return this.request<T>(() =>
      this.http.get<T>(this.base + path, { params: toParams(query), withCredentials: true }),
    );
  }

  async list<T>(path: string, query?: ListQuery): Promise<ListResponse<T>> {
    return this.get<ListResponse<T>>(path, query);
  }

  async post<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>(() =>
      this.http.post<T>(this.base + path, body ?? {}, { withCredentials: true }),
    );
  }

  /** For endpoints that answer 204. */
  async postVoid(path: string, body?: unknown): Promise<void> {
    await this.request<unknown>(() =>
      this.http.post(this.base + path, body ?? {}, {
        withCredentials: true,
        observe: 'response',
      }),
    );
  }

  private async request<T>(run: () => import('rxjs').Observable<unknown>): Promise<T> {
    try {
      return (await firstValueFrom(run())) as T;
    } catch (err) {
      throw ApiError.from(err);
    }
  }
}

function toParams(query?: ListQuery): HttpParams {
  let params = new HttpParams();
  if (!query) {
    return params;
  }
  for (const [key, value] of Object.entries(query)) {
    // An absent filter and an empty filter are the same thing to the user,
    // and sending `q=` would make the server search for nothing.
    if (value === undefined || value === null || value === '') {
      continue;
    }
    params = params.set(key, String(value));
  }
  return params;
}
