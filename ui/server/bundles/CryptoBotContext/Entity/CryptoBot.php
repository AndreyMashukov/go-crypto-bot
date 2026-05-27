<?php
namespace Bundles\CryptoBotContext\Entity;

use Bundles\OxaPayContext\Service\PaidServiceManager;
use Bundles\UserContext\Entity\User;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\Common\Collections\Criteria;
use Ramsey\Uuid\Uuid;
use Symfony\Component\Validator\Constraints as Assert;
use Symfony\Component\Validator\Context\ExecutionContextInterface;

class CryptoBot
{
    public const PROVIDER_BINANCE = 'binance';

    public const PROVIDER_BYBIT = 'bybit';

    public const STATUS_NEW = 'new';

    public const STATUS_RUNNING = 'running';

    public const STATUS_STOPPED = 'stopped';

    public const STATUS_DEPLOY = 'deploy';

    public const STATUS_ERROR = 'error';

    private ?int $id = null;

    private $uuid;

    private string $provider = self::PROVIDER_BINANCE;

    private ?string $apiKey = null;

    private ?string $apiSecret = null;

    private ?Server $server = null;

    private ?string $port = null;

    private ?string $containerId = null;

    private string $status = self::STATUS_NEW;

    private Collection $cryptoTradeConfigs;

    private bool $master = false;

    private ?string $errorMessage = null;

    private ?Server $dedicated = null;

    private bool $restartRequired = false;

    private ?\DateTimeImmutable $dedicatedServerExpiresAt = null;

    private bool $test = false;

    public function __construct(private User $user)
    {
        $this->uuid               = Uuid::uuid4();
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
     * @param mixed $payload
     */
    public function validateLimits(ExecutionContextInterface $context, $payload): bool
    {
        $symbols = [];
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
        usort($pairs, fn(CryptoTradeConfig $a, CryptoTradeConfig $b) => $a->getId() > $b->getId() ? 1 : -1);

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

    public function hasDedicatedServer(): bool
    {
        return $this->getDedicated() instanceof Server;
    }

    public function getDedicatedPaidServiceCode(): ?string
    {
        return match ($this->provider) {
            self::PROVIDER_BINANCE => PaidServiceManager::DEDICATED_SERVER_BINANCE_CODE,
            self::PROVIDER_BYBIT => PaidServiceManager::DEDICATED_SERVER_BYBIT_CODE,
            default => null,
        };
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
