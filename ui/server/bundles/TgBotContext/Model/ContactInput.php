<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\TgBotContext\Model;

class ContactInput implements Input
{
    private string $phone;

    private string $eid;

    public function __construct(string $phone, string $eid)
    {
        $this->phone = $phone;
        $this->eid   = $eid;
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
