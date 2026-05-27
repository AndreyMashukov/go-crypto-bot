<?php

declare(strict_types=1);

namespace Bundles\TgBotContext\Model;

class ContactInput implements Input
{
    public function __construct(private readonly string $phone, private readonly string $eid)
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

    public function getPhone(): string
    {
        return $this->phone;
    }

    public function getEid(): int
    {
        return (int) $this->eid;
    }
}
