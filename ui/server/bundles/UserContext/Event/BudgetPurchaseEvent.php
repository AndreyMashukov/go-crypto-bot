<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Event;

use Bundles\UserContext\Entity\User;
use Symfony\Contracts\EventDispatcher\Event;

class BudgetPurchaseEvent extends Event
{
    private User $user;

    private float $amount;

    public function __construct(User $user, float $amount)
    {
        $this->user   = $user;
        $this->amount = $amount;
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
