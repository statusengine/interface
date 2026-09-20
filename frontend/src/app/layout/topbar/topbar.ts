import {
  ChangeDetectionStrategy,
  Component,
  computed,
  inject,
  output,
  signal,
} from '@angular/core';
import { Router } from '@angular/router';
import { TranslocoDirective } from '@jsverse/transloco';
import { Auth } from '../../core/auth/auth.service';
import { LanguageService, type Language } from '../../core/i18n/language.service';
import { Theme } from '../../core/theme/theme.service';
import { Icon } from '../../shared/ui/icon';

@Component({
  selector: 'sei-topbar',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [TranslocoDirective, Icon],
  templateUrl: './topbar.html',
})
export class Topbar {
  private readonly auth = inject(Auth);
  private readonly router = inject(Router);
  readonly theme = inject(Theme);
  readonly language = inject(LanguageService);

  readonly openDrawer = output<void>();

  readonly menuOpen = signal(false);
  readonly displayName = this.auth.displayName;
  readonly isDemo = this.auth.isDemo;
  readonly role = computed(() => this.auth.identity()?.role ?? '');

  readonly themeIcon = computed(() => {
    switch (this.theme.preference()) {
      case 'light':
        return 'sun' as const;
      case 'dark':
        return 'moon' as const;
      default:
        return 'monitor' as const;
    }
  });

  toggleMenu(): void {
    this.menuOpen.update((open) => !open);
  }

  closeMenu(): void {
    this.menuOpen.set(false);
  }

  setLanguage(lang: Language): void {
    this.language.set(lang);
  }

  async signOut(): Promise<void> {
    this.closeMenu();
    await this.auth.logout();
    await this.router.navigate(['/login']);
  }
}
