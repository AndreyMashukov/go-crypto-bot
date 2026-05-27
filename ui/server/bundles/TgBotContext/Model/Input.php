<?php

declare(strict_types=1);

namespace Bundles\TgBotContext\Model;

interface Input
{
    public function getText(): string;

    public function isJson(): bool;

    public function hasText(string $text): bool;

    public function getEditMessage(): array;
}
