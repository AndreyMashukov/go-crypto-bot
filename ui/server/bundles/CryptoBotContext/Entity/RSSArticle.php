<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Entity;

use Bundles\CryptoBotContext\Repository\RSSArticleRepository;
use Doctrine\ORM\Mapping as ORM;

/**
 * @ORM\Table(name="rss_article", indexes={
 *     @ORM\Index(columns={"rsa_url", "rsa_url_crc_32"}),
 *     @ORM\Index(columns={"rsa_expired", "rsa_expires_at"}),
 * })
 *
 * @ORM\Entity(repositoryClass=RSSArticleRepository::class)
 */
class RSSArticle
{
    /**
     * @ORM\Id
     * @ORM\GeneratedValue
     * @ORM\Column(name="rsa_id", type="integer")
     */
    private ?int $id = null;

    /**
     * @ORM\Column(name="rsa_url", type="string", length=255)
     */
    private ?string $url = null;

    /**
     * @ORM\Column(name="rsa_url_crc_32", type="bigint")
     */
    private ?int $urlCrc32 = null;

    /**
     * @ORM\Column(name="rsa_expires_at", type="datetime_immutable")
     */
    private ?\DateTimeImmutable $expiresAt = null;

    /**
     * @ORM\Column(name="rsa_label", type="string", length=20)
     */
    private ?string $label = null;

    /**
     * @ORM\Column(name="rsa_score", type="float")
     */
    private ?float $score = null;

    /**
     * @ORM\Column(name="rsa_expired", type="boolean", options={"default": 0})
     */
    private ?bool $expired = false;

    /**
     * @ORM\Column(name="rsa_coins", type="json")
     */
    private array $coins = [];

    /**
     * @ORM\Column(name="rsa_created_at", type="datetime_immutable")
     */
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
