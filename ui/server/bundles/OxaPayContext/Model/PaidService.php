<?php

declare(strict_types=1);

namespace Bundles\OxaPayContext\Model;

class PaidService
{
    public function __construct(private readonly string $code, private readonly bool $isActive, private readonly float $price, private readonly int $days, private readonly ?\DateTimeImmutable $expiresAt)
    {
    }

    public function getCode(): string
    {
        return $this->code;
    }

    public function isActive(): bool
    {
        return $this->isActive;
    }

    public function getPrice(): float
    {
        return $this->price;
    }

    public function getDays(): int
    {
        return $this->days;
    }

    public function getExpiresAt(): ?\DateTimeImmutable
    {
        return $this->expiresAt;
    }
}
