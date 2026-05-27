<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\TgBotContext\Model;

class ChatConfiguration implements ConfigurationInterface
{
    private int $chatId;

    public function __construct(int $chatId)
    {
        $this->chatId = $chatId;
    }

    public function getChatId(): int
    {
        return $this->chatId;
    }
}
