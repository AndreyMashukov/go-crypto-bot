import {Alert} from '~/model/alert';

export class AlertEvent {
  constructor(public alert: Alert, public timeout: number) {
  }
}
