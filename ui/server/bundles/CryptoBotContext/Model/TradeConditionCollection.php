<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Model;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Doctrine\Common\Collections\ArrayCollection;
use Symfony\Component\Validator\Constraints as Assert;

class TradeConditionCollection
{
    #[Assert\NotNull]
    public ?string $symbol = null;

    #[Assert\Valid]
    public ArrayCollection $conditions;

    public function __construct(public CryptoBot $cryptobot)
    {
        $this->conditions = new ArrayCollection();
    }
}
