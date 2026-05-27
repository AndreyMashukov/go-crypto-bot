<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Model;

use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\ORM\Mapping as ORM;
use Pd\UserBundle\Model\GroupInterface;
use Pd\UserBundle\Model\ProfileInterface;
use Pd\UserBundle\Model\UserInterface;
use Symfony\Component\Validator\Constraints as Assert;

/**
 * User Account.
 *
 * @author Ramazan APAYDIN <apaydin541@gmail.com>
 */
abstract class User implements UserInterface, \Serializable
{
    public const ROLE_DEFAULT = 'ROLE_USER';

    public const ROLE_ADMIN = 'ROLE_ADMIN';

    public const ROLE_ALL_ACCESS = 'ROLE_SUPER_ADMIN';

    /**
     * @ORM\Column(type="integer")
     * @ORM\Id
     * @ORM\GeneratedValue(strategy="AUTO")
     */
    protected $id;

    /**
     * @ORM\OneToOne(targetEntity="Profile", cascade={"persist", "merge", "remove"})
     * @ORM\JoinColumn(name="profile_id", referencedColumnName="id")
     * @Assert\Valid
     */
    protected $profile;

    /**
     * @ORM\Column(type="string", length=100)
     */
    protected $password;

    /**
     * @ORM\Column(type="string", length=100, unique=true)
     */
    protected $email;

    /**
     * @ORM\Column(type="boolean")
     */
    protected $active;

    /**
     * @ORM\Column(type="boolean")
     */
    protected $freeze;

    /**
     * @ORM\Column(type="datetime", nullable=true)
     */
    protected $lastLogin;

    /**
     * @ORM\Column(type="string", length=32, nullable=true)
     */
    protected $lastLoginIp;

    /**
     * @ORM\Column(type="string", length=180, unique=true, nullable=true)
     */
    protected $confirmationToken;

    /**
     * @ORM\Column(type="datetime", nullable=true)
     */
    protected $passwordRequestedAt;

    /**
     * @ORM\Column(type="datetime", nullable=true)
     */
    protected $createdAt;

    /**
     * @ORM\Column(type="array")
     */
    protected $roles;

    /**
     * @ORM\ManyToMany(targetEntity="Group")
     * @ORM\JoinTable(name="user_group_tax",
     *     joinColumns={@ORM\JoinColumn(name="user_id", referencedColumnName="id", onDelete="CASCADE")},
     *     inverseJoinColumns={@ORM\JoinColumn(name="group_id", referencedColumnName="id", onDelete="CASCADE")}
     * )
     */
    protected $groups;

    public function __construct()
    {
        $this->active    = true;
        $this->freeze    = false;
        $this->roles     = [static::ROLE_DEFAULT];
        $this->createdAt = new \DateTime();
        $this->groups    = new ArrayCollection();
    }

    public function getSalt()
    {
        return null;
    }

    public function __toString()
    {
        return $this->getUsername();
    }

    public function getId(): int
    {
        return $this->id;
    }

    public function getProfile(): ?ProfileInterface
    {
        return $this->profile;
    }

    public function setProfile(?ProfileInterface $profile): UserInterface
    {
        $this->profile = $profile;

        return $this;
    }

    public function getPassword(): string
    {
        return $this->password;
    }

    public function setPassword(string $password): UserInterface
    {
        $this->password = $password;

        return $this;
    }

    public function getEmail(): ?string
    {
        return $this->email;
    }

    public function setEmail(?string $email): UserInterface
    {
        $this->email = $email;

        return $this;
    }

    public function isActive(): bool
    {
        return $this->active;
    }

    public function setActive(bool $active): UserInterface
    {
        $this->active = $active;

        return $this;
    }

    public function isFreeze(): bool
    {
        return $this->freeze;
    }

    public function setFreeze(bool $enabled): UserInterface
    {
        $this->freeze = $enabled;

        return $this;
    }

    public function getLastLogin(): ?\DateTime
    {
        return $this->lastLogin;
    }

    public function setLastLogin(\DateTime $time = null): UserInterface
    {
        $this->lastLogin = $time;

        return $this;
    }

    public function getLastLoginIp(): ?string
    {
        return $this->lastLoginIp;
    }

    public function setLastLoginIp(?string $lastLoginIp): UserInterface
    {
        $this->lastLoginIp = $lastLoginIp;

        return $this;
    }

    public function getConfirmationToken(): ?string
    {
        return $this->confirmationToken;
    }

    public function setConfirmationToken(?string $confirmationToken): UserInterface
    {
        $this->confirmationToken = $confirmationToken;

        return $this;
    }

    public function createConfirmationToken(): UserInterface
    {
        $this->confirmationToken = rtrim(strtr(base64_encode(random_bytes(32)), '+/', '-_'), '=');

        return $this;
    }

    public function getPasswordRequestedAt(): ?\DateTime
    {
        return $this->passwordRequestedAt;
    }

    public function setPasswordRequestedAt(\DateTime $date = null): UserInterface
    {
        $this->passwordRequestedAt = $date;

        return $this;
    }

    public function isPasswordRequestNonExpired($ttl): bool
    {
        return $this->getPasswordRequestedAt() instanceof \DateTime && $this->getPasswordRequestedAt()->getTimestamp() + $ttl > time();
    }

    public function getCreatedAt(): ?\DateTime
    {
        return $this->createdAt;
    }

    public function setCreatedAt(\DateTime $time = null): UserInterface
    {
        $this->createdAt = $time;

        return $this;
    }

    public function getRoles(): ?array
    {
        $roles      = $this->roles;
        $groupRoles = [[]];

        foreach ($this->getGroups() as $group) {
            $groupRoles[] = $group->getRoles();
        }
        $groupRoles = array_merge(...$groupRoles);

        return array_unique(array_merge($roles, $groupRoles));
    }

    public function getRolesUser(): ?array
    {
        return $this->roles;
    }

    public function setRoles(array $roles): UserInterface
    {
        $this->roles = $roles;

        return $this;
    }

    public function addRoles(array $roles): UserInterface
    {
        $this->roles = [];

        foreach ($roles as $role) {
            $this->addRole($role);
        }

        return $this;
    }

    public function addRole(string $role): UserInterface
    {
        $role = mb_strtoupper($role);

        if (!\in_array($role, $this->roles, true)) {
            $this->roles[] = $role;
        }

        return $this;
    }

    public function removeRole(string $role): UserInterface
    {
        if (false !== $key = array_search(mb_strtoupper($role), $this->roles, true)) {
            unset($this->roles[$key]);
            $this->roles = array_values($this->roles);
        }

        return $this;
    }

    public function hasRole(string $role): bool
    {
        return \in_array(mb_strtoupper($role), $this->getRoles(), true);
    }

    public function getGroups()
    {
        return $this->groups;
    }

    public function setGroups(ArrayCollection $groups): UserInterface
    {
        $this->groups = $groups;

        return $this;
    }

    public function getGroupNames(): ?array
    {
        $names = [];

        foreach ($this->getGroups() as $group) {
            $names[] = $group->getName();
        }

        return $names;
    }

    public function hasGroup(string $name): bool
    {
        return \in_array($name, $this->getGroupNames(), true);
    }

    public function addGroup(GroupInterface $group): UserInterface
    {
        if (!$this->getGroups()->contains($group)) {
            $this->getGroups()->add($group);
        }

        return $this;
    }

    public function removeGroup(GroupInterface $group): UserInterface
    {
        if ($this->getGroups()->contains($group)) {
            $this->getGroups()->removeElement($group);
        }

        return $this;
    }

    public function eraseCredentials()
    {
    }

    public function serialize()
    {
        return serialize([
            $this->id,
            $this->password,
            $this->email,
            $this->active,
            $this->lastLogin,
            $this->createdAt,
        ]);
    }

    public function unserialize($serialized)
    {
        [
            $this->id,
            $this->password,
            $this->email,
            $this->active,
            $this->lastLogin,
            $this->createdAt
        ] = unserialize($serialized);
    }
}
