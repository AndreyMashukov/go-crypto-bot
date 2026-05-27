<?php
namespace Bundles\CryptoBotContext\Entity\Embedded;

use Doctrine\ORM\Mapping as ORM;
use JMS\Serializer\Annotation as Serializer;
use Symfony\Component\Validator\Constraints as Assert;

class SignalConfig
{
    public const SELL_PRICE_CORRECTION_MODE_MAX = 'max';

    public const SELL_PRICE_CORRECTION_MODE_MIN = 'min';

    public const SELL_PRICE_CORRECTION_MODE_EQUAL = 'equal';

    /**
     *
     *
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    #[ORM\Column(name: 'percent_filter', type: 'float', options: ['default' => '1.00'])]
    #[Assert\NotNull]
    #[Assert\GreaterThanOrEqual(value: 0.5)]
    private float $percentFilter = 1.00;

    /**
     *
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    #[ORM\Column(name: 'rating_filter', type: 'boolean', options: ['default' => 0])]
    private bool $ratingFilter = false;

    /**
     *
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    #[ORM\Column(name: 'avg_buy_filter', type: 'boolean', options: ['default' => 0])]
    private bool $avgBuyFilter = false;

    /**
     *
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    #[ORM\Column(name: 'avg_sell_filter', type: 'boolean', options: ['default' => 0])]
    private bool $avgSellFilter = false;

    /**
     *
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    #[ORM\Column(name: 'avg_buy_correction', type: 'boolean', options: ['default' => 1])]
    private bool $avgBuyCorrection = true;

    /**
     *
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    #[ORM\Column(name: 'avg_sell_correction', type: 'boolean', options: ['default' => 0])]
    private bool $avgSellCorrection = false;

    /**
     *
     *
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    #[ORM\Column(name: 'sell_price_correction_mode', type: 'string', length: 10, nullable: false, options: ['default' => 'equal'])]
    #[Assert\NotNull]
    #[Assert\Choice(choices: ['max', 'min', 'equal'])]
    private string $sellPriceCorrectionMode = self::SELL_PRICE_CORRECTION_MODE_EQUAL;

    /**
     *
     *
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    #[ORM\Column(name: 'signal_period_days', type: 'integer', options: ['default' => 1])]
    #[Assert\NotNull]
    #[Assert\Choice(choices: [1, 7, 14, 30])]
    private int $signalPeriodDays = 1;

    public function getPercentFilter(): float
    {
        return $this->percentFilter;
    }

    public function setPercentFilter(float $percentFilter): self
    {
        $this->percentFilter = $percentFilter;

        return $this;
    }

    public function isAvgBuyFilter(): bool
    {
        return $this->avgBuyFilter;
    }

    public function setAvgBuyFilter(bool $avgBuyFilter): self
    {
        $this->avgBuyFilter = $avgBuyFilter;

        return $this;
    }

    public function isAvgSellFilter(): bool
    {
        return $this->avgSellFilter;
    }

    public function setAvgSellFilter(bool $avgSellFilter): self
    {
        $this->avgSellFilter = $avgSellFilter;

        return $this;
    }

    public function isAvgBuyCorrection(): bool
    {
        return $this->avgBuyCorrection;
    }

    public function setAvgBuyCorrection(bool $avgBuyCorrection): self
    {
        $this->avgBuyCorrection = $avgBuyCorrection;

        return $this;
    }

    public function isAvgSellCorrection(): bool
    {
        return $this->avgSellCorrection;
    }

    public function setAvgSellCorrection(bool $avgSellCorrection): self
    {
        $this->avgSellCorrection = $avgSellCorrection;

        return $this;
    }

    public function isRatingFilter(): bool
    {
        return $this->ratingFilter;
    }

    public function setRatingFilter(bool $ratingFilter): self
    {
        $this->ratingFilter = $ratingFilter;

        return $this;
    }

    public function getSellPriceCorrectionMode(): string
    {
        return $this->sellPriceCorrectionMode;
    }

    public function setSellPriceCorrectionMode(string $sellPriceCorrectionMode): self
    {
        $this->sellPriceCorrectionMode = $sellPriceCorrectionMode;

        return $this;
    }

    public function getSignalPeriodDays(): int
    {
        return $this->signalPeriodDays;
    }

    public function setSignalPeriodDays(int $signalPeriodDays): self
    {
        $this->signalPeriodDays = $signalPeriodDays;

        return $this;
    }
}
