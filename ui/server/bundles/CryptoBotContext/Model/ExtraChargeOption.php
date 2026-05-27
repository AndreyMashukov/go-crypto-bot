<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Model;

use Symfony\Component\Validator\Constraints as Assert;

class ExtraChargeOption
{
    /**
     * @Assert\NotNull
     */
    public ?int $index = null;

    /**
     * @Assert\NotNull
     * @Assert\LessThan(value="0.00")
     */
    public ?float $percent = null;

    /**
     * @Assert\NotNull
     * @Assert\GreaterThanOrEqual(value="15.00")
     */
    public ?float $amountUsdt = null;
}
