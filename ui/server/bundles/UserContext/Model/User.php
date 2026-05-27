<?php

declare(strict_types=1);

namespace Bundles\UserContext\Model;

use Bundles\UserContext\Entity\Group;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Stringable;
use Symfony\Component\Security\Core\User\PasswordAuthenticatedUserInterface;
use Symfony\Component\Security\Core\User\UserInterface;

abstract class User implements UserInterface, PasswordAuthenticatedUserInterface, Stringable
{
    public const ROLE_DEFAULT = 'ROLE_USER';

    public const ROLE_ADMIN = 'ROLE_ADMIN';

    public const ROLE_ALL_ACCESS = 'ROLE_SUPER_ADMIN';

    protected ?int $id = null;

    protected ?string $password = null;

    protected ?string $email = null;

    protected bool $active = true;

    protected bool $freeze = false;

    protected ?\DateTimeInterface $lastLogin = null;

    protected ?string $lastLoginIp = null;

    protected ?string $confirmationToken = null;

    protected ?\DateTimeInterface $passwordRequestedAt = null;

    protected ?\DateTimeInterface $createdAt = null;

    /** @var list<string> */
    protected array $roles = [];

    /** @var Collection<int, Group> */
    protected Collection $groups;

    public function __construct()
    {
        $this->roles     = [self::ROLE_DEFAULT];
        $this->createdAt = new \DateTime();
        $this->groups    = new ArrayCollection();
    }

    public function __toString(): string
    {
        return (string) $this->getUserIdentifier();
    }

    abstract public function getUsername(): string;

    public function getUserIdentifier(): string
    {
        return $this->getUsername();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getPassword(): ?string
    {
        return $this->password;
    }

    public function setPassword(string $password): static
    {
        $this->password = $password;

        return $this;
    }

    public function getEmail(): ?string
    {
        return $this->email;
    }

    public function setEmail(?string $email): static
    {
        $this->email = $email;

        return $this;
    }

    public function isActive(): bool
    {
        return $this->active;
    }

    public function setActive(bool $active): static
    {
        $this->active = $active;

        return $this;
    }

    public function isFreeze(): bool
    {
        return $this->freeze;
    }

    public function setFreeze(bool $enabled): static
    {
        $this->freeze = $enabled;

        return $this;
    }

    public function getLastLogin(): ?\DateTimeInterface
    {
        return $this->lastLogin;
    }

    public function setLastLogin(?\DateTimeInterface $time = null): static
    {
        $this->lastLogin = $time;

        return $this;
    }

    public function getLastLoginIp(): ?string
    {
        return $this->lastLoginIp;
    }

    public function setLastLoginIp(?string $lastLoginIp): static
    {
        $this->lastLoginIp = $lastLoginIp;

        return $this;
    }

    public function getConfirmationToken(): ?string
    {
        return $this->confirmationToken;
    }

    public function setConfirmationToken(?string $confirmationToken): static
    {
        $this->confirmationToken = $confirmationToken;

        return $this;
    }

    public function createConfirmationToken(): static
    {
        $this->confirmationToken = rtrim(strtr(base64_encode(random_bytes(32)), '+/', '-_'), '=');

        return $this;
    }

    public function getPasswordRequestedAt(): ?\DateTimeInterface
    {
        return $this->passwordRequestedAt;
    }

    public function setPasswordRequestedAt(?\DateTimeInterface $date = null): static
    {
        $this->passwordRequestedAt = $date;

        return $this;
    }

    public function isPasswordRequestNonExpired(int $ttl): bool
    {
        return $this->passwordRequestedAt instanceof \DateTimeInterface
            && $this->passwordRequestedAt->getTimestamp() + $ttl > time();
    }

    public function getCreatedAt(): ?\DateTimeInterface
    {
        return $this->createdAt;
    }

    public function setCreatedAt(?\DateTimeInterface $time = null): static
    {
        $this->createdAt = $time;

        return $this;
    }

    /**
     * @return list<string>
     */
    public function getRoles(): array
    {
        $groupRoles = [[]];
        foreach ($this->groups as $group) {
            $groupRoles[] = $group->getRoles();
        }

        return \array_values(\array_unique(\array_merge($this->roles, ...$groupRoles)));
    }

    /**
     * @return list<string>
     */
    public function getRolesUser(): array
    {
        return $this->roles;
    }

    /**
     * @param list<string> $roles
     */
    public function setRoles(array $roles): static
    {
        $this->roles = $roles;

        return $this;
    }

    /**
     * @param list<string> $roles
     */
    public function addRoles(array $roles): static
    {
        $this->roles = [];

        foreach ($roles as $role) {
            $this->addRole($role);
        }

        return $this;
    }

    public function addRole(string $role): static
    {
        $role = \mb_strtoupper($role);

        if (!\in_array($role, $this->roles, true)) {
            $this->roles[] = $role;
        }

        return $this;
    }

    public function removeRole(string $role): static
    {
        $key = \array_search(\mb_strtoupper($role), $this->roles, true);

        if ($key !== false) {
            unset($this->roles[$key]);
            $this->roles = \array_values($this->roles);
        }

        return $this;
    }

    public function hasRole(string $role): bool
    {
        return \in_array(\mb_strtoupper($role), $this->getRoles(), true);
    }

    /**
     * @return Collection<int, Group>
     */
    public function getGroups(): Collection
    {
        return $this->groups;
    }

    /**
     * @param Collection<int, Group> $groups
     */
    public function setGroups(Collection $groups): static
    {
        $this->groups = $groups;

        return $this;
    }

    /**
     * @return list<string>
     */
    public function getGroupNames(): array
    {
        $names = [];

        foreach ($this->groups as $group) {
            $names[] = $group->getName();
        }

        return $names;
    }

    public function hasGroup(string $name): bool
    {
        return \in_array($name, $this->getGroupNames(), true);
    }

    public function addGroup(Group $group): static
    {
        if (!$this->groups->contains($group)) {
            $this->groups->add($group);
        }

        return $this;
    }

    public function removeGroup(Group $group): static
    {
        $this->groups->removeElement($group);

        return $this;
    }

    public function eraseCredentials(): void
    {
    }
}
