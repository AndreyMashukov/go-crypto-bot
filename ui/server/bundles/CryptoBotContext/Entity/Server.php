<?php
namespace Bundles\CryptoBotContext\Entity;

use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;

class Server
{
    private ?int $id = null;

    private ?string $ip = null;

    private int $slots = 3;

    private ?int $cpu = null;

    private ?int $ram = null;

    private bool $master = false;

    private Collection $cryptoBots;

    public function __construct()
    {
        $this->cryptoBots = new ArrayCollection();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getIp(): ?string
    {
        return $this->ip;
    }

    public function setIp(string $ip): self
    {
        $this->ip = $ip;

        return $this;
    }

    public function getSlots(): ?int
    {
        return $this->slots;
    }

    public function setSlots(int $slots): self
    {
        $this->slots = $slots;

        return $this;
    }

    public function getCpu(): ?int
    {
        return $this->cpu;
    }

    public function setCpu(int $cpu): self
    {
        $this->cpu = $cpu;

        return $this;
    }

    public function getRam(): ?int
    {
        return $this->ram;
    }

    public function setRam(int $ram): self
    {
        $this->ram = $ram;

        return $this;
    }

    /**
     * @return Collection<int, CryptoBot>
     */
    public function getCryptoBots(): Collection
    {
        return $this->cryptoBots;
    }

    public function isMaster(): bool
    {
        return $this->master;
    }

    public function setMaster(bool $master): self
    {
        $this->master = $master;

        return $this;
    }
}
