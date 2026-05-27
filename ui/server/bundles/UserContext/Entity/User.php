<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Entity;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\OxaPayContext\Entity\PromoCode;
use Bundles\OxaPayContext\Entity\Transaction;
use Bundles\UserContext\Model\User as BaseUser;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\Common\Collections\Criteria;
use Doctrine\ORM\Mapping as ORM;
use JMS\Serializer\Annotation as Serializer;
use Pd\UserBundle\Model\ProfileInterface;
use Symfony\Bridge\Doctrine\Validator\Constraints\UniqueEntity;
use Symfony\Component\Security\Core\User\PasswordAuthenticatedUserInterface;
use Symfony\Component\Validator\Constraints as Assert;

/**
 * @ORM\Table(name="user")
 * @ORM\Entity(repositoryClass="Bundles\UserContext\Repository\UserRepository")
 *
 * @Serializer\ExclusionPolicy(Serializer\ExclusionPolicy::ALL)
 *
 * @SuppressWarnings(PHPMD)
 *
 * @UniqueEntity(fields={"username"}, entityClass=User::class)
 * @UniqueEntity(fields={"nickname"}, entityClass=User::class)
 */
class User extends BaseUser implements ProfileInterface, PasswordAuthenticatedUserInterface
{
    public const ROLE_PARTNER = 'ROLE_PARTNER';

    public const ROLE_API = 'ROLE_API';

    /**
     * @ORM\Column(type="integer")
     * @ORM\Id
     * @ORM\GeneratedValue(strategy="AUTO")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"admin"})
     */
    protected $id;

    /**
     * @ORM\Column(type="string", length=100, nullable=true)
     */
    protected $email;

    /**
     * @var string
     *
     * @Assert\Length(max="40")
     *
     * @ORM\Column(name="username", type="string", length=40, unique=true)
     */
    protected string $username;

    /**
     * @ORM\Column(type="json")
     */
    protected $roles;

    // todo: remove and implement self interface
    protected $profile;

    /**
     * @ORM\Column(type="string", length=255, nullable=true)
     */
    protected $lastLoginIp;

    /**
     * @var null|string
     *
     * @Assert\NotBlank
     * @Assert\Length(max="70")
     *
     * @ORM\Column(name="nickname", type="string", length=70)
     */
    protected ?string $nickname = null;

    /**
     * @var null|string
     *
     * @ORM\Column(name="phone", type="string", length=15, nullable=true)
     */
    protected ?string $phone = null;

    /**
     * @var null|string
     *
     * @ORM\Column(name="language", type="string", length=2, nullable=true)
     */
    protected ?string $language = null;

    /**
     * @ORM\Column(type="datetime_immutable")
     */
    protected $createdAt;

    protected string $plainPassword = '';

    /**
     * @ORM\OneToMany(targetEntity=CryptoBot::class, mappedBy="user")
     */
    private $cryptoBots;

    /**
     * @ORM\Column(type="float")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"user_extended", "admin"})
     */
    private float $budget = 0.00;

    /**
     * @ORM\Column(type="float")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"user_extended", "admin"})
     */
    private float $partnerBudget = 0.00;

    /**
     * @ORM\OneToMany(targetEntity=Transaction::class, mappedBy="user")
     */
    private $transactions;

    /**
     * @Serializer\Expose
     * @Serializer\Groups(groups={"user_extended", "admin"})
     *
     * @ORM\Column(type="datetime_immutable", nullable=true)
     */
    private ?\DateTimeImmutable $signalSubscriptionExpiresAt = null;

    /**
     * @Serializer\Expose
     * @Serializer\Groups(groups={"user_extended", "admin"})
     *
     * @ORM\Column(type="datetime_immutable", nullable=true)
     */
    private ?\DateTimeImmutable $basicSubscriptionExpiresAt = null;

    /**
     * @ORM\Column(type="boolean", options={"default": 1})
     */
    private bool $publicTradeView = true;

    /**
     * @ORM\ManyToOne(targetEntity="Bundles\OxaPayContext\Entity\PromoCode", inversedBy="users")
     * @ORM\JoinColumn(name="promo_code", nullable=true, referencedColumnName="pcd_id")
     */
    private ?PromoCode $promoCode = null;

    /**
     * @ORM\Column(type="datetime_immutable", nullable=true)
     */
    private ?\DateTimeImmutable $apiSubscriptionExpiresAt = null;

    public function __construct()
    {
        parent::__construct();

        $now                = new \DateTimeImmutable();
        $this->createdAt    = $now;
        $this->cryptoBots   = new ArrayCollection();
        $this->transactions = new ArrayCollection();
    }

    public function getId(): int
    {
        return $this->id ?: 0;
    }

    public function getUserIdentifier(): string
    {
        return $this->username;
    }

    public function isEnabled(): bool
    {
        return true;
    }

    /**
     * @Serializer\VirtualProperty("username")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"user_extended", "admin"})
     *
     * @return string
     */
    public function getUsername(): string
    {
        return $this->username;
    }

    /**
     * @Serializer\VirtualProperty("nickname")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"user_extended", "admin"})
     *
     * @return string
     */
    public function getNickname(): ?string
    {
        return $this->nickname;
    }

    public function setNickname(?string $nickname): self
    {
        $this->nickname = $nickname;

        return $this;
    }

    /**
     * @Serializer\VirtualProperty("phone")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"user_extended", "admin"})
     *
     * @return string
     */
    public function getPhone(): ?string
    {
        return $this->phone;
    }

    /**
     * @Serializer\VirtualProperty("email")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"user_extended", "admin"})
     *
     * @return string
     */
    public function getEmail(): ?string
    {
        return $this->email;
    }

    public function getPassword(): string
    {
        return $this->password ?: '';
    }

    public function getPlainPassword(): string
    {
        return $this->plainPassword;
    }

    public function setPlainPassword(string $plainPassword): self
    {
        $this->plainPassword = $plainPassword;

        return $this;
    }

    public function getCreatedAt(): ?\DateTime
    {
        return \DateTime::createFromImmutable($this->createdAt);
    }

    public function getCreatedAtImmutable(): ?\DateTimeImmutable
    {
        return $this->createdAt;
    }

    /**
     * {@inheritDoc}
     */
    public function getWebsite(): ?string
    {
        return null;
    }

    /**
     * {@inheritDoc}
     */
    public function setWebsite(?string $website): ProfileInterface
    {
        unset($website);

        return $this;
    }

    /**
     * {@inheritDoc}
     */
    public function getCompany(): ?string
    {
        return null;
    }

    /**
     * {@inheritDoc}
     */
    public function setCompany(string $company): ProfileInterface
    {
        unset($company);

        return $this;
    }

    /**
     * @Serializer\VirtualProperty("language")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"user_extended", "admin"})
     *
     * @return string
     */
    public function getLanguage(): ?string
    {
        return $this->language ?: 'ru';
    }

    /**
     * {@inheritDoc}
     */
    public function setLanguage(?string $language): ProfileInterface
    {
        $this->language = $language;

        return $this;
    }

    public function setPhone(?string $phone): ProfileInterface
    {
        $this->phone = $phone;

        return $this;
    }

    public function setUsername(string $username): \Pd\UserBundle\Model\UserInterface
    {
        $this->username = $username;

        return $this;
    }

    /**
     * @Serializer\VirtualProperty("roles")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"user_extended", "admin"})
     *
     * @return array
     */
    public function getRoles(): ?array
    {
        $roles = parent::getRoles() ?: [];

        if ($this->hasActiveApiSubscription()) {
            $roles[] = self::ROLE_API;
        }

        return array_unique($roles);
    }

    /**
     * @Serializer\VirtualProperty("uid")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"user_public"})
     *
     * @return string
     */
    public function getUid(): string
    {
        return $this->id;
    }

    public function isAdmin(): bool
    {
        return \in_array(self::ROLE_ADMIN, $this->getRoles(), true);
    }

    /**
     * @return Collection<int, CryptoBot>
     */
    public function getCryptoBots(): Collection
    {
        return $this->cryptoBots;
    }

    public function addCryptoBot(CryptoBot $cryptoBot): self
    {
        if (!$this->cryptoBots->contains($cryptoBot)) {
            $this->cryptoBots[] = $cryptoBot;
            $cryptoBot->setUser($this);
        }

        return $this;
    }

    public function removeCryptoBot(CryptoBot $cryptoBot): self
    {
        if ($this->cryptoBots->removeElement($cryptoBot)) {
            // set the owning side to null (unless already changed)
            if ($cryptoBot->getUser() === $this) {
                $cryptoBot->setUser(null);
            }
        }

        return $this;
    }

    public function getFirstName(): ?string
    {
        return '';
    }

    public function getLastName(): ?string
    {
        return '';
    }

    public function getFullName(): ?string
    {
        return '';
    }

    public function setLastName(?string $lastname): ProfileInterface
    {
        unset($lastname);

        return $this;
    }

    public function setFirstName(?string $firstname): ProfileInterface
    {
        unset($firstname);

        return $this;
    }

    /**
     * @Serializer\VirtualProperty("hasActiveBasicSubscription")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"user_extended"})
     *
     * @return bool
     */
    public function hasActiveBasicSubscription(): bool
    {
        if (!$this->basicSubscriptionExpiresAt instanceof \DateTimeImmutable) {
            return false;
        }

        return $this->basicSubscriptionExpiresAt->getTimestamp() > (new \DateTimeImmutable())->getTimestamp();
    }

    public function getMaxPairs(): int
    {
        if (1 === $this->id) {
            return 50;
        }

        if ($this->hasActiveBasicSubscription()) {
            return 30;
        }

        return 5;
    }

    public function getBudget(): ?float
    {
        return $this->budget;
    }

    public function setBudget(float $budget): self
    {
        $this->budget = $budget;

        return $this;
    }

    /**
     * @return Collection<int, Transaction>
     */
    public function getTransactions(): Collection
    {
        return $this->transactions;
    }

    public function addTransaction(Transaction $transaction): self
    {
        if (!$this->transactions->contains($transaction)) {
            $this->transactions[] = $transaction;
            $transaction->setUser($this);
        }

        return $this;
    }

    public function removeTransaction(Transaction $transaction): self
    {
        if ($this->transactions->removeElement($transaction)) {
            // set the owning side to null (unless already changed)
            if ($transaction->getUser() === $this) {
                $transaction->setUser(null);
            }
        }

        return $this;
    }

    public function getSignalSubscriptionExpiresAt(): ?\DateTimeImmutable
    {
        return $this->signalSubscriptionExpiresAt;
    }

    public function setSignalSubscriptionExpiresAt(?\DateTimeImmutable $signalSubscriptionExpiresAt): self
    {
        $this->signalSubscriptionExpiresAt = $signalSubscriptionExpiresAt;

        return $this;
    }

    /**
     * @Serializer\VirtualProperty("hasActiveSignalSubscription")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"user_extended", "admin"})
     *
     * @return bool
     */
    public function hasActiveSignalSubscription(): bool
    {
        if (!$this->signalSubscriptionExpiresAt instanceof \DateTimeImmutable) {
            return false;
        }

        return $this->signalSubscriptionExpiresAt->getTimestamp() > (new \DateTimeImmutable())->getTimestamp();
    }

    /**
     * @Serializer\VirtualProperty("isFreeze")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"admin"})
     *
     * @return bool
     */
    public function isFreeze(): bool
    {
        return parent::isFreeze();
    }

    public function isPublicTradeView(): ?bool
    {
        return $this->publicTradeView;
    }

    public function setPublicTradeView(bool $publicTradeView): self
    {
        $this->publicTradeView = $publicTradeView;

        return $this;
    }

    public function getPromoCode(): ?PromoCode
    {
        return $this->promoCode;
    }

    public function setPromoCode(?PromoCode $promoCode): self
    {
        $this->promoCode = $promoCode;

        return $this;
    }

    public function getPartnerBudget(): float
    {
        return $this->partnerBudget;
    }

    public function setPartnerBudget(float $partnerBudget): self
    {
        $this->partnerBudget = $partnerBudget;

        return $this;
    }

    public function getPromoCodePartner(): ?self
    {
        $promoCode = $this->getPromoCode();

        if (!$promoCode instanceof PromoCode) {
            return null;
        }

        return $promoCode->getPartner();
    }

    /**
     * @Serializer\VirtualProperty("isPartner")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"user_extended", "admin"})
     *
     * @return bool
     */
    public function isPartner(): bool
    {
        return $this->hasRole(self::ROLE_PARTNER);
    }

    public function getBasicSubscriptionExpiresAt(): ?\DateTimeImmutable
    {
        return $this->basicSubscriptionExpiresAt;
    }

    public function setBasicSubscriptionExpiresAt(?\DateTimeImmutable $basicSubscriptionExpiresAt): self
    {
        $this->basicSubscriptionExpiresAt = $basicSubscriptionExpiresAt;

        return $this;
    }

    public function getBinanceBot(): ?CryptoBot
    {
        $criteria = Criteria::create();
        $criteria->andWhere(Criteria::expr()->eq('provider', CryptoBot::PROVIDER_BINANCE));

        return $this->cryptoBots->matching($criteria)->first() ?: null;
    }

    public function getByBitBot(): ?CryptoBot
    {
        $criteria = Criteria::create();
        $criteria->andWhere(Criteria::expr()->eq('provider', CryptoBot::PROVIDER_BYBIT));

        return $this->cryptoBots->matching($criteria)->first() ?: null;
    }

    public function getApiSubscriptionExpiresAt(): ?\DateTimeImmutable
    {
        return $this->apiSubscriptionExpiresAt;
    }

    public function setApiSubscriptionExpiresAt(?\DateTimeImmutable $apiSubscriptionExpiresAt): self
    {
        $this->apiSubscriptionExpiresAt = $apiSubscriptionExpiresAt;

        return $this;
    }

    /**
     * @Serializer\VirtualProperty("hasActiveApiSubscription")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"user_extended", "admin"})
     *
     * @return bool
     */
    public function hasActiveApiSubscription(): bool
    {
        if (!$this->apiSubscriptionExpiresAt instanceof \DateTimeImmutable) {
            return false;
        }

        return $this->apiSubscriptionExpiresAt->getTimestamp() > (new \DateTimeImmutable())->getTimestamp();
    }
}
