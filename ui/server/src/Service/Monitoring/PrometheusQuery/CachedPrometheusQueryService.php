<?php

declare(strict_types=1);

namespace App\Service\Monitoring\PrometheusQuery;

use Symfony\Component\DependencyInjection\Attribute\AsDecorator;
use Symfony\Contracts\Cache\CacheInterface;
use Symfony\Contracts\Cache\ItemInterface;

#[AsDecorator(decorates: PrometheusQueryServiceInterface::class)]
final readonly class CachedPrometheusQueryService implements PrometheusQueryServiceInterface
{
    /**
     * Live charts (short windows) benefit from a sub-second cache to
     * absorb burst polling from many FE tabs. Longer windows hold for
     * 10 seconds since they hardly move per-second.
     */
    private const int LIVE_WINDOW_THRESHOLD_SECONDS = 600;
    private const int LIVE_TTL_SECONDS              = 1;
    private const int HISTORIC_TTL_SECONDS          = 10;

    public function __construct(
        private PrometheusQueryServiceInterface $inner,
        private CacheInterface $cache,
    ) {}

    public function queryRange(
        string $promQL,
        \DateTimeImmutable $from,
        \DateTimeImmutable $to,
        string $step,
    ): array {
        $window = max(0, $to->getTimestamp() - $from->getTimestamp());
        $ttl    = $window <= self::LIVE_WINDOW_THRESHOLD_SECONDS
            ? self::LIVE_TTL_SECONDS
            : self::HISTORIC_TTL_SECONDS;

        $key = $this->cacheKey($promQL, $from, $to, $step);

        return $this->cache->get(
            $key,
            function (ItemInterface $item) use ($promQL, $from, $to, $step, $ttl): array {
                $item->expiresAfter($ttl);

                return $this->inner->queryRange($promQL, $from, $to, $step);
            },
        );
    }

    private function cacheKey(
        string $promQL,
        \DateTimeImmutable $from,
        \DateTimeImmutable $to,
        string $step,
    ): string {
        return 'prometheus.range.'.hash('sha256', \sprintf(
            '%s|%d|%d|%s',
            $promQL,
            $from->getTimestamp(),
            $to->getTimestamp(),
            $step,
        ));
    }
}
