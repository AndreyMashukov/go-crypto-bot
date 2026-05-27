<?php

declare(strict_types=1);

namespace Bundles\TgBotContext\Model;

interface Output
{
    public function getKeyboard(): ?Keyboard;

    public function getMessage(): string;
}
