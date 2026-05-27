<?php

declare(strict_types=1);

namespace App\Controller\PublicRoute;

use Bundles\CryptoBotContext\Service\CachedTradeListServiceInterface;
use FOS\RestBundle\Controller\AbstractFOSRestController;

class CryptoBotController extends AbstractFOSRestController
{
    public function __construct(private readonly CachedTradeListServiceInterface $tradeListService)
    {
    }

    public function getTrades(): array
    {
        return $this->tradeListService->getPublicTradeList();
    }

    public function getSwaps(): array
    {
        return $this->tradeListService->getSwapList();
    }

    public function getPositions(): array
    {
        return $this->tradeListService->getPublicPositionList();
    }
}
