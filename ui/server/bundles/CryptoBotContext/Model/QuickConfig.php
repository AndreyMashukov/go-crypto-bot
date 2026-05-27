<?php
namespace Bundles\CryptoBotContext\Model;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\ExchangeSymbol;
use Symfony\Component\Validator\Constraints as Assert;

class QuickConfig
{
    #[Assert\NotNull]
    private ?ExchangeSymbol $exchangeSymbol = null;

    private bool $restartBot = false;

    public function __construct(private readonly CryptoBot $cryptoBot)
    {
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
