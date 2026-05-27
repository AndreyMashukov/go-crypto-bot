<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Entity;

use Bundles\CryptoBotContext\Repository\CryptoBotRepository;
use Bundles\OxaPayContext\Service\PaidServiceManager;
use Bundles\UserContext\Entity\User;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\Common\Collections\Criteria;
use Doctrine\ORM\Mapping as ORM;
use JMS\Serializer\Annotation as Serializer;
use Ramsey\Uuid\Uuid;
use Symfony\Component\Validator\Constraints as Assert;
use Symfony\Component\Validator\Context\ExecutionContextInterface;

/**
 * @ORM\Entity(repositoryClass=CryptoBotRepository::class)
 *
 * @Serializer\ExclusionPolicy(Serializer\ExclusionPolicy::ALL)
 */
class CryptoBot
{
    public const PROVIDER_BINANCE = 'binance';

    public const PROVIDER_BYBIT = 'bybit';

    public const STATUS_NEW = 'new';

    public const STATUS_RUNNING = 'running';

    public const STATUS_STOPPED = 'stopped';

    public const STATUS_DEPLOY = 'deploy';

    public const STATUS_ERROR = 'error';

    /**
     * @ORM\Id
     * @ORM\GeneratedValue
     * @ORM\Column(name="ctb_id", type="integer")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot"})
     */
    private $id;

    /**
     * @ORM\Column(name="ctb_uuid", type="uuid")
     */
    private $uuid;

    /**
     * @ORM\Column(name="ctb_provider", type="string", length=50)
     *
     * @Assert\Choice(choices={"binance", "bybit"})
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot"})
     */
    private string $provider = self::PROVIDER_BINANCE;

    /**
     * @ORM\Column(name="ctb_api_key", type="string", length=255, nullable=true)
     *
     * @Assert\Regex(pattern="/^[a-z0-9]+$/ui")
     * @Assert\NotBlank
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot_secured"})
     */
    private $apiKey;

    /**
     * @ORM\Column(name="ctb_api_secret", type="string", length=255, nullable=true)
     *
     * @Assert\Regex(pattern="/^[a-z0-9]+$/ui")
     * @Assert\NotBlank
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot_secured"})
     */
    private $apiSecret;

    /**
     * @ORM\ManyToOne(targetEntity="Bundles\CryptoBotContext\Entity\Server", inversedBy="cryptoBots")
     * @ORM\JoinColumn(name="ctb_server", nullable=true, referencedColumnName="srv_id")
     */
    private ?Server $server = null;

    /**
     * @ORM\Column(name="ctb_port", type="string", length=255, nullable=true)
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot_secured"})
     */
    private $port;

    /**
     * @ORM\ManyToOne(targetEntity=User::class, inversedBy="cryptoBots")
     * @ORM\JoinColumn(name="ctb_user", nullable=false)
     */
    private $user;

    /**
     * @ORM\Column(name="ctb_container_id", type="string", length=255, nullable=true)
     */
    private $containerId;

    /**
     * @ORM\Column(name="ctb_status", type="string", length=255)
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot"})
     */
    private string $status = self::STATUS_NEW;

    /**
     * @ORM\OneToMany(targetEntity=CryptoTradeConfig::class, mappedBy="cryptobot", cascade={"persist", "remove"},
     * orphanRemoval=true)
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot"})
     */
    private Collection $cryptoTradeConfigs;

    /**
     * @ORM\Column(name="ctb_master", type="boolean")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot"})
     */
    private bool $master = false;

    /**
     * @ORM\Column(name="ctb_error_message", type="text", nullable=true)
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot"})
     */
    private ?string $errorMessage = null;

    /**
     * @ORM\ManyToOne(targetEntity="Bundles\CryptoBotContext\Entity\Server")
     * @ORM\JoinColumn(name="ctb_dedicated_server", nullable=true, referencedColumnName="srv_id")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot"})
     */
    private ?Server $dedicated = null;

    /**
     * @ORM\Column(name="ctb_restart_required", type="boolean", options={"default": 0})
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot"})
     */
    private bool $restartRequired = false;

    /**
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot"})
     *
     * @ORM\Column(name="ctb_dedicated_server_expires_at", type="datetime_immutable", nullable=true)
     */
    private ?\DateTimeImmutable $dedicatedServerExpiresAt = null;

    /**
     * @ORM\Column(name="ctb_test", type="boolean", options={"default": 0})
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot"})
     */
    private bool $test = false;

    public function __construct(User $user)
    {
        $this->uuid               = Uuid::uuid4();
        $this->user               = $user;
        $this->cryptoTradeConfigs = new ArrayCollection();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getUuid()
    {
        return $this->uuid;
    }

    public function setUuid($uuid): self
    {
        $this->uuid = $uuid;

        return $this;
    }

    public function getProvider(): ?string
    {
        return $this->provider;
    }

    public function setProvider(string $provider): self
    {
        $this->provider = $provider;

        return $this;
    }

    public function getApiKey(): ?string
    {
        return $this->apiKey;
    }

    public function setApiKey(string $apiKey): self
    {
        $this->apiKey = $apiKey;

        return $this;
    }

    public function getApiSecret(): ?string
    {
        return $this->apiSecret;
    }

    public function setApiSecret(string $apiSecret): self
    {
        $this->apiSecret = $apiSecret;

        return $this;
    }

    /**
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot"})
     * @Serializer\VirtualProperty("ipAddress")
     */
    public function getIpAddress(): ?string
    {
        if (!$this->server instanceof Server) {
            return null;
        }

        return $this->server->getIp();
    }

    public function setServer(?Server $server): self
    {
        $this->server = $server;

        return $this;
    }

    public function getPort(): ?string
    {
        return $this->port;
    }

    public function setPort(?string $port): self
    {
        $this->port = $port;

        return $this;
    }

    public function getUser(): ?User
    {
        return $this->user;
    }

    public function setUser(?User $user): self
    {
        $this->user = $user;

        return $this;
    }

    public function isOwnedBy(User $user): bool
    {
        return $this->getUser()->getId() === $user->getId();
    }

    public function getContainerId(): ?string
    {
        return $this->containerId;
    }

    public function setContainerId(string $containerId)
    {
        $this->containerId = $containerId;

        return $this;
    }

    public function getStatus(): string
    {
        return $this->status;
    }

    public function setStatus(string $status): self
    {
        $this->status = $status;

        return $this;
    }

    /**
     * @return Collection<int, CryptoTradeConfig>
     */
    public function getCryptoTradeConfigs(): Collection
    {
        return $this->cryptoTradeConfigs;
    }

    public function getActiveSymbols(): Collection
    {
        $criteria = Criteria::create();
        $criteria->andWhere($criteria->expr()->eq('enabled', true));

        return $this->cryptoTradeConfigs->matching($criteria);
    }

    public function getSymbolConfig(string $symbol): ?CryptoTradeConfig
    {
        $criteria = Criteria::create();
        $criteria->andWhere($criteria->expr()->eq('symbol', $symbol));

        return $this->cryptoTradeConfigs->matching($criteria)->first() ?: null;
    }

    public function addCryptoTradeConfig(CryptoTradeConfig $cryptoTradeConfig): self
    {
        if (!$this->cryptoTradeConfigs->contains($cryptoTradeConfig)) {
            $this->cryptoTradeConfigs[] = $cryptoTradeConfig;
            $cryptoTradeConfig->setCryptobot($this);
        }

        return $this;
    }

    public function removeCryptoTradeConfig(CryptoTradeConfig $cryptoTradeConfig): self
    {
        $this->cryptoTradeConfigs->removeElement($cryptoTradeConfig);

        return $this;
    }

    public function isRunning(): bool
    {
        return self::STATUS_RUNNING === $this->getStatus();
    }

    public function isPaid(): bool
    {
        $user = $this->getUser();

        if (!$user instanceof User) {
            return false;
        }

        return $user->hasActiveBasicSubscription() || $user->getBudget() >= 1;
    }

    /**
     * @Assert\Callback
     *
     * @param mixed $payload
     */
    public function validateLimits(ExecutionContextInterface $context, $payload): bool
    {
        $symbols = [];
        /** @var CryptoTradeConfig $config */
        foreach ($this->cryptoTradeConfigs as $config) {
            $symbols[] = $config->getSymbol();
        }

        if (\count(array_unique($symbols)) !== \count($symbols)) {
            $context->buildViolation('Duplicated symbols are not allowed. Please remove duplicates.')
                ->atPath('cryptoTradeConfigs')
                ->setCode(Assert\Count::TOO_MANY_ERROR)
                ->addViolation();

            return false;
        }

        return true;
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

    public function getServer(): ?Server
    {
        return $this->server;
    }

    public function getErrorMessage(): ?string
    {
        return $this->errorMessage;
    }

    public function setErrorMessage(?string $errorMessage): self
    {
        $this->errorMessage = $errorMessage;

        return $this;
    }

    public function isStopped(): bool
    {
        return self::STATUS_STOPPED === $this->getStatus();
    }

    public function getPairsHash(): string
    {
        $pairs = $this->getCryptoTradeConfigs()->toArray();
        usort($pairs, function (CryptoTradeConfig $a, CryptoTradeConfig $b) {
            return $a->getId() > $b->getId() ? 1 : -1;
        });

        $symbols = array_map(fn (CryptoTradeConfig $x) => $x->getSymbol(), $pairs);

        return md5(implode('-', $symbols));
    }

    public function getDedicated(): ?Server
    {
        return $this->dedicated;
    }

    public function setDedicated(?Server $dedicated): self
    {
        $this->dedicated = $dedicated;

        return $this;
    }

    public function isDedicated(Server $server): bool
    {
        $dedicated = $this->getDedicated();

        if (!$dedicated instanceof Server) {
            return false;
        }

        return $dedicated->getId() === $server->getId();
    }

    /**
     * @Serializer\VirtualProperty("hasActiveSignalSubscription")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot"})
     *
     * @return bool
     */
    public function hasActiveSignalSubscription(): bool
    {
        return $this->user->hasActiveSignalSubscription();
    }

    public function isRestartRequired(): bool
    {
        return $this->restartRequired;
    }

    public function setRestartRequired(bool $restartRequired): self
    {
        $this->restartRequired = $restartRequired;

        return $this;
    }

    /**
     * @Serializer\VirtualProperty("hasDedicatedServer")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot"})
     *
     * @return bool
     */
    public function hasDedicatedServer(): bool
    {
        return $this->getDedicated() instanceof Server;
    }

    /**
     * @Serializer\VirtualProperty("dedicatedPaidServiceCode")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot"})
     *
     * @return null|string
     */
    public function getDedicatedPaidServiceCode(): ?string
    {
        switch ($this->provider) {
            case self::PROVIDER_BINANCE:
                return PaidServiceManager::DEDICATED_SERVER_BINANCE_CODE;
            case self::PROVIDER_BYBIT:
                return PaidServiceManager::DEDICATED_SERVER_BYBIT_CODE;
            default:
                return null;
        }
    }

    public function getDedicatedServerExpiresAt(): ?\DateTimeImmutable
    {
        return $this->dedicatedServerExpiresAt;
    }

    public function setDedicatedServerExpiresAt(?\DateTimeImmutable $dedicatedServerExpiresAt): self
    {
        $this->dedicatedServerExpiresAt = $dedicatedServerExpiresAt;

        return $this;
    }

    /**
     * @Serializer\VirtualProperty("hasActiveDedicatedServerSubscription")
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptobot"})
     *
     * @return bool
     */
    public function hasActiveDedicatedServerSubscription(): bool
    {
        if (!$this->dedicatedServerExpiresAt instanceof \DateTimeImmutable) {
            return false;
        }

        return $this->dedicatedServerExpiresAt->getTimestamp() > (new \DateTimeImmutable())->getTimestamp();
    }

    public function isTest(): bool
    {
        return $this->test;
    }

    public function setTest(bool $test): self
    {
        $this->test = $test;

        return $this;
    }
}
