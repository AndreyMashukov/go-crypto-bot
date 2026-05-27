<?php
namespace Bundles\TgBotContext\Model;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Symfony\Component\Validator\Constraints as Assert;

class OrderMessage
{
    /**
     * @Assert\NotNull
     */
    private ?CryptoBot $bot = null;

    /**
     * @Assert\NotBlank
     */
    private ?string $symbol = null;

    /**
     * @Assert\NotBlank
     */
    private ?float $amount = null;

    /**
     * @Assert\NotBlank
     */
    private ?float $price = null;

    /**
     * @Assert\NotBlank
     */
    private ?string $operation = null;

    /**
     * @Assert\NotBlank
     */
    private ?string $dateTime = null;

    private ?string $details = null;

    public function getBot(): ?CryptoBot
    {
        return $this->bot;
    }

    public function setBot(?CryptoBot $bot): self
    {
        $this->bot = $bot;

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

    public function getAmount(): ?float
    {
        return $this->amount;
    }

    public function setAmount(?float $amount): self
    {
        $this->amount = $amount;

        return $this;
    }

    public function getPrice(): ?float
    {
        return $this->price;
    }

    public function setPrice(?float $price): self
    {
        $this->price = $price;

        return $this;
    }

    public function getOperation(): ?string
    {
        return $this->operation;
    }

    public function setOperation(?string $operation): self
    {
        $this->operation = $operation;

        return $this;
    }

    public function getDateTime(): ?string
    {
        return $this->dateTime;
    }

    public function setDateTime(?string $dateTime): self
    {
        $this->dateTime = $dateTime;

        return $this;
    }

    public function getDetails(): ?string
    {
        return $this->details;
    }

    public function setDetails(?string $details): self
    {
        $this->details = $details;

        return $this;
    }
}
