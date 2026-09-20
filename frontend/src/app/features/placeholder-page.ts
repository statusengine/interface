import { ChangeDetectionStrategy, Component, input } from '@angular/core';
import { TranslocoDirective } from '@jsverse/transloco';
import { NotBuiltYet } from '../shared/ui/not-built-yet';
import { PageHeader } from '../shared/ui/page-header';

/**
 * One routed component for every page that exists in the navigation but
 * is scheduled for a later phase. The route supplies its title key and
 * phase through `data`, so adding a page to the rail costs one line.
 */
@Component({
  selector: 'sei-placeholder-page',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [PageHeader, NotBuiltYet, TranslocoDirective],
  template: `
    <ng-container *transloco="let t">
      <sei-page-header
        [title]="t('pages.' + titleKey() + '.title')"
        [description]="t('pages.' + titleKey() + '.description')"
      />
      <sei-not-built-yet [phase]="phase()" />
    </ng-container>
  `,
})
export class PlaceholderPage {
  readonly titleKey = input.required<string>();
  readonly phase = input.required<number>();
}
