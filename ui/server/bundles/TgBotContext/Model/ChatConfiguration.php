<?php

declare(strict_types=1);

namespace Bundles\TgBotContext\Model;

class ChatConfiguration implements ConfigurationInterface
{
    public function __construct(private readonly int $chatId)
    {
    }

    public function getChatId(): int
    {
        return $this->chatId;
    }
}
