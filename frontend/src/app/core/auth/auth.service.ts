import { Injectable, computed, inject, signal } from '@angular/core';
import { Api } from '../api/api.service';
import { ApiError } from '../api/api.error';
import type { Identity, Permission } from '../api/types';

/**
 * Who is signed in, and what they may do.
 *
 * The permission list is advisory here: it decides what the UI offers, and
 * the server decides what actually happens. Both checks exist on purpose -
 * hiding a button an operator may not press is a courtesy, refusing the
 * request is the control.
 */
@Injectable({ providedIn: 'root' })
export class Auth {
  private readonly api = inject(Api);

  private readonly _identity = signal<Identity | null>(null);
  private readonly _resolved = signal(false);

  /** The signed-in user, or null. */
  readonly identity = this._identity.asReadonly();

  /** False until the first /auth/me has answered, so a guard can wait
   *  instead of bouncing a logged-in user to the login page on reload. */
  readonly resolved = this._resolved.asReadonly();

  readonly isAuthenticated = computed(() => this._identity() !== null);
  readonly isDemo = computed(() => this._identity()?.is_demo ?? false);
  readonly isReadOnly = computed(() => this._identity()?.read_only ?? true);
  readonly displayName = computed(
    () => this._identity()?.display_name || this._identity()?.username || '',
  );

  /** Resolve the session from the cookie. Safe to call more than once. */
  async restore(): Promise<void> {
    try {
      this._identity.set(await this.api.get<Identity>('/auth/me'));
    } catch (err) {
      // 401 is the normal answer for "nobody is signed in"; anything else
      // is worth knowing about but still leaves us signed out.
      if (!(err instanceof ApiError) || !err.isUnauthorized) {
        console.error('could not restore the session', err);
      }
      this._identity.set(null);
    } finally {
      this._resolved.set(true);
    }
  }

  async login(username: string, password: string): Promise<void> {
    this._identity.set(await this.api.post<Identity>('/auth/login', { username, password }));
    this._resolved.set(true);
  }

  async loginDemo(): Promise<void> {
    this._identity.set(await this.api.post<Identity>('/auth/login/demo'));
    this._resolved.set(true);
  }

  async logout(): Promise<void> {
    try {
      await this.api.postVoid('/auth/logout');
    } finally {
      this._identity.set(null);
    }
  }

  /** Called by the interceptor when the server says the session is gone. */
  forget(): void {
    this._identity.set(null);
    this._resolved.set(true);
  }

  can(permission: Permission): boolean {
    return this._identity()?.permissions.includes(permission) ?? false;
  }

  /** A signal-shaped `can`, for use directly in templates. */
  allows(permission: Permission) {
    return computed(() => this.can(permission));
  }
}
