<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Model;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\ExchangeSymbol;
use Symfony\Component\Validator\Constraints as Assert;

class QuickConfig
{
    /**
     * @Assert\NotNull
     *
     * @var null|ExchangeSymbol
     */
    private ?ExchangeSymbol $exchangeSymbol = null;

    private bool $restartBot = false;

    private CryptoBot $cryptoBot;

    public function __construct(CryptoBot $cryptoBot)
    {
        $this->cryptoBot = $cryptoBot;
    }

    public function getExchangeSymbol(): ?ExchangeSymbol
    {
        return $this->exchangeSymbol;
    }

    public function setExchangeSymbol(?ExchangeSymbol $exchangeSymbol): self
    {
        $this->exchangeSymbol = $exchangeSymbol;

        return $this;
    }

    public function getCryptoBot(): CryptoBot
    {
        return $this->cryptoBot;
    }

    public function isRestartBot(): bool
    {
        return $this->restartBot;
    }

    public function setRestartBot(bool $restartBot): self
    {
        $this->restartBot = $restartBot;

        return $this;
    }
}
