<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Service;

interface CachedTradeListServiceInterface
{
    public function getPublicPositionList(): array;

    public function getPublicTradeList(): array;

    public function getSwapList(): array;

    public function invalidateCache();
}
