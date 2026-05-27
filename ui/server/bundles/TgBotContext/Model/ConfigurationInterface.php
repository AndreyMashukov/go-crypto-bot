<?php

declare(strict_types=1);

namespace Bundles\TgBotContext\Model;

interface ConfigurationInterface
{
    public function getChatId(): int;
}
