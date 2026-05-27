<?php

declare(strict_types=1);

namespace Bundles\TgBotContext\Model;

interface CityInterface
{
    public function getCityName(): string;

    public function getName(): string;
}
