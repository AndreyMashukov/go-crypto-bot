<?php
namespace Bundles\CryptoBotContext\Model;

use JMS\Serializer\Annotation as Serializer;
use Symfony\Component\Validator\Constraints as Assert;

class SignalProfitOption extends ProfitOption
{
    /**
     * @Serializer\Expose
     * @Assert\NotNull
     */
    private ?float $sellPrice = null;

    public function getSellPrice(): ?float
    {
        return $this->sellPrice;
    }

    public function setSellPrice(?float $sellPrice): self
    {
        $this->sellPrice = $sellPrice;

        return $this;
    }

    public static function create(
        float $sellPrice,
        int $index,
        float $optionValue,
        string $optionUnit,
        float $optionPercent,
        bool $isTriggerOption
    ): self {
        $newProfitOption                  = (new self())->setSellPrice($sellPrice);
        $newProfitOption->index           = $index;
        $newProfitOption->optionValue     = $optionValue;
        $newProfitOption->optionUnit      = $optionUnit;
        $newProfitOption->optionPercent   = $optionPercent;
        $newProfitOption->isTriggerOption = $isTriggerOption;

        return $newProfitOption;
    }
}
