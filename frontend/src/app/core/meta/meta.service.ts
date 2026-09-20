import { Injectable, inject, signal } from '@angular/core';
import { Api } from '../api/api.service';
import type { ServerMeta } from '../api/types';

/**
 * Facts about the deployment that the UI needs before anyone signs in:
 * whether the demo account is offered, whether commands are wired up,
 * whether live updates are available.
 *
 * A UI that offers a button the backend cannot honour is worse than one
 * that explains the gap, so these drive real decisions rather than
 * decoration.
 */
@Injectable({ providedIn: 'root' })
export class ServerInfo {
  private readonly api = inject(Api);
  private readonly _meta = signal<ServerMeta | null>(null);

  readonly meta = this._meta.asReadonly();

  async load(): Promise<void> {
    try {
      this._meta.set(await this.api.get<ServerMeta>('/meta'));
    } catch (err) {
      // Not fatal: the app still works, it just cannot tailor itself.
      console.error('could not read server metadata', err);
    }
  }

  get demoMode(): boolean {
    return this._meta()?.demo_mode ?? false;
  }

  /** The commands the demo account may submit. Empty means read-only,
   *  which is what the login page must say when it is true and must not
   *  say when it is not. */
  get demoCommands(): string[] {
    return this._meta()?.demo_commands ?? [];
  }

  get commandsEnabled(): boolean {
    return this._meta()?.commands_enabled ?? false;
  }

  get eventsEnabled(): boolean {
    return this._meta()?.events_enabled ?? false;
  }

  get version(): string {
    return this._meta()?.version ?? '';
  }

  get defaultPageSize(): number {
    return this._meta()?.default_page_size ?? 50;
  }
}
