<?php

declare(strict_types=1);

namespace Bundles\UserContext\Entity;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\OxaPayContext\Entity\PromoCode;
use Bundles\OxaPayContext\Entity\Transaction;
use Bundles\UserContext\Model\User as BaseUser;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\Common\Collections\Criteria;
use Doctrine\ORM\Mapping as ORM;
use Symfony\Component\Validator\Constraints as Assert;

class User extends BaseUser
{
    public const ROLE_PARTNER = 'ROLE_PARTNER';

    public const ROLE_API = 'ROLE_API';

    protected ?int $id = null;

    protected ?string $email = null;

    #[Assert\Length(max: 40)]
    #[ORM\Column(name: 'username', type: 'string', length: 40, unique: true)]
    protected string $username;

    /** @var list<string> */
    protected array $roles = [];

    protected ?string $lastLoginIp = null;

    #[Assert\NotBlank]
    #[Assert\Length(max: 70)]
    #[ORM\Column(name: 'nickname', type: 'string', length: 70)]
    protected ?string $nickname = null;

    #[ORM\Column(name: 'phone', type: 'string', length: 15, nullable: true)]
    protected ?string $phone = null;

    #[ORM\Column(name: 'language', type: 'string', length: 2, nullable: true)]
    protected ?string $language = null;

    protected ?\DateTimeInterface $createdAt = null;

    protected string $plainPassword = '';

    private $cryptoBots;

    private float $budget = 0.00;

    private float $partnerBudget = 0.00;

    private $transactions;

    private ?\DateTimeImmutable $signalSubscriptionExpiresAt = null;

    private ?\DateTimeImmutable $basicSubscriptionExpiresAt = null;

    private bool $publicTradeView = true;

    private ?PromoCode $promoCode = null;

    private ?\DateTimeImmutable $apiSubscriptionExpiresAt = null;

    public function __construct()
    {
        parent::__construct();

        $now                = new \DateTimeImmutable();
        $this->createdAt    = $now;
        $this->cryptoBots   = new ArrayCollection();
        $this->transactions = new ArrayCollection();
    }

    #[\Override]
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

    public function getUsername(): string
    {
        return $this->username;
    }

    /**
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
     * @return string
     */
    public function getPhone(): ?string
    {
        return $this->phone;
    }

    /**
     * @return string
     */
    #[\Override]
    public function getEmail(): ?string
    {
        return $this->email;
    }

    #[\Override]
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

    #[\Override]
    public function getCreatedAt(): ?\DateTime
    {
        return \DateTime::createFromImmutable($this->createdAt);
    }

    public function getCreatedAtImmutable(): ?\DateTimeImmutable
    {
        return $this->createdAt;
    }

    public function getWebsite(): ?string
    {
        return null;
    }

    public function setWebsite(?string $website): ProfileInterface
    {
        unset($website);

        return $this;
    }

    public function getCompany(): ?string
    {
        return null;
    }

    public function setCompany(string $company): ProfileInterface
    {
        unset($company);

        return $this;
    }

    /**
     * @return string
     */
    public function getLanguage(): ?string
    {
        return $this->language ?: 'ru';
    }

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

    public function setUsername(string $username): UserInterface
    {
        $this->username = $username;

        return $this;
    }

    /**
     * @return list<string>
     */
    #[\Override]
    public function getRoles(): array
    {
        $roles = parent::getRoles();

        if ($this->hasActiveApiSubscription()) {
            $roles[] = self::ROLE_API;
        }

        return \array_values(\array_unique($roles));
    }

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
        if ($this->cryptoBots->removeElement($cryptoBot) && $cryptoBot->getUser() === $this) {
            $cryptoBot->setUser(null);
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
        if ($this->transactions->removeElement($transaction) && $transaction->getUser() === $this) {
            $transaction->setUser(null);
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

    public function hasActiveSignalSubscription(): bool
    {
        if (!$this->signalSubscriptionExpiresAt instanceof \DateTimeImmutable) {
            return false;
        }

        return $this->signalSubscriptionExpiresAt->getTimestamp() > (new \DateTimeImmutable())->getTimestamp();
    }

    #[\Override]
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

    public function hasActiveApiSubscription(): bool
    {
        if (!$this->apiSubscriptionExpiresAt instanceof \DateTimeImmutable) {
            return false;
        }

        return $this->apiSubscriptionExpiresAt->getTimestamp() > (new \DateTimeImmutable())->getTimestamp();
    }
}
