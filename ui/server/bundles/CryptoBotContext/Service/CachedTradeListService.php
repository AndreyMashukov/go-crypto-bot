<?php
namespace Bundles\CryptoBotContext\Service;

use Symfony\Contracts\Cache\CacheInterface;
use Symfony\Contracts\Cache\ItemInterface;

class CachedTradeListService implements CachedTradeListServiceInterface
{
    public const PUBLIC_POSITIONS_CACHE_KEY = 'public_positions_cache_key';

    public const PUBLIC_TRADES_CACHE_KEY = 'public_trades_cache_key';

    public const PUBLIC_SWAPS_CACHE_KEY = 'public_swaps_cache_key';

    public const CACHE_TTL = '+30 seconds';

    public function __construct(private readonly TradeListService $tradeListService, private readonly CacheInterface $cache)
    {
    }

    public function getPublicPositionList(): array
    {
        return $this->cache->get(self::PUBLIC_POSITIONS_CACHE_KEY, function (ItemInterface $cache) {
            $cache->expiresAt(new \DateTimeImmutable(self::CACHE_TTL));

            return $this->tradeListService->getPublicPositionList();
        });
    }

    public function getPublicTradeList(): array
    {
        return $this->cache->get(self::PUBLIC_TRADES_CACHE_KEY, function (ItemInterface $cache) {
            $cache->expiresAt(new \DateTimeImmutable(self::CACHE_TTL));

            return $this->tradeListService->getPublicTradeList();
        });
    }

    public function getSwapList(): array
    {
        return $this->cache->get(self::PUBLIC_SWAPS_CACHE_KEY, function (ItemInterface $cache) {
            $cache->expiresAt(new \DateTimeImmutable(self::CACHE_TTL));

            return $this->tradeListService->getSwapList();
        });
    }

    public function invalidateCache()
    {
        $keys = [
            self::PUBLIC_TRADES_CACHE_KEY,
            self::PUBLIC_POSITIONS_CACHE_KEY,
            self::PUBLIC_SWAPS_CACHE_KEY,
        ];

        foreach ($keys as $key) {
            $this->cache->delete($key);
        }
    }
}
