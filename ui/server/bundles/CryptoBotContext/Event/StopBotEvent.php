<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Event;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Symfony\Contracts\EventDispatcher\Event;

class StopBotEvent extends Event
{
    private CryptoBot $cryptoBot;

    public function __construct(CryptoBot $cryptoBot)
    {
        $this->cryptoBot = $cryptoBot;
    }

    public function getCryptoBot(): CryptoBot
    {
        return $this->cryptoBot;
    }
}
