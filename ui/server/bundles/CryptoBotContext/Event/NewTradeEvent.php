<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Event;

use Bundles\CryptoBotContext\Entity\Trade;
use Symfony\Contracts\EventDispatcher\Event;

class NewTradeEvent extends Event
{
    public function __construct(private readonly Trade $trade)
    {
    }

    public function getTrade(): Trade
    {
        return $this->trade;
    }
}
