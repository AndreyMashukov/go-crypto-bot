<?php
namespace Bundles\CryptoBotContext\Model;

use JMS\Serializer\Annotation as Serializer;
use Symfony\Component\Validator\Constraints as Assert;

class SignalExtraChargeOption
{
    /**
     * @Serializer\Expose
     * @Assert\NotNull
     */
    public ?int $index = null;

    /**
     * @Serializer\Expose
     * @Assert\NotNull
     * @Assert\LessThan(value="0.00")
     */
    public ?float $percent = null;

    /**
     * @Serializer\Expose
     * @Assert\NotNull
     * @Assert\GreaterThan(value="0.00")
     */
    public ?float $buyPrice = null;

    /**
     * @Serializer\Expose
     * @Assert\NotNull
     * @Assert\GreaterThan(value="0.00")
     */
    public ?float $budgetPercentage = null;

    public function getIndex(): ?int
    {
        return $this->index;
    }

    public function setIndex(?int $index): self
    {
        $this->index = $index;

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

    public function getBuyPrice(): ?float
    {
        return $this->buyPrice;
    }

    public function setBuyPrice(?float $buyPrice): self
    {
        $this->buyPrice = $buyPrice;

        return $this;
    }

    public function getBudgetPercentage(): ?float
    {
        return $this->budgetPercentage;
    }

    public function setBudgetPercentage(?float $budgetPercentage): self
    {
        $this->budgetPercentage = $budgetPercentage;

        return $this;
    }
}
