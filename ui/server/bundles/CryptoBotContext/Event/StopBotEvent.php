<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Event;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Symfony\Contracts\EventDispatcher\Event;

class StopBotEvent extends Event
{
    public function __construct(private readonly CryptoBot $cryptoBot)
    {
    }

    public function getCryptoBot(): CryptoBot
    {
        return $this->cryptoBot;
    }
}
