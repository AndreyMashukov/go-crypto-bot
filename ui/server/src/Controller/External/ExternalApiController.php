<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Controller\External;

use Bundles\CryptoBotContext\Repository\CryptoTradeConfigRepository;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use Sensio\Bundle\FrameworkExtraBundle\Configuration\Security;
use Swagger\Annotations as SWG;
use Symfony\Component\HttpFoundation\Request;

/**
 * @Rest\Route("/external/api", name="external_api_")
 */
class ExternalApiController extends AbstractFOSRestController
{
    private CryptoTradeConfigRepository $configRepository;

    public function __construct(CryptoTradeConfigRepository $configRepository)
    {
        $this->configRepository = $configRepository;
    }

    /**
     * @Rest\Route("/sentiment/list", methods={"GET"}, name="sentiment_list")
     * @Rest\View
     * @SWG\Get(
     *     description="Coin Sentiment Analysis information",
     *
     *     @SWG\Parameter(
     *         name="coin",
     *         in="query",
     *         type="string",
     *         required=false,
     *         description="Coin to filter list"
     *     ),
     *     @SWG\Parameter(
     *         name="Authorization",
     *         in="header",
     *         type="string",
     *         required=true,
     *         description="Authorization Token"
     *     ),
     *
     *     @SWG\Response(
     *         response=200,
     *         description="Coin Sentiment Analysis result list",
     *
     *         @SWG\Schema(
     *             type="array",
     *             @SWG\Items(ref="#/definitions/CoinSentimentAnalysis")
     *         )
     *     ),
     *
     *     @SWG\Response(
     *         response=404,
     *         description="Not Found",
     *     ),
     *     @SWG\Response(
     *         response=403,
     *         description="Access Denied",
     *     ),
     * )
     * @Security("is_granted('ROLE_API')")
     * @SWG\Tag(name="Sentiment Analysis")
     *
     * @param Request $request
     */
    public function getSentimentAction(Request $request): array
    {
        $coin = $request->get('coin');
        $list = $this->configRepository->getSentimentList();

        if ($coin) {
            foreach ($list as $key => $item) {
                if ($item['coin'] !== $coin) {
                    unset($list[$key]);
                }
            }
        }

        return $list;
    }
}
