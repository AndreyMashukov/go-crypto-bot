<?php

declare(strict_types=1);

namespace Bundles\UserContext\Event;

use Bundles\OxaPayContext\Entity\Payment;
use Symfony\Contracts\EventDispatcher\Event;

class PaymentCompletedEvent extends Event
{
    public function __construct(private readonly Payment $payment)
    {
    }

    public function getPayment(): Payment
    {
        return $this->payment;
    }
}
