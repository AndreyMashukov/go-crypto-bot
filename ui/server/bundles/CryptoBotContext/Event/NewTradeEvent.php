<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Event;

use Bundles\CryptoBotContext\Entity\Trade;
use Symfony\Contracts\EventDispatcher\Event;

class NewTradeEvent extends Event
{
    private Trade $trade;

    public function __construct(Trade $trade)
    {
        $this->trade = $trade;
    }

    public function getTrade(): Trade
    {
        return $this->trade;
    }
}
