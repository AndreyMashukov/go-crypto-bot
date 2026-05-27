<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Event;

use Bundles\OxaPayContext\Entity\Payment;
use Symfony\Contracts\EventDispatcher\Event;

class PaymentCompletedEvent extends Event
{
    private Payment $payment;

    public function __construct(Payment $payment)
    {
        $this->payment = $payment;
    }

    public function getPayment(): Payment
    {
        return $this->payment;
    }
}
