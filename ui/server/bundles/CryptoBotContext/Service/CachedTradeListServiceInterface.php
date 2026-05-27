<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Service;

interface CachedTradeListServiceInterface
{
    public function getPublicPositionList(): array;

    public function getPublicTradeList(): array;

    public function getSwapList(): array;

    public function invalidateCache();
}
