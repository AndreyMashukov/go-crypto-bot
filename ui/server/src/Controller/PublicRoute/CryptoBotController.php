<?php

declare(strict_types=1);

namespace App\Controller\PublicRoute;

use Bundles\CryptoBotContext\Service\CachedTradeListServiceInterface;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;

class CryptoBotController extends AbstractFOSRestController
{
    public function __construct(private readonly CachedTradeListServiceInterface $tradeListService)
    {
    }

    /**
     * @Rest\Route("/trades", methods={"GET"}, name="trades")
     * @Rest\View
     */
    public function getTrades(): array
    {
        return $this->tradeListService->getPublicTradeList();
    }

    /**
     * @Rest\Route("/swaps", methods={"GET"}, name="swaps")
     * @Rest\View
     */
    public function getSwaps(): array
    {
        return $this->tradeListService->getSwapList();
    }

    /**
     * @Rest\Route("/positions", methods={"GET"}, name="positions")
     * @Rest\View
     */
    public function getPositions(): array
    {
        return $this->tradeListService->getPublicPositionList();
    }
}
