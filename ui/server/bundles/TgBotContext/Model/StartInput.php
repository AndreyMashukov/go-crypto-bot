<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\TgBotContext\Model;

class StartInput implements Input
{
    private ?string $tracker;

    public function __construct(?string $tracker)
    {
        $this->tracker = $tracker;
    }

    public function getText(): string
    {
        return '/start';
    }

    public function hasText(string $text): bool
    {
        return '/start' === trim($text);
    }

    public function isJson(): bool
    {
        return null !== json_decode($this->getText(), true);
    }

    public function getEditMessage(): array
    {
        return [];
    }

    public function getTracker(): ?string
    {
        return $this->tracker;
    }
}
