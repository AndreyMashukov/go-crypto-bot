<?php

declare(strict_types=1);

namespace App\Controller\External;

use Bundles\CryptoBotContext\Repository\CryptoTradeConfigRepository;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use Symfony\Component\HttpFoundation\Request;

class ExternalApiController extends AbstractFOSRestController
{
    public function __construct(private readonly CryptoTradeConfigRepository $configRepository)
    {
    }

    public function getSentiment(Request $request): array
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
