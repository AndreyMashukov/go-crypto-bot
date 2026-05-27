<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\OxaPayContext\Model;

class OxaCallback
{
    private ?string $status = null;

    private ?int $trackId = null;

    private ?float $amount = null;

    private ?string $currency = null;

    private ?bool $feePaidByPayer = null;

    private ?float $underPaidCover = null;

    private ?string $email = null;

    private ?string $orderId = null;

    private ?string $description = null;

    private ?int $date = null;

    private ?int $payDate = null;

    private ?string $type = null;

    private ?float $payAmount = null;

    private ?string $payCurrency = null;

    private ?float $price = null;

    public function getStatus(): ?string
    {
        return $this->status;
    }

    public function setStatus(?string $status): self
    {
        $this->status = $status;

        return $this;
    }

    public function getTrackId(): ?int
    {
        return $this->trackId;
    }

    public function setTrackId(?int $trackId): self
    {
        $this->trackId = $trackId;

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

    public function getCurrency(): ?string
    {
        return $this->currency;
    }

    public function setCurrency(?string $currency): self
    {
        $this->currency = $currency;

        return $this;
    }

    public function getFeePaidByPayer(): ?bool
    {
        return $this->feePaidByPayer;
    }

    public function setFeePaidByPayer(?bool $feePaidByPayer): self
    {
        $this->feePaidByPayer = $feePaidByPayer;

        return $this;
    }

    public function getUnderPaidCover(): ?float
    {
        return $this->underPaidCover;
    }

    public function setUnderPaidCover(?float $underPaidCover): self
    {
        $this->underPaidCover = $underPaidCover;

        return $this;
    }

    public function getEmail(): ?string
    {
        return $this->email;
    }

    public function setEmail(?string $email): self
    {
        $this->email = $email;

        return $this;
    }

    public function getOrderId(): ?string
    {
        return $this->orderId;
    }

    public function setOrderId(?string $orderId): self
    {
        $this->orderId = $orderId;

        return $this;
    }

    public function getDescription(): ?string
    {
        return $this->description;
    }

    public function setDescription(?string $description): self
    {
        $this->description = $description;

        return $this;
    }

    public function getDate(): ?int
    {
        return $this->date;
    }

    public function setDate(?int $date): self
    {
        $this->date = $date;

        return $this;
    }

    public function getPayDate(): ?int
    {
        return $this->payDate;
    }

    public function setPayDate(?int $payDate): self
    {
        $this->payDate = $payDate;

        return $this;
    }

    public function getType(): ?string
    {
        return $this->type;
    }

    public function setType(?string $type): self
    {
        $this->type = $type;

        return $this;
    }

    public function getPayAmount(): ?float
    {
        return $this->payAmount;
    }

    public function setPayAmount(?float $payAmount): self
    {
        $this->payAmount = $payAmount;

        return $this;
    }

    public function getPayCurrency(): ?string
    {
        return $this->payCurrency;
    }

    public function setPayCurrency(?string $payCurrency): self
    {
        $this->payCurrency = $payCurrency;

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
}
