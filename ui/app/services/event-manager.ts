import {Subject} from 'rxjs';
import type {AlertEvent} from '~/model/alert-event';
import type {Confirmation} from '~/model/confirmation';

export class EventManager {
  alertSubject: Subject<AlertEvent> = new Subject<AlertEvent>();

  confirmationSubject: Subject<Confirmation> = new Subject<Confirmation>();

  public alert(alert: AlertEvent) {
    this.alertSubject.next(alert);
  }

  public confirmation(confirmation: Confirmation) {
      this.confirmationSubject.next(confirmation);
  }
}
