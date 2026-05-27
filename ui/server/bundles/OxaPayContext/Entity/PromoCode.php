<?php
namespace Bundles\OxaPayContext\Entity;

use Bundles\UserContext\Entity\User;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;

class PromoCode
{
    private ?int $id = null;

    private ?int $complimentaryBudget = 0;

    private ?int $complimentarySignalsDays = 0;

    private int $maxActivationLimit = 0;

    private int $activationCount = 0;

    private float $partnerFeePercent = 0.00;

    private bool $active = true;

    private $users;

    public function __construct(private User $partner, private string $code)
    {
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
        unset($user); if (!$this->isActive()) {
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
        if ($this->users->removeElement($user) && $user->getPromoCode() === $this) {
            $user->setPromoCode(null);
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
