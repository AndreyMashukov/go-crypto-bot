<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Controller\PublicRoute;

use Bundles\CryptoBotContext\Service\CachedTradeListServiceInterface;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;

/**
 * @Rest\Route("/public/cryptobot", name="public_crypto_bot_")
 */
class CryptoBotController extends AbstractFOSRestController
{
    private CachedTradeListServiceInterface $tradeListService;

    public function __construct(CachedTradeListServiceInterface $tradeListService)
    {
        $this->tradeListService = $tradeListService;
    }

    /**
     * @Rest\Route("/trades", methods={"GET"}, name="trades")
     * @Rest\View
     *
     * @return array
     */
    public function getTradesAction(): array
    {
        return $this->tradeListService->getPublicTradeList();
    }

    /**
     * @Rest\Route("/swaps", methods={"GET"}, name="swaps")
     * @Rest\View
     *
     * @return array
     */
    public function getSwapsAction(): array
    {
        return $this->tradeListService->getSwapList();
    }

    /**
     * @Rest\Route("/positions", methods={"GET"}, name="positions")
     * @Rest\View
     *
     * @return array
     */
    public function getPositionsAction(): array
    {
        return $this->tradeListService->getPublicPositionList();
    }
}
