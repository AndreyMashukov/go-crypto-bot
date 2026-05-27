<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\Server;
use Bundles\TgBotContext\Service\AlertService;
use GuzzleHttp\ClientInterface;
use Psr\Log\LoggerInterface;

class DeployService extends AbstractHttpService
{
    private HealthCheckService $healthCheckService;

    private AlertService $alertService;

    public function __construct(
        ClientInterface $client,
        LoggerInterface $logger,
        HealthCheckService $healthCheckService,
        AlertService $alertService
    ) {
        parent::__construct($client, $logger);

        $this->healthCheckService = $healthCheckService;
        $this->alertService       = $alertService;
    }

    public function stop(CryptoBot $cryptoBot, Server $server): void
    {
        $json = $this->request(
            'POST',
            "http://{$server->getIp()}:8095/stop",
            [
                'botUuid'          => $cryptoBot->getUuid(),
            ]
        );

        if (isset($json['status']) && 'OK' === $json['status']) {
            $cryptoBot->setStatus(CryptoBot::STATUS_STOPPED);

            if (!$cryptoBot->isDedicated($server)) {
                $cryptoBot->setServer(null);
            }

            return;
        }

        throw new \RuntimeException("Couldn't stop container");
    }

    public function deploy(CryptoBot $cryptoBot, Server $server): bool
    {
        $binanceApiKey    = '';
        $binanceApiSecret = '';
        $bybitApiKey      = '';
        $bybitApiSecret   = '';

        switch ($cryptoBot->getProvider()) {
            case CryptoBot::PROVIDER_BINANCE:
                $binanceApiKey    = $cryptoBot->getApiKey();
                $binanceApiSecret = $cryptoBot->getApiSecret();
                break;
            case CryptoBot::PROVIDER_BYBIT:
                $bybitApiKey    = $cryptoBot->getApiKey();
                $bybitApiSecret = $cryptoBot->getApiSecret();
                break;
            default:
                throw new \BadMethodCallException("Wrong provider given: {$cryptoBot->getProvider()}");
        }

        $json = $this->request(
            'POST',
            "http://{$server->getIp()}:8095/deploy",
            [
                'botUuid'          => $cryptoBot->getUuid(),
                'botExchange'      => $cryptoBot->getProvider(),
                'binanceApiKey'    => $binanceApiKey,
                'binanceApiSecret' => $binanceApiSecret,
                'bybitApiKey'      => $bybitApiKey,
                'bybitApiSecret'   => $bybitApiSecret,
                'isTestNet'        => $cryptoBot->isTest(),
            ]
        );

        $cryptoBot->setServer($server);
        $cryptoBot->setContainerId($json['containerId']);
        $cryptoBot->setPort($json['port']);

        $health   = [];
        $attempts = 0;

        do {
            try {
                $health = $this->healthCheckService->healthCheck($cryptoBot);
                if ('api_key_checking' === $health['binanceStatus']) {
                    $health = [];
                    throw new \LogicException('API key checking...');
                }
            } catch (\Throwable $exception) {
                sleep(1);
                ++$attempts;
            }
        } while ($attempts < 15 && !$health);

        $user = $cryptoBot->getUser();

        if (!$health) {
            $errorMessage = 'Bot was not deployed, please try later.';
            $cryptoBot->setErrorMessage($errorMessage);
            $this->alertService->alert("DeployService: Bot #{$cryptoBot->getId()} ({$cryptoBot->getProvider()}) User {$user->getEmail()} - {$errorMessage}");

            return false;
        }

        if ('ok' !== $health['dbStatus']) {
            $errorMessage = 'Database (MySQL connection) error.';
            $cryptoBot->setErrorMessage($errorMessage);
            $this->alertService->alert("DeployService: Bot #{$cryptoBot->getId()} ({$cryptoBot->getProvider()}) User {$user->getEmail()} - {$errorMessage}");

            return false;
        }

        if ('ok' !== $health['redisStatus']) {
            $errorMessage = 'Cache (Redis connection) error.';
            $cryptoBot->setErrorMessage($errorMessage);
            $this->alertService->alert("DeployService: Bot #{$cryptoBot->getId()} ({$cryptoBot->getProvider()}) User {$user->getEmail()} - {$errorMessage}");

            return false;
        }

        if ('ok' !== $health['binanceStatus']) {
            $errorMessage = 'Exchange (Binance API connection) error.';
            $cryptoBot->setErrorMessage($errorMessage);
            $this->alertService->alert("DeployService: Bot #{$cryptoBot->getId()} ({$cryptoBot->getProvider()}) User {$user->getEmail()} - {$errorMessage}");

            return false;
        }

        $cryptoBot->setStatus(CryptoBot::STATUS_RUNNING);
        $this->alertService->alert("DeployService: Bot #{$cryptoBot->getId()} ({$cryptoBot->getProvider()}) User {$user->getEmail()} is deployed.");

        return true;
    }
}
