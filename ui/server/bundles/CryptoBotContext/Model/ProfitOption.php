<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Model;

use JMS\Serializer\Annotation as Serializer;
use Symfony\Component\Validator\Constraints as Assert;

class ProfitOption
{
    /**
     * @Serializer\Expose
     * @Assert\NotNull
     */
    public ?int $index = null;

    /**
     * @Serializer\Expose
     * @Assert\NotNull
     * @Assert\GreaterThanOrEqual(value="0.50")
     */
    public ?float $optionPercent = null;

    /**
     * @Serializer\Expose
     * @Assert\NotNull
     * @Assert\Choice(choices={"i", "h", "d", "m"})
     */
    public ?string $optionUnit = null;

    /**
     * @Serializer\Expose
     */
    public ?bool $isTriggerOption = null;

    /**
     * @Serializer\Expose
     * @Assert\NotNull
     * @Assert\GreaterThan(value="0.00")
     */
    public ?float $optionValue = null;
}
