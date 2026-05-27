<?php
namespace Bundles\OxaPayContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\Server;
use Bundles\CryptoBotContext\Repository\ServerRepository;
use Bundles\OxaPayContext\Model\PaidService;
use Bundles\TgBotContext\Service\AlertService;
use Bundles\UserContext\Entity\User;
use Doctrine\ORM\EntityManagerInterface;

class PaidServiceManager
{
    public const SIGNAL_SUBSCRIPTION_CODE = 'signal_subscription';

    public const BASIC_SUBSCRIPTION_CODE = 'basic_subscription';

    public const API_SUBSCRIPTION_CODE = 'api_subscription';

    public const DEDICATED_SERVER_BINANCE_CODE = 'dedicated_binance_server';

    public const DEDICATED_SERVER_BYBIT_CODE = 'dedicated_bybit_server';

    public function __construct(private readonly EntityManagerInterface $entityManager, private readonly ServerRepository $serverRepository, private readonly AlertService $alertService)
    {
    }

    /**
     * @return array|PaidService[]
     */
    public function getServiceList(User $user): array
    {
        $services = [];

        $hasBasicSubscription = $user->hasActiveBasicSubscription();
        $services[]           = new PaidService(
            self::BASIC_SUBSCRIPTION_CODE,
            $hasBasicSubscription,
            890.00, 30, $hasBasicSubscription ? $user->getBasicSubscriptionExpiresAt() : null
        );

        $hasSignalSubscription = $user->hasActiveSignalSubscription();
        $services[]            = new PaidService(
            self::SIGNAL_SUBSCRIPTION_CODE,
            $hasSignalSubscription,
            20.00, 30, $hasSignalSubscription ? $user->getSignalSubscriptionExpiresAt() : null
        );

        if (($binanceBot = $user->getBinanceBot()) instanceof CryptoBot) {
            $hasDedicatedBinance = $binanceBot->hasDedicatedServer() && $binanceBot->hasActiveDedicatedServerSubscription();

            $services[] = new PaidService(
                self::DEDICATED_SERVER_BINANCE_CODE,
                $hasDedicatedBinance,
                15.00,
                30,
                $hasDedicatedBinance ? $binanceBot->getDedicatedServerExpiresAt() : null
            );
        }

        if (($bybitBot = $user->getByBitBot()) instanceof CryptoBot) {
            $hasDedicatedByBit = $bybitBot->hasDedicatedServer() && $bybitBot->hasActiveDedicatedServerSubscription();

            $services[] = new PaidService(
                self::DEDICATED_SERVER_BYBIT_CODE,
                $hasDedicatedByBit,
                15.00,
                30,
                $hasDedicatedByBit ? $bybitBot->getDedicatedServerExpiresAt() : null
            );
        }

        $hasApiSubscription = $user->hasActiveApiSubscription();
        $services[]         = new PaidService(
            self::API_SUBSCRIPTION_CODE,
            $hasApiSubscription,
            89.00, 30, $hasApiSubscription ? $user->getApiSubscriptionExpiresAt() : null
        );

        return $services;
    }

    public function getPaidService(string $code, User $user): PaidService
    {
        foreach ($this->getServiceList($user) as $item) {
            if ($item->getCode() === $code) {
                return $item;
            }
        }

        throw new \BadMethodCallException("No such service: {$code}.");
    }

    public function purchase(User $user, PaidService $paidService): void
    {
        $price = $paidService->getPrice();
        if ($price > $user->getBudget()) {
            $this->alertService->alert("PaidServiceManager: User {$user->getEmail()} does not have enough balance for service {$paidService->getCode()}");

            throw new \BadMethodCallException('Not enough budget to pay the service.');
        }

        $paidService = $this->getPaidService($paidService->getCode(), $user);
        $expiresAt   = $paidService->getExpiresAt();
        if (!$expiresAt instanceof \DateTimeImmutable) {
            $expiresAt = new \DateTimeImmutable();
        }

        $expiresAt = $expiresAt->add(new \DateInterval("P{$paidService->getDays()}D"));

        $this->entityManager->wrapInTransaction(function () use ($expiresAt, $user, $paidService) {
            if (self::SIGNAL_SUBSCRIPTION_CODE === $paidService->getCode()) {
                $user->setSignalSubscriptionExpiresAt($expiresAt);
            }
            if (self::BASIC_SUBSCRIPTION_CODE === $paidService->getCode()) {
                $user->setBasicSubscriptionExpiresAt($expiresAt);
            }
            if (self::API_SUBSCRIPTION_CODE === $paidService->getCode()) {
                $user->setApiSubscriptionExpiresAt($expiresAt);
            }
            if (\in_array($paidService->getCode(), [self::DEDICATED_SERVER_BYBIT_CODE, self::DEDICATED_SERVER_BINANCE_CODE], true)) {
                $exchangeBot = null;
                if (self::DEDICATED_SERVER_BINANCE_CODE === $paidService->getCode()) {
                    $exchangeBot = $user->getBinanceBot();
                }
                if (self::DEDICATED_SERVER_BYBIT_CODE === $paidService->getCode()) {
                    $exchangeBot = $user->getByBitBot();
                }
                if (!$exchangeBot instanceof CryptoBot) {
                    throw new \BadMethodCallException("Unable to find bot for service: {$paidService->getCode()}");
                }

                $exchangeBot->setDedicatedServerExpiresAt($expiresAt);
                if (!$exchangeBot->getDedicated() instanceof Server) {
                    $dedicatedServer = $this->serverRepository->getAvailableServer($exchangeBot);
                    if (!$dedicatedServer instanceof Server) {
                        throw new \BadMethodCallException("Unable to provide dedicated server for bot #{$exchangeBot->getId()}");
                    }
                    $exchangeBot->setDedicated($dedicatedServer);
                }
            }

            $this->alertService->alert("PaidServiceManager: service {$paidService->getCode()} is paid by User {$user->getEmail()}");

            $user->setBudget($user->getBudget() - $paidService->getPrice());
        });
    }
}
