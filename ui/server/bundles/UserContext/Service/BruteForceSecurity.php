<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Service;

use Bundles\UserContext\Entity\User;
use Bundles\UserContext\Exception\TooManyAuthAttemptsFailed;
use Symfony\Component\Cache\CacheItem;
use Symfony\Contracts\Cache\CacheInterface;

class BruteForceSecurity
{
    public const MAX_FAIL_AMOUNT = 3;

    private CacheInterface $cache;

    public function __construct(CacheInterface $cache)
    {
        $this->cache = $cache;
    }

    public function trackFail(User $user): void
    {
        $fails = $this->getFailedAmount($user);
        ++$fails;
        $this->invalidate($user);
        $this->cache->get($this->getKey($user), function (CacheItem $item) use ($fails) {
            $item->expiresAt(new \DateTimeImmutable('+1 hour'));

            return $fails;
        });

        if ($this->isBlocked($user)) {
            throw new TooManyAuthAttemptsFailed('Too many auth attempts failed, please try later.');
        }
    }

    public function getFailedAmount(User $user): int
    {
        return $this->cache->get($this->getKey($user), function (CacheItem $item) {
            $item->expiresAt(new \DateTimeImmutable('+1 hour'));

            return 0;
        });
    }

    public function isBlocked(User $user): bool
    {
        $fails = $this->getFailedAmount($user);

        return $fails >= self::MAX_FAIL_AMOUNT;
    }

    public function invalidate(User $user): void
    {
        $this->cache->delete($this->getKey($user));
    }

    private function getKey(User $user): string
    {
        return "auth-fail-{$user->getId()}";
    }
}
