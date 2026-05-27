<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Repository\TradeRepository;
use JMS\Serializer\ArrayTransformerInterface;
use JMS\Serializer\SerializationContext;

class TradeStackService
{
    private HealthCheckService $healthCheckService;

    private CryptoBotService $cryptoBotService;

    private TradeRepository $tradeRepository;

    private ArrayTransformerInterface $arrayTransformer;

    private PivotService $pivotService;

    public function __construct(
        HealthCheckService $healthCheckService,
        CryptoBotService $cryptoBotService,
        TradeRepository $tradeRepository,
        ArrayTransformerInterface $arrayTransformer,
        PivotService $pivotService
    ) {
        $this->healthCheckService = $healthCheckService;
        $this->cryptoBotService   = $cryptoBotService;
        $this->tradeRepository    = $tradeRepository;
        $this->arrayTransformer   = $arrayTransformer;
        $this->pivotService       = $pivotService;
    }

    public function getTradeStack(CryptoBot $cryptoBot): array
    {
        $stack        = $this->cryptoBotService->getTradeStack($cryptoBot);
        $healthCheck  = $this->healthCheckService->healthCheck($cryptoBot);
        $bestTradeMap = $this->tradeRepository->getBestMonthSymbols();

        $configMap = [];
        foreach ($cryptoBot->getCryptoTradeConfigs() as $config) {
            $configMap[$config->getSymbol()] = $config;
        }

        $pivotGrid = $this->pivotService->getPivotGrid($cryptoBot);

        foreach ($stack as &$stackItem) {
            $config = $configMap[$stackItem['symbol']] ?? null;
            $pivots = $pivotGrid[$stackItem['symbol']] ?? [];

            $sentiment = null;

            if ($config instanceof CryptoTradeConfig) {
                $serializerContext = SerializationContext::create();
                $serializerContext->setGroups(['cryptotrade_config']);
                $stackItem['signalTrading'] = $config->isSignalTrading();
                $stackItem['signalConfig']  = $this->arrayTransformer->toArray($config->getSignalConfig(), $serializerContext);

                if ($config->getLabel()) {
                    $sentiment = [
                        'label' => $config->getLabel(),
                        'score' => $config->getScore(),
                    ];
                }
            } else {
                $stackItem['signalTrading'] = false;
                $stackItem['signalConfig']  = null;
            }

            $stackItem['sentiment'] = $sentiment;
            $stackItem['pivots']    = $pivots;
            $stackItem['rating']    = $bestTradeMap[$stackItem['symbol']]     ?? null;
        }

        return [
            'restartRequired'             => $cryptoBot->isRestartRequired(),
            'stack'                       => $stack,
            'updates'                     => $healthCheck['updates'] ?? [],
            'sorting'                     => $healthCheck['bot']['tradeStackSorting'] ?? 'percent',
            'hasActiveSignalSubscription' => $cryptoBot->hasActiveSignalSubscription(),
        ];
    }

    public function switchSymbol(CryptoBot $cryptoBot, string $symbol): array
    {
        return $this->cryptoBotService->switchSymbol($cryptoBot, $symbol);
    }

    public function setSorting(CryptoBot $cryptoBot, string $sorting): array
    {
        $healthCheck = $this->healthCheckService->healthCheck($cryptoBot);

        return $this->cryptoBotService->stackSorting($cryptoBot, $sorting, $healthCheck['bot']);
    }
}
