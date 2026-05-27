<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Entity;

use Bundles\CryptoBotContext\Repository\ExchangeSymbolRepository;
use Doctrine\ORM\Mapping as ORM;
use JMS\Serializer\Annotation as Serializer;

/**
 * @ORM\Entity(repositoryClass=ExchangeSymbolRepository::class)
 *
 * @Serializer\ExclusionPolicy(Serializer\ExclusionPolicy::ALL)
 */
class ExchangeSymbol //exchange_symbol
{
    /**
     * @ORM\Id
     * @ORM\GeneratedValue
     * @ORM\Column(type="integer")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"exchange_symbol"})
     */
    private ?int $id = null;

    /**
     * @ORM\Column(type="string", length=10)
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"exchange_symbol"})
     */
    private ?string $symbol = null;

    /**
     * @ORM\Column(type="string")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"exchange_symbol"})
     */
    private ?string $exchange = null;

    /**
     * @ORM\Column(type="boolean")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"exchange_symbol"})
     */
    private bool $enabled = false;

    public function getId(): ?int
    {
        return $this->id;
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

    public function getExchange(): ?string
    {
        return $this->exchange;
    }

    public function setExchange(?string $exchange): self
    {
        $this->exchange = $exchange;

        return $this;
    }

    public function isEnabled(): ?bool
    {
        return $this->enabled;
    }

    public function setEnabled(bool $enabled): self
    {
        $this->enabled = $enabled;

        return $this;
    }
}
