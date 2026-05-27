<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\TgBotContext\Model;

use TgBotApi\BotApiBase\BotApiComplete;

interface MessageInterface
{
    public function send(BotApiComplete $bot, ConfigurationInterface $configuration): void;

    public function getKeyboard(): ?Keyboard;
}
