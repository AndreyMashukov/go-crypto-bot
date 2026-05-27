<?php

declare(strict_types=1);

namespace Bundles\TgBotContext\Model;

class StartInput implements Input
{
    public function __construct(private readonly ?string $tracker)
    {
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
