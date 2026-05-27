<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Model;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Doctrine\Common\Collections\ArrayCollection;
use Symfony\Component\Validator\Constraints as Assert;

class MultiProfitOption
{
    #[Assert\NotNull]
    public ?int $orderId = null;

    #[Assert\Valid]
    public ArrayCollection $profitOptions;

    public function __construct(public CryptoBot $cryptobot)
    {
        $this->profitOptions = new ArrayCollection();
    }
}
