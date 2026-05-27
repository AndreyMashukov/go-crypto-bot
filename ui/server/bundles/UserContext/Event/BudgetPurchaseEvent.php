<?php

declare(strict_types=1);

namespace Bundles\UserContext\Event;

use Bundles\UserContext\Entity\User;
use Symfony\Contracts\EventDispatcher\Event;

class BudgetPurchaseEvent extends Event
{
    public function __construct(private readonly User $user, private readonly float $amount)
    {
    }

    public function getUser(): User
    {
        return $this->user;
    }

    public function getAmount(): float
    {
        return $this->amount;
    }
}
