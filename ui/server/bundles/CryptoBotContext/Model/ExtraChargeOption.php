<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Model;

use Symfony\Component\Validator\Constraints as Assert;

class ExtraChargeOption
{
    /**
     * @Assert\NotNull
     *
     * @var null|int
     */
    public ?int $index = null;

    /**
     * @Assert\NotNull
     * @Assert\LessThan(value="0.00")
     *
     * @var null|float
     */
    public ?float $percent = null;

    /**
     * @Assert\NotNull
     * @Assert\GreaterThanOrEqual(value="15.00")
     *
     * @var null|float
     */
    public ?float $amountUsdt = null;
}
