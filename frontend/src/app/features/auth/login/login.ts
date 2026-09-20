import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { TranslocoDirective, TranslocoService } from '@jsverse/transloco';
import { ApiError } from '../../../core/api/api.error';
import { Auth } from '../../../core/auth/auth.service';
import { LanguageService, type Language } from '../../../core/i18n/language.service';
import { ServerInfo } from '../../../core/meta/meta.service';
import { Theme } from '../../../core/theme/theme.service';
import { Icon } from '../../../shared/ui/icon';

@Component({
  selector: 'sei-login',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormsModule, TranslocoDirective, Icon],
  templateUrl: './login.html',
})
export class Login {
  private readonly auth = inject(Auth);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly transloco = inject(TranslocoService);
  readonly server = inject(ServerInfo);
  readonly theme = inject(Theme);
  readonly language = inject(LanguageService);

  readonly username = signal('');
  readonly password = signal('');
  readonly busy = signal(false);
  readonly error = signal<string | null>(null);

  /** Why the visitor was sent here, if they were sent. Saying "your
   *  session expired" during a database outage is a lie that sends
   *  someone looking for the wrong problem. */
  readonly reason = computed(() => this.route.snapshot.queryParamMap.get('reason'));
  readonly expired = computed(() => this.reason() === 'expired');
  readonly unavailable = computed(() => this.reason() === 'unavailable');

  readonly demoAvailable = computed(() => this.server.demoMode);

  async submit(): Promise<void> {
    if (this.busy()) {
      return;
    }
    this.busy.set(true);
    this.error.set(null);
    try {
      await this.auth.login(this.username().trim(), this.password());
      await this.goOn();
    } catch (err) {
      this.error.set(this.describe(err));
    } finally {
      this.busy.set(false);
    }
  }

  async enterDemo(): Promise<void> {
    if (this.busy()) {
      return;
    }
    this.busy.set(true);
    this.error.set(null);
    try {
      await this.auth.loginDemo();
      await this.goOn();
    } catch (err) {
      this.error.set(this.describe(err));
    } finally {
      this.busy.set(false);
    }
  }

  setLanguage(lang: Language): void {
    this.language.set(lang);
  }

  private async goOn(): Promise<void> {
    const next = this.route.snapshot.queryParamMap.get('next');
    // Only ever follow a same-app path, so a crafted link cannot use the
    // login to bounce someone off-site.
    const target = next && next.startsWith('/') && !next.startsWith('//') ? next : '/dashboard';
    await this.router.navigateByUrl(target);
  }

  private describe(err: unknown): string {
    const api = ApiError.from(err);
    const translated = this.transloco.translate(api.translationKey);
    // translate() echoes the key back when there is no entry for it.
    return translated === api.translationKey ? api.message : translated;
  }
}
