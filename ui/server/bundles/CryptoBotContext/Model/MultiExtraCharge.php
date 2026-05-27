<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Model;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Doctrine\Common\Collections\ArrayCollection;
use Symfony\Component\Validator\Constraints as Assert;

class MultiExtraCharge
{
    /**
     * @Assert\NotNull
     *
     * @var null|int
     */
    public ?int $orderId = null;

    /**
     * @Assert\Valid
     *
     * @var ArrayCollection
     */
    public ArrayCollection $extraChargeOptions;

    public CryptoBot $cryptobot;

    public function __construct(CryptoBot $cryptobot)
    {
        $this->extraChargeOptions = new ArrayCollection();
        $this->cryptobot          = $cryptobot;
    }
}
