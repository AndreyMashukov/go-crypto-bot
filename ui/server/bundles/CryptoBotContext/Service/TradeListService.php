<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Repository\CryptoBotRepository;
use Bundles\CryptoBotContext\Repository\TradeRepository;
use GuzzleHttp\Exception\GuzzleException;

class TradeListService
{
    private CryptoBotRepository $repository;

    private CryptoBotService $cryptoBotService;

    private TradeRepository $tradeRepository;

    private PivotService $pivotService;

    public function __construct(
        CryptoBotRepository $repository,
        CryptoBotService $cryptoBotService,
        TradeRepository $tradeRepository,
        PivotService $pivotService
    ) {
        $this->repository       = $repository;
        $this->cryptoBotService = $cryptoBotService;
        $this->tradeRepository  = $tradeRepository;
        $this->pivotService     = $pivotService;
    }

    public function getPublicPositionList(): array
    {
        $positions   = [];
        $runningBots = $this->repository->findBy([
            'status' => CryptoBot::STATUS_RUNNING,
            'test'   => false,
        ]);

        foreach ($runningBots as $cryptoBot) {
            $trader = $cryptoBot->getUser();
            if (!$trader->isPublicTradeView()) {
                continue;
            }

            try {
                $list = $this->cryptoBotService->getPositionsList($cryptoBot);
            } catch (\BadMethodCallException|\RuntimeException|\LogicException|GuzzleException $exception) {
                unset($exception);
                continue;
            }

            foreach ($list as $item) {
                $positions[] = [
                    'trader'   => $trader->getNickname(),
                    'position' => $item,
                    'exchange' => $cryptoBot->getProvider(),
                ];
            }
        }

        return $positions;
    }

    public function getBotPositionList(CryptoBot $cryptoBot): array
    {
        $positions = [];
        $result    = $this->cryptoBotService->getPositionsList($cryptoBot);
        $ratingMap = $this->tradeRepository->getBestMonthSymbols();

        $configMap = [];
        foreach ($cryptoBot->getCryptoTradeConfigs() as $config) {
            $configMap[$config->getSymbol()] = $config;
        }

        $pivotGrid = $this->pivotService->getPivotGrid($cryptoBot);

        foreach ($result as $item) {
            $config = $configMap[$item['symbol']] ?? null;
            $pivots = $pivotGrid[$item['symbol']] ?? [];

            $sentiment = null;

            if ($config instanceof CryptoTradeConfig && $config->getPositionUpdatedAt() instanceof \DateTimeImmutable) {
                $budget = $config->getAveragePrice() * $config->getQuantity();

                $profitFirst = ($config->getPriceTodayFirst() * $config->getQuantity()) - $budget;
                $profitLast  = ($config->getPriceTodayLast() * $config->getQuantity())  - $budget;

                $percentFirst = $profitFirst * 100 / $budget;
                $percentLast  = $profitLast  * 100 / $budget;

                $item['changeToday'] = [
                    'profit'     => round($profitLast - $profitFirst, 2),
                    'percent'    => round($percentLast - $percentFirst, 2),
                    'firstPrice' => $config->getPriceTodayFirst(),
                    'lastPrice'  => $config->getPriceTodayLast(),
                    'updatedAt'  => $config->getPositionUpdatedAt(),
                ];

                if ($config->getLabel()) {
                    $sentiment = [
                        'label' => $config->getLabel(),
                        'score' => $config->getScore(),
                    ];
                }
            } else {
                $item['changeToday'] = null;
            }

            $item['sentiment'] = $sentiment;
            $item['pivots']    = $pivots;
            $item['rating']    = $ratingMap[$item['symbol']] ?? null;
            $positions[]       = $item;
        }

        return $positions;
    }

    public function getPublicTradeList(): array
    {
        $tradeList   = [];
        $runningBots = $this->repository->findBy([
            'status' => CryptoBot::STATUS_RUNNING,
            'test'   => false,
        ]);

        foreach ($runningBots as $cryptoBot) {
            try {
                $list = $this->cryptoBotService->getTradeList($cryptoBot);
            } catch (\BadMethodCallException|\RuntimeException|\LogicException|GuzzleException $exception) {
                unset($exception);
                continue;
            }

            foreach ($list as $key => $trade) {
                $sellPrecision    = mb_strlen(explode('.', ((string) $trade['sell']))[1] ?? 0);
                $sellQtyPrecision = mb_strlen(explode('.', ((string) $trade['sellQuantity']))[1] ?? 0);

                $list[$key]['buy']         = round($trade['buy'], $sellPrecision);
                $list[$key]['buyQuantity'] = round($trade['buyQuantity'], $sellQtyPrecision);

                $list[$key]['nickname'] = $cryptoBot->getUser()->getNickname();
                $list[$key]['exchange'] = $cryptoBot->getProvider();
                $list[$key]['profit']   = number_format(round($trade['profit'], 2), 2, '.', '');
            }

            $tradeList = array_merge($list, $tradeList);
        }

        usort($tradeList, function (array $a, array $b) {
            $dateA = \DateTimeImmutable::createFromFormat('Y-m-d H:i:s', $a['close']);
            $dateB = \DateTimeImmutable::createFromFormat('Y-m-d H:i:s', $b['close']);

            if ($dateA->getTimestamp() === $dateB->getTimestamp()) {
                return 0;
            }

            return $dateA->getTimestamp() < $dateB->getTimestamp() ? 1 : -1;
        });

        return \array_slice($tradeList, 0, 40);
    }

    public function getSwapList(): array
    {
        $masterBot = $this->repository->find(1);

        if (!$masterBot instanceof CryptoBot) {
            return [];
        }

        try {
            return $this->cryptoBotService->getSwapList($masterBot);
        } catch (\BadMethodCallException|\RuntimeException|\LogicException|GuzzleException $exception) {
            unset($exception);

            return [];
        }
    }

    public function getSwapActionList(CryptoBot $cryptoBot): array
    {
        try {
            return $this->cryptoBotService->getSwapActionList($cryptoBot);
        } catch (\BadMethodCallException|\RuntimeException|\LogicException|GuzzleException $exception) {
            unset($exception);

            return [];
        }
    }
}
