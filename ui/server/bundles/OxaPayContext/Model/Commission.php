<?php

declare(strict_types=1);

namespace Bundles\OxaPayContext\Model;

class Commission
{
    public function __construct(private readonly float $percent, private readonly float $minValueUsd, private readonly int $level, private readonly float $botMonthlyVolume)
    {
    }

    public function getPercent(): float
    {
        return $this->percent;
    }

    public function getMinValueUsd(): float
    {
        return $this->minValueUsd;
    }

    public function getLevel(): int
    {
        return $this->level;
    }

    public function getBotMonthlyVolume(): float
    {
        return $this->botMonthlyVolume;
    }
}
