<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use GuzzleHttp\ClientInterface;
use Psr\Cache\CacheItemInterface;
use Psr\Log\LoggerInterface;
use Symfony\Contracts\Cache\CacheInterface;

class PivotService
{
    private ClientInterface $client;

    private CacheInterface $cache;

    private LoggerInterface $logger;

    public function __construct(
        ClientInterface $client,
        CacheInterface $cache,
        LoggerInterface $logger
    ) {
        $this->client = $client;
        $this->cache  = $cache;
        $this->logger = $logger;
    }

    public function getPivotGrid(CryptoBot $cryptoBot): array
    {
        return $this->cache->get('pivot-result-grid', function (CacheItemInterface $item) use ($cryptoBot) {
            try {
                $response = $this->client->request(
                    'GET',
                    "http://localhost:8080/stats/pivot/grid?exchange={$cryptoBot->getProvider()}"
                );
                $json = json_decode($response->getBody()->getContents(), true);
                if (!$json) {
                    $item->expiresAt(new \DateTimeImmutable('now'));

                    return [];
                }

                $item->expiresAt(new \DateTimeImmutable('+5 minutes'));

                return $json;
            } catch (\Throwable $exception) {
                $this->logger->error($exception->getMessage(), [
                    'file' => $exception->getFile(),
                    'line' => $exception->getLine(),
                    'code' => $exception->getCode(),
                ]);
                $item->expiresAt(new \DateTimeImmutable('now'));

                return [];
            }
        });
    }

    public function getStats(CryptoTradeConfig $config, int $intervalMinutes, int $periodDays): array
    {
        return $this->cache->get("pivot-result-{$config->getSymbol()}-{$intervalMinutes}-{$periodDays}", function (CacheItemInterface $item) use ($config, $periodDays, $intervalMinutes) {
            $cryptoBot = $config->getCryptobot();

            try {
                $response = $this->client->request(
                    'GET',
                    "http://localhost:8080/stats/pivot/{$config->getSymbol()}?interval={$intervalMinutes}&periodDays={$periodDays}&exchange={$cryptoBot->getProvider()}"
                );
                $json = json_decode($response->getBody()->getContents(), true);
                if (!$json) {
                    $item->expiresAt(new \DateTimeImmutable('now'));

                    return [];
                }

                $item->expiresAt(new \DateTimeImmutable('+5 minutes'));

                return $json;
            } catch (\Throwable $exception) {
                $this->logger->error($exception->getMessage(), [
                    'file' => $exception->getFile(),
                    'line' => $exception->getLine(),
                    'code' => $exception->getCode(),
                ]);
                $item->expiresAt(new \DateTimeImmutable('now'));

                return [];
            }
        });
    }
}
