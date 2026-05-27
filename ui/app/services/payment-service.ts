import {Subject} from "rxjs";

export class PaymentService {
    public paymentModal: Subject<void> = new Subject<void>();
}
