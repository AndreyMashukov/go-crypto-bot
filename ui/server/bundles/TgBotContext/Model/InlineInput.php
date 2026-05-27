<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\TgBotContext\Model;

class InlineInput implements Input
{
    private string $text;

    private ?int $messageId;

    private ?string $messageText;

    public function __construct(string $text, ?int $messageId = null, ?string $messageText = null)
    {
        $this->text        = $text;
        $this->messageId   = $messageId;
        $this->messageText = $messageText;
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
