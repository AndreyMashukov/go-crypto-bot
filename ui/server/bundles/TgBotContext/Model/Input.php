<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\TgBotContext\Model;

interface Input
{
    public function getText(): string;

    public function isJson(): bool;

    public function hasText(string $text): bool;

    public function getEditMessage(): array;
}
