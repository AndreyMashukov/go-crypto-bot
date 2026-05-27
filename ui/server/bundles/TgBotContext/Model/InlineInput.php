<?php

declare(strict_types=1);

namespace Bundles\TgBotContext\Model;

class InlineInput implements Input
{
    public function __construct(private readonly string $text, private readonly ?int $messageId = null, private readonly ?string $messageText = null)
    {
    }

    public function getText(): string
    {
        return trim($this->text);
    }

    public function hasText(string $text): bool
    {
        return $this->text === trim($text);
    }

    public function isJson(): bool
    {
        return null !== json_decode($this->getText(), true);
    }

    public function getEditMessage(): array
    {
        return [$this->messageId, $this->messageText];
    }
}
