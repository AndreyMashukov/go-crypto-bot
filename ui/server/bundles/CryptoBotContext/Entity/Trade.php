<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Entity;

use Bundles\CryptoBotContext\Repository\TradeRepository;
use Bundles\UserContext\Entity\User;
use Doctrine\ORM\Mapping as ORM;

/**
 * @ORM\Entity(repositoryClass=TradeRepository::class)
 */
class Trade
{
    /**
     * @ORM\Id
     * @ORM\GeneratedValue
     * @ORM\Column(name="trd_id", type="integer")
     */
    private $id;

    /**
     * @ORM\ManyToOne(targetEntity="Bundles\UserContext\Entity\User")
     * @ORM\JoinColumn(name="trd_user", nullable=false, referencedColumnName="id")
     */
    private $user;

    /**
     * @ORM\ManyToOne(targetEntity="Bundles\CryptoBotContext\Entity\CryptoBot")
     * @ORM\JoinColumn(name="trd_bot", nullable=false, referencedColumnName="ctb_id")
     */
    private $bot;

    /**
     * @ORM\Column(name="trd_profit", type="float")
     */
    private $profit;

    /**
     * @ORM\Column(name="trd_symbol", type="string", length=10)
     */
    private $symbol;

    /**
     * @ORM\Column(name="trd_buy_qty", type="float")
     */
    private $buyQty;

    /**
     * @ORM\Column(name="trd_sell_qty", type="float")
     */
    private $sellQty;

    /**
     * @ORM\Column(name="trd_buy_price", type="float")
     */
    private $buyPrice;

    /**
     * @ORM\Column(name="trd_sell_price", type="float")
     */
    private $sellPrice;

    /**
     * @ORM\Column(name="trd_percent", type="float")
     */
    private $percent;

    /**
     * @ORM\Column(name="trd_buy_date", type="datetime_immutable")
     */
    private $buyDate;

    /**
     * @ORM\Column(name="trd_sell_date", type="datetime_immutable")
     */
    private $sellDate;

    /**
     * @ORM\Column(name="trd_order_id", type="integer")
     */
    private $orderId;

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
