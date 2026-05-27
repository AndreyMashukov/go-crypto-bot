<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\OxaPayContext\Model;

use JMS\Serializer\Annotation as Serializer;

/**
 * @Serializer\ExclusionPolicy(Serializer\ExclusionPolicy::ALL)
 */
class Commission
{
    /**
     * @var float
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"commission"})
     */
    private float $percent;

    /**
     * @var float
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"commission"})
     */
    private float $minValueUsd;

    /**
     * @var int
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"commission"})
     */
    private int $level;

    /**
     * @var float
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"commission"})
     */
    private float $botMonthlyVolume;

    public function __construct(
        float $percent,
        float $minValueUsd,
        int $level,
        float $botMonthlyVolume
    ) {
        $this->percent          = $percent;
        $this->minValueUsd      = $minValueUsd;
        $this->level            = $level;
        $this->botMonthlyVolume = $botMonthlyVolume;
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
