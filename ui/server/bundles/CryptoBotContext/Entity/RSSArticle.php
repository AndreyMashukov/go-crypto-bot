<?php
namespace Bundles\CryptoBotContext\Entity;

class RSSArticle
{
    private ?int $id = null;

    private ?string $url = null;

    private ?int $urlCrc32 = null;

    private ?\DateTimeImmutable $expiresAt = null;

    private ?string $label = null;

    private ?float $score = null;

    private ?bool $expired = false;

    private array $coins = [];

    private ?\DateTimeImmutable $createdAt = null;

    public function __construct()
    {
        $this->createdAt = new \DateTimeImmutable('now');
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getUrl(): ?string
    {
        return $this->url;
    }

    public function setUrl(string $url): self
    {
        $this->url = $url;

        return $this;
    }

    public function getUrlCrc32(): ?int
    {
        return $this->urlCrc32;
    }

    public function setUrlCrc32(int $urlCrc32): self
    {
        $this->urlCrc32 = $urlCrc32;

        return $this;
    }

    public function getExpiresAt(): ?\DateTimeImmutable
    {
        return $this->expiresAt;
    }

    public function setExpiresAt(\DateTimeImmutable $expiresAt): self
    {
        $this->expiresAt = $expiresAt;

        return $this;
    }

    public function getLabel(): ?string
    {
        return $this->label;
    }

    public function setLabel(string $label): self
    {
        $this->label = $label;

        return $this;
    }

    public function getScore(): ?float
    {
        return $this->score;
    }

    public function setScore(float $score): self
    {
        $this->score = $score;

        return $this;
    }

    public function getCoins(): ?array
    {
        return $this->coins;
    }

    public function setCoins(array $coins): self
    {
        $this->coins = $coins;

        return $this;
    }

    public function getCreatedAt(): ?\DateTimeImmutable
    {
        return $this->createdAt;
    }

    public function setCreatedAt(\DateTimeImmutable $createdAt): self
    {
        $this->createdAt = $createdAt;

        return $this;
    }

    public function getExpired(): ?bool
    {
        return $this->expired;
    }

    public function setExpired(?bool $expired): self
    {
        $this->expired = $expired;

        return $this;
    }
}
