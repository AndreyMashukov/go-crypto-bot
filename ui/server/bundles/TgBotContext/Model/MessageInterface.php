<?php

declare(strict_types=1);

namespace Bundles\TgBotContext\Model;

use TgBotApi\BotApiBase\BotApiComplete;

interface MessageInterface
{
    public function send(BotApiComplete $bot, ConfigurationInterface $configuration): void;

    public function getKeyboard(): ?Keyboard;
}
