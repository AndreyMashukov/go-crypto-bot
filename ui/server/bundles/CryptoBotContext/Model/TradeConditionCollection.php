<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Model;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Doctrine\Common\Collections\ArrayCollection;
use Symfony\Component\Validator\Constraints as Assert;

class TradeConditionCollection
{
    /**
     * @Assert\NotNull
     *
     * @var null|string
     */
    public ?string $symbol = null;

    /**
     * @Assert\Valid
     *
     * @var ArrayCollection
     */
    public ArrayCollection $conditions;

    public CryptoBot $cryptobot;

    public function __construct(CryptoBot $cryptobot)
    {
        $this->conditions = new ArrayCollection();
        $this->cryptobot  = $cryptobot;
    }
}
