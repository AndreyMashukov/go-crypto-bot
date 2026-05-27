<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Model;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Symfony\Component\Validator\Constraints as Assert;

class ManualOrder
{
    #[Assert\GreaterThan(value: 0)]
    public float $price = 0.00;

    #[Assert\NotBlank]
    public string $symbol = '';

    #[Assert\NotBlank]
    public string $operation = '';

    #[Assert\GreaterThanOrEqual(value: 0)]
    public int $ttl = 3600 * 24;

    public function __construct(private readonly CryptoBot $cryptoBot)
    {
    }

    public function getCryptoBot(): CryptoBot
    {
        return $this->cryptoBot;
    }
}
