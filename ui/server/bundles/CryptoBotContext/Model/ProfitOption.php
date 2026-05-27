<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Model;

use JMS\Serializer\Annotation as Serializer;
use Symfony\Component\Validator\Constraints as Assert;

class ProfitOption
{
    /**
     * @Serializer\Expose
     * @Assert\NotNull
     *
     * @var null|int
     */
    public ?int $index = null;

    /**
     * @Serializer\Expose
     * @Assert\NotNull
     * @Assert\GreaterThanOrEqual(value="0.50")
     *
     * @var null|float
     */
    public ?float $optionPercent = null;

    /**
     * @Serializer\Expose
     * @Assert\NotNull
     * @Assert\Choice(choices={"i", "h", "d", "m"})
     *
     * @var null|string
     */
    public ?string $optionUnit = null;

    /**
     * @Serializer\Expose
     *
     * @var bool
     */
    public ?bool $isTriggerOption = null;

    /**
     * @Serializer\Expose
     * @Assert\NotNull
     * @Assert\GreaterThan(value="0.00")
     *
     * @var null|float
     */
    public ?float $optionValue = null;
}
