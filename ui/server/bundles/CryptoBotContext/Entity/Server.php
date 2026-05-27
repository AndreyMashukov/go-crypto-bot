<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Entity;

use Bundles\CryptoBotContext\Repository\ServerRepository;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;
use JMS\Serializer\Annotation as Serializer;

/**
 * @ORM\Entity(repositoryClass=ServerRepository::class)
 *
 * @Serializer\ExclusionPolicy(Serializer\ExclusionPolicy::ALL)
 */
class Server
{
    /**
     * @ORM\Id
     * @ORM\GeneratedValue
     * @ORM\Column(name="srv_id", type="integer")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"server"})
     */
    private $id;

    /**
     * @ORM\Column(name="srv_ip", type="string", length=255)
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"server"})
     */
    private $ip;

    /**
     * @ORM\Column(name="srv_slots", type="integer")
     */
    private int $slots = 3;

    /**
     * @ORM\Column(name="srv_cpu", type="integer")
     */
    private $cpu;

    /**
     * @ORM\Column(name="srv_ram", type="integer")
     */
    private $ram;

    /**
     * @ORM\Column(name="srv_master", type="boolean")
     */
    private bool $master = false;

    /**
     * @ORM\OneToMany(targetEntity="Bundles\CryptoBotContext\Entity\CryptoBot", mappedBy="server")
     */
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
