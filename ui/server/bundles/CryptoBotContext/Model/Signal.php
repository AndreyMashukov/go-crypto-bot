<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Model;

use JMS\Serializer\Annotation as Serializer;
use Symfony\Component\Validator\Constraints as Assert;

/**
 * @Serializer\ExclusionPolicy(Serializer\ExclusionPolicy::ALL)
 */
class Signal
{
    /**
     * @Serializer\Expose
     * @Assert\NotNull
     *
     * @var null|string
     */
    private ?string $exchange = null;

    /**
     * @Serializer\Expose
     * @Assert\NotNull
     *
     * @var null|string
     */
    private ?string $symbol = null;

    /**
     * @Serializer\Expose
     * @Assert\NotNull
     * @Assert\GreaterThan(value="0")
     *
     * @var null|float
     */
    private ?float $buyPrice = null;

    /**
     * @Serializer\Expose
     * @Assert\NotNull
     * @Assert\GreaterThan(value="0")
     *
     * @var null|float
     */
    private ?float $percent = null;

    /**
     * @Serializer\Expose
     * @Serializer\Type("array<Bundles\CryptoBotContext\Model\SignalProfitOption>")
     * @Assert\Valid
     * @Assert\Count(min="1")
     *
     * @var array
     */
    private array $profitOptions = [];

    /**
     * @Serializer\Expose
     * @Serializer\Type("array<Bundles\CryptoBotContext\Model\SignalExtraChargeOption>")
     * @Assert\Valid
     * @Assert\Count(min="1")
     *
     * @var array
     */
    private array $extraChargeOptions = [];

    /**
     * @Serializer\Expose
     * @Assert\NotNull
     * @Assert\GreaterThan(value="0")
     *
     * @var null|int
     */
    private ?int $expireTimestamp = null;

    /**
     * @Assert\NotNull
     * @Assert\Choice(choices={1, 7, 14, 30})
     *
     * @var null|int
     */
    private ?int $periodDays = null;

    public function getExchange(): ?string
    {
        return $this->exchange;
    }

    public function setExchange(?string $exchange): self
    {
        $this->exchange = $exchange;

        return $this;
    }

    public function getSymbol(): ?string
    {
        return $this->symbol;
    }

    public function setSymbol(?string $symbol): self
    {
        $this->symbol = $symbol;

        return $this;
    }

    public function getBuyPrice(): ?float
    {
        return $this->buyPrice;
    }

    public function setBuyPrice(?float $buyPrice): self
    {
        $this->buyPrice = $buyPrice;

        return $this;
    }

    public function getPercent(): ?float
    {
        return $this->percent;
    }

    public function setPercent(?float $percent): self
    {
        $this->percent = $percent;

        return $this;
    }

    public function getProfitOptions(): array
    {
        return $this->profitOptions;
    }

    public function setProfitOptions(array $profitOptions): self
    {
        $this->profitOptions = $profitOptions;

        return $this;
    }

    public function getExtraChargeOptions(): array
    {
        return $this->extraChargeOptions;
    }

    public function setExtraChargeOptions(array $extraChargeOptions): self
    {
        $this->extraChargeOptions = $extraChargeOptions;

        return $this;
    }

    public function getExpireTimestamp(): ?int
    {
        return $this->expireTimestamp;
    }

    public function setExpireTimestamp(?int $expireTimestamp): self
    {
        $this->expireTimestamp = $expireTimestamp;

        return $this;
    }

    public function getMinPercent(): float
    {
        $minPercent = null;

        /** @var SignalProfitOption $option */
        foreach ($this->getProfitOptions() as $option) {
            if (null === $minPercent) {
                $minPercent = $option->optionPercent;
                continue;
            }

            $minPercent = min($minPercent, $option->optionPercent);
        }

        return $minPercent ?: 0.5;
    }

    public function getMinSellPrice(): float
    {
        $minSellPrice = null;

        /** @var SignalProfitOption $option */
        foreach ($this->getProfitOptions() as $option) {
            if (null === $minSellPrice) {
                $minSellPrice = $option->getSellPrice();
                continue;
            }

            $minSellPrice = min($minSellPrice, $option->getSellPrice());
        }

        if (null === $minSellPrice) {
            throw new \BadMethodCallException('Sell price must be set!');
        }

        return $minSellPrice;
    }

    public function getMaxSellPrice(): float
    {
        $minSellPrice = null;

        /** @var SignalProfitOption $option */
        foreach ($this->getProfitOptions() as $option) {
            if (null === $minSellPrice) {
                $minSellPrice = $option->getSellPrice();
                continue;
            }

            $minSellPrice = max($minSellPrice, $option->getSellPrice());
        }

        if (null === $minSellPrice) {
            throw new \BadMethodCallException('Sell price must be set!');
        }

        return $minSellPrice;
    }

    public function getPeriodDays(): ?int
    {
        return $this->periodDays;
    }

    public function setPeriodDays(?int $periodDays): self
    {
        $this->periodDays = $periodDays;

        return $this;
    }
}
