<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Model;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Symfony\Component\Validator\Constraints as Assert;

class ManualOrder
{
    /**
     * @Assert\GreaterThan(value="0.00")
     *
     * @var float
     */
    public float $price = 0.00;

    /**
     * @Assert\NotBlank
     *
     * @var string
     */
    public string $symbol = '';

    /**
     * @Assert\NotBlank
     *
     * @var string
     */
    public string $operation = '';

    /**
     * @Assert\GreaterThanOrEqual(value="0")
     *
     * @var int
     */
    public int $ttl = 3600 * 24;

    private CryptoBot $cryptoBot;

    public function __construct(CryptoBot $cryptoBot)
    {
        $this->cryptoBot = $cryptoBot;
    }

    public function getCryptoBot(): CryptoBot
    {
        return $this->cryptoBot;
    }
}
