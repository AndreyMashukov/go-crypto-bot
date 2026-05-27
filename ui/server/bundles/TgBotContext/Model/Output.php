<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\TgBotContext\Model;

interface Output
{
    public function getKeyboard(): ?Keyboard;

    public function getMessage(): string;
}
