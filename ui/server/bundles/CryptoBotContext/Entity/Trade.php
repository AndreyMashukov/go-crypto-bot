<?php
namespace Bundles\CryptoBotContext\Entity;

use Bundles\UserContext\Entity\User;

class Trade
{
    private ?int $id = null;

    private ?User $user = null;

    private ?CryptoBot $bot = null;

    private ?float $profit = null;

    private ?string $symbol = null;

    private ?float $buyQty = null;

    private ?float $sellQty = null;

    private ?float $buyPrice = null;

    private ?float $sellPrice = null;

    private ?float $percent = null;

    private $buyDate;

    private $sellDate;

    private ?int $orderId = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getUser(): ?User
    {
        return $this->user;
    }

    public function setUser(?User $user): self
    {
        $this->user = $user;

        return $this;
    }

    public function getBot(): ?CryptoBot
    {
        return $this->bot;
    }

    public function setBot(?CryptoBot $bot): self
    {
        $this->bot = $bot;

        return $this;
    }

    public function getProfit(): ?float
    {
        return $this->profit;
    }

    public function setProfit(float $profit): self
    {
        $this->profit = $profit;

        return $this;
    }

    public function getSymbol(): ?string
    {
        return $this->symbol;
    }

    public function setSymbol(string $symbol): self
    {
        $this->symbol = $symbol;

        return $this;
    }

    public function getBuyQty(): ?float
    {
        return $this->buyQty;
    }

    public function setBuyQty(float $buyQty): self
    {
        $this->buyQty = $buyQty;

        return $this;
    }

    public function getSellQty(): ?float
    {
        return $this->sellQty;
    }

    public function setSellQty(float $sellQty): self
    {
        $this->sellQty = $sellQty;

        return $this;
    }

    public function getBuyPrice(): ?float
    {
        return $this->buyPrice;
    }

    public function setBuyPrice(float $buyPrice): self
    {
        $this->buyPrice = $buyPrice;

        return $this;
    }

    public function getSellPrice(): ?float
    {
        return $this->sellPrice;
    }

    public function setSellPrice(float $sellPrice): self
    {
        $this->sellPrice = $sellPrice;

        return $this;
    }

    public function getPercent(): ?float
    {
        return $this->percent;
    }

    public function setPercent(float $percent): self
    {
        $this->percent = $percent;

        return $this;
    }

    public function getBuyDate(): ?\DateTimeImmutable
    {
        return $this->buyDate;
    }

    public function setBuyDate(\DateTimeImmutable $buyDate): self
    {
        $this->buyDate = $buyDate;

        return $this;
    }

    public function getSellDate(): ?\DateTimeImmutable
    {
        return $this->sellDate;
    }

    public function setSellDate(\DateTimeImmutable $sellDate): self
    {
        $this->sellDate = $sellDate;

        return $this;
    }

    public function getOrderId(): ?int
    {
        return $this->orderId;
    }

    public function setOrderId(int $orderId): self
    {
        $this->orderId = $orderId;

        return $this;
    }

    public function getCommission(): float
    {
        return ($this->getProfit() ?: 0) * 0.5;
    }
}
