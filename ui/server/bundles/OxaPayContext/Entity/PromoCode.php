<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\OxaPayContext\Entity;

use Bundles\OxaPayContext\Repository\PromoCodeRepository;
use Bundles\UserContext\Entity\User;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;
use JMS\Serializer\Annotation as Serializer;

/**
 * @ORM\Entity(repositoryClass=PromoCodeRepository::class)
 *
 * @Serializer\ExclusionPolicy(Serializer\ExclusionPolicy::ALL)
 */
class PromoCode
{
    /**
     * @ORM\Id
     * @ORM\GeneratedValue
     * @ORM\Column(name="pcd_id", type="integer")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"promocode"})
     */
    private ?int $id = null;

    /**
     * @ORM\ManyToOne(targetEntity=User::class)
     * @ORM\JoinColumn(name="pcd_partner", nullable=false)
     */
    private User $partner;

    /**
     * @ORM\Column(name="pcd_complimentary_budget", type="float")
     */
    private ?int $complimentaryBudget = 0;

    /**
     * @ORM\Column(name="pcd_complimentary_signals_days", type="smallint")
     */
    private ?int $complimentarySignalsDays = 0;

    /**
     * @ORM\Column(name="pcd_code", type="string", length=20)
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"promocode"})
     */
    private string $code;

    /**
     * @ORM\Column(name="pcd_max_activation_limit", type="integer")
     */
    private int $maxActivationLimit = 0;

    /**
     * @ORM\Column(name="pcd_activation_count", type="integer")
     */
    private int $activationCount = 0;

    /**
     * @ORM\Column(name="pcd_partner_fee_percent", type="float")
     */
    private float $partnerFeePercent = 0.00;

    /**
     * @ORM\Column(name="pcd_active", type="boolean")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"promocode"})
     */
    private bool $active = true;

    /**
     * @ORM\OneToMany(targetEntity=User::class, mappedBy="promoCode")
     */
    private $users;

    public function __construct(User $partner, string $code)
    {
        $this->partner = $partner;
        $this->code    = $code;
        $this->users   = new ArrayCollection();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getPartner(): User
    {
        return $this->partner;
    }

    public function getComplimentaryBudget(): ?float
    {
        return $this->complimentaryBudget;
    }

    public function setComplimentaryBudget(float $complimentaryBudget): self
    {
        $this->complimentaryBudget = $complimentaryBudget;

        return $this;
    }

    public function getComplimentarySignalsDays(): ?int
    {
        return $this->complimentarySignalsDays;
    }

    public function setComplimentarySignalsDays(int $complimentarySignalsDays): self
    {
        $this->complimentarySignalsDays = $complimentarySignalsDays;

        return $this;
    }

    public function getCode(): ?string
    {
        return $this->code;
    }

    public function setCode(string $code): self
    {
        $this->code = $code;

        return $this;
    }

    public function getMaxActivationLimit(): ?int
    {
        return $this->maxActivationLimit;
    }

    public function setMaxActivationLimit(int $maxActivationLimit): self
    {
        $this->maxActivationLimit = $maxActivationLimit;

        return $this;
    }

    public function getActivationCount(): ?int
    {
        return $this->activationCount;
    }

    public function setActivationCount(int $activationCount): self
    {
        $this->activationCount = $activationCount;

        return $this;
    }

    public function isActive(): ?bool
    {
        return $this->active;
    }

    public function setActive(bool $active): self
    {
        $this->active = $active;

        return $this;
    }

    public function canUse(?User $user): bool
    {
        unset($user); // unused

        if (!$this->isActive()) {
            return false;
        }

        if (0 === $this->maxActivationLimit) {
            return true;
        }

        return $this->activationCount < $this->maxActivationLimit;
    }

    /**
     * @return Collection<int, User>
     */
    public function getUsers(): Collection
    {
        return $this->users;
    }

    public function addUser(User $user): self
    {
        if (!$this->users->contains($user)) {
            $this->users[] = $user;
            $user->setPromoCode($this);
        }

        return $this;
    }

    public function removeUser(User $user): self
    {
        if ($this->users->removeElement($user)) {
            // set the owning side to null (unless already changed)
            if ($user->getPromoCode() === $this) {
                $user->setPromoCode(null);
            }
        }

        return $this;
    }

    public function getPartnerFeePercent(): float
    {
        return $this->partnerFeePercent;
    }

    public function setPartnerFeePercent(float $partnerFeePercent): self
    {
        $this->partnerFeePercent = $partnerFeePercent;

        return $this;
    }
}
