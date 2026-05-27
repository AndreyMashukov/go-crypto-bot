<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Model\ExtraChargeOption;
use Bundles\CryptoBotContext\Model\ManualOrder;
use Bundles\CryptoBotContext\Model\MultiExtraCharge;
use Bundles\CryptoBotContext\Model\MultiProfitOption;
use Bundles\CryptoBotContext\Model\ProfitOption;
use Bundles\CryptoBotContext\Model\Signal;
use GuzzleHttp\ClientInterface;
use GuzzleHttp\Exception\BadResponseException;
use JMS\Serializer\ArrayTransformerInterface;
use JMS\Serializer\SerializationContext;
use Psr\Log\LoggerInterface;

class CryptoBotService extends AbstractHttpService
{
    public const OPERATION_BUY = 'buy';

    public const OPERATION_SELL = 'sell';

    private ArrayTransformerInterface $arrayTransformer;

    public function __construct(
        ClientInterface $client,
        LoggerInterface $logger,
        ArrayTransformerInterface $arrayTransformer
    ) {
        parent::__construct($client, $logger);

        $this->arrayTransformer = $arrayTransformer;
    }

    public function setupLimits(CryptoBot $cryptoBot): void
    {
        if (!$cryptoBot->getContainerId()) {
            throw new \BadMethodCallException('Container ID must be set');
        }

        if (!$cryptoBot->getPort()) {
            throw new \BadMethodCallException('Port must be set');
        }

        if (!$cryptoBot->isRunning()) {
            throw new \BadMethodCallException('Service must be running');
        }

        if (!$cryptoBot->getIpAddress()) {
            throw new \BadMethodCallException('Bot must have IP address');
        }

        $currentLimits    = $this->getCurrentLimit($cryptoBot);
        $currentLimitsMap = [];
        foreach ($currentLimits as $currentLimit) {
            $currentLimitsMap[$currentLimit['symbol']] = $currentLimit;
        }

        foreach ($cryptoBot->getCryptoTradeConfigs() as $config) {
            if (isset($currentLimitsMap[$config->getSymbol()])) {
                $this->updateLimit(
                    $cryptoBot,
                    $currentLimitsMap[$config->getSymbol()],
                    $config
                );
                unset($currentLimitsMap[$config->getSymbol()]);
            } else {
                $this->createLimit(
                    $cryptoBot,
                    [],
                    $config
                );
            }
        }

        // disable deleted limits
        foreach ($currentLimitsMap as $deletedLimit) {
            $deletedLimit['isEnabled'] = false;
            $this->updateLimit(
                $cryptoBot,
                $deletedLimit,
                null
            );
        }
    }

    public function syncConfig(CryptoBot $cryptoBot): void
    {
        $previousConfigs = [];
        foreach ($cryptoBot->getCryptoTradeConfigs() as $tradeConfig) {
            $previousConfigs[$tradeConfig->getSymbol()] = $tradeConfig;
        }

        $cryptoBot->getCryptoTradeConfigs()->clear();

        foreach ($this->getCurrentLimit($cryptoBot) as $config) {
            $cryptoConfig = new CryptoTradeConfig();
            $cryptoConfig->setSymbol($config['symbol']);
            $cryptoConfig->setUsdtLimit($config['USDTLimit']);
            $cryptoConfig->setProfitOptions($config['profitOptions']);
            $cryptoConfig->setEnabled($config['isEnabled']);
            $cryptoConfig->setMinPriceMinutesPeriod($config['minPriceMinutesPeriod']);
            $cryptoConfig->setFrameInterval($config['frameInterval']);
            $cryptoConfig->setFramePeriod($config['framePeriod']);
            $cryptoConfig->setBuyPriceHistoryCheckInterval($config['buyPriceHistoryCheckInterval']);
            $cryptoConfig->setBuyPriceHistoryCheckPeriod($config['buyPriceHistoryCheckPeriod']);
            $cryptoConfig->setExtraChargeOptions($config['extraChargeOptions']);
            $cryptoConfig->setBuyConditions($config['tradeFiltersBuy'] ?? []);
            $cryptoConfig->setSellConditions($config['tradeFiltersSell'] ?? []);
            $cryptoConfig->setAvgConditions($config['tradeFiltersExtraCharge'] ?? []);

            // todo: test it!
            $previousConfig = $previousConfigs[$cryptoConfig->getSymbol()] ?? null;
            if ($previousConfig instanceof CryptoTradeConfig) {
                $previousSignalConfig = $previousConfig->getSignalConfig();
                $cryptoConfig->setSignalTrading($previousConfig->isSignalTrading());
                $cryptoConfig->getSignalConfig()
                    ->setPercentFilter($previousSignalConfig->getPercentFilter())
                    ->setRatingFilter($previousSignalConfig->isRatingFilter())
                    ->setAvgBuyFilter($previousSignalConfig->isAvgBuyFilter())
                    ->setAvgSellFilter($previousSignalConfig->isAvgSellFilter())
                    ->setAvgBuyCorrection($previousSignalConfig->isAvgBuyCorrection())
                    ->setAvgSellCorrection($previousSignalConfig->isAvgSellCorrection())
                    ->setSellPriceCorrectionMode($previousSignalConfig->getSellPriceCorrectionMode());

                $cryptoConfig->setLabel($previousConfig->getLabel());
                $cryptoConfig->setScore($previousConfig->getScore());
            }

            $cryptoBot->addCryptoTradeConfig($cryptoConfig);
        }
    }

    public function getChart(CryptoBot $cryptoBot, string $symbol): string
    {
        return $this->requestPlain(
            'GET',
            "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/chart/list?botUuid={$cryptoBot->getUuid()}&symbol={$symbol}",
            [],
            240 // 4 minutes
        );
    }

    public function getTradeStack(CryptoBot $cryptoBot): array
    {
        return $this->request(
            'GET',
            "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/trade/stack?botUuid={$cryptoBot->getUuid()}",
            []
        );
    }

    public function switchSymbol(CryptoBot $cryptoBot, string $symbol): array
    {
        return $this->request(
            'PUT',
            "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/trade/limit/switch/{$symbol}?botUuid={$cryptoBot->getUuid()}",
            []
        );
    }

    public function stackSorting(CryptoBot $cryptoBot, string $sorting, array $bot): array
    {
        return $this->request(
            'PUT',
            "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/bot/update?botUuid={$cryptoBot->getUuid()}",
            [
                'isMasterBot'       => $bot['isMasterBot'],
                'tradeStackSorting' => $sorting,
                'isSwapEnabled'     => $bot['isSwapEnabled'],
                'swapConfig'        => $bot['swapConfig'],
            ]
        );
    }

    public function getTradeList(CryptoBot $cryptoBot): array
    {
        if (!$cryptoBot->getIpAddress()) {
            throw new \BadMethodCallException('Bot must have IP address');
        }

        return $this->request(
            'GET',
            "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/order/trade/list?botUuid={$cryptoBot->getUuid()}",
            []
        );
    }

    public function getSwapList(CryptoBot $cryptoBot): array
    {
        if (!$cryptoBot->getIpAddress()) {
            throw new \BadMethodCallException('Bot must have IP address');
        }

        return $this->request(
            'GET',
            "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/swap/list?botUuid={$cryptoBot->getUuid()}",
            []
        );
    }

    public function getSwapActionList(CryptoBot $cryptoBot): array
    {
        if (!$cryptoBot->getIpAddress()) {
            throw new \BadMethodCallException('Bot must have IP address');
        }

        return $this->request(
            'GET',
            "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/swap/action/list?botUuid={$cryptoBot->getUuid()}",
            []
        );
    }

    public function getAccount(CryptoBot $cryptoBot): array
    {
        if (!$cryptoBot->getIpAddress()) {
            throw new \BadMethodCallException('Bot must have IP address');
        }

        return $this->request(
            'GET',
            "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/account?botUuid={$cryptoBot->getUuid()}",
            []
        );
    }

    public function getPositionsList(CryptoBot $cryptoBot): array
    {
        if (!$cryptoBot->getIpAddress()) {
            throw new \BadMethodCallException('Bot must have IP address');
        }

        return $this->request(
            'GET',
            "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/order/position/list?botUuid={$cryptoBot->getUuid()}",
            []
        );
    }

    public function createLimit(CryptoBot $cryptoBot, array $currentLimit, CryptoTradeConfig $config): void
    {
        $this->request(
            'POST',
            "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/trade/limit/create?botUuid={$cryptoBot->getUuid()}",
            [
                'symbol'                       => $config->getSymbol(),
                'USDTLimit'                    => $config->getUsdtLimit(),
                'minPrice'                     => $currentLimit['minPrice'] ?? 0.001,
                'minQuantity'                  => $currentLimit['minQuantity'] ?? 0.001,
                'minNotional'                  => $currentLimit['minNotional'] ?? 0.001,
                'isEnabled'                    => $config->isEnabled(),
                'minPriceMinutesPeriod'        => $config->getMinPriceMinutesPeriod(),
                'frameInterval'                => $config->getFrameInterval(),
                'framePeriod'                  => $config->getFramePeriod(),
                'buyPriceHistoryCheckInterval' => $config->getBuyPriceHistoryCheckInterval(),
                'buyPriceHistoryCheckPeriod'   => $config->getBuyPriceHistoryCheckPeriod(),
                'profitOptions'                => $config->getProfitOptionsMapped(),
                'extraChargeOptions'           => $config->getExtraChargeOptions(),
                'tradeFiltersBuy'              => [],
                'tradeFiltersSell'             => [],
                'tradeFiltersExtraCharge'      => [],
                'sentimentLabel'               => $config->getLabel(),
                'sentimentScore'               => $config->getScore(),
            ]
        );
    }

    public function sendTradeSignal(CryptoBot $cryptoBot, Signal $signal): void
    {
        if (!$cryptoBot->getIpAddress()) {
            throw new \BadMethodCallException('Bot must have IP address');
        }

        $context = SerializationContext::create();
        $context->setGroups('Default');

        $this->request(
            'POST',
            "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/trade/signal?botUuid={$cryptoBot->getUuid()}",
            $this->arrayTransformer->toArray($signal, $context)
        );
    }

    public function manualOrder(ManualOrder $order): void
    {
        $cryptoBot = $order->getCryptoBot();

        if (!$cryptoBot->getIpAddress()) {
            throw new \BadMethodCallException('Bot must have IP address');
        }

        try {
            $this->request(
                'POST',
                "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/order?botUuid={$cryptoBot->getUuid()}",
                [
                    'botUuid'   => $cryptoBot->getUuid(),
                    'operation' => mb_strtoupper($order->operation),
                    'symbol'    => mb_strtoupper($order->symbol),
                    'price'     => $order->price,
                    'ttl'       => $order->ttl,
                ]
            );
        } catch (BadResponseException $exception) {
            $message = $exception->getResponse()->getBody()->getContents();

            throw new \BadMethodCallException($message);
        }
    }

    public function cancelManualOrder(
        CryptoBot $cryptoBot,
        string $symbol
    ): array {
        return $this->request(
            'DELETE',
            "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/order/{$symbol}?botUuid={$cryptoBot->getUuid()}",
            []
        );
    }

    public function cancelExchangeOrder(
        CryptoBot $cryptoBot,
        string $symbol,
        string $operation
    ): array {
        if (!$cryptoBot->getIpAddress()) {
            throw new \BadMethodCallException('Bot must have IP address');
        }

        return $this->request(
            'DELETE',
            "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/order/cancel/{$operation}/{$symbol}?botUuid={$cryptoBot->getUuid()}",
            []
        );
    }

    /**
     * @param MultiExtraCharge $extraCharge
     *
     * @throws \GuzzleHttp\Exception\GuzzleException
     */
    public function setMultiExtraCharge(MultiExtraCharge $extraCharge): void
    {
        if (!$extraCharge->cryptobot->getIpAddress()) {
            throw new \BadMethodCallException('Bot must have IP address');
        }

        try {
            $this->request(
                'PUT',
                "http://{$extraCharge->cryptobot->getIpAddress()}:{$extraCharge->cryptobot->getPort()}/order/extra/charge/update?botUuid={$extraCharge->cryptobot->getUuid()}",
                [
                    'orderId'            => $extraCharge->orderId,
                    'extraChargeOptions' => $extraCharge->extraChargeOptions->map(function (
                        ExtraChargeOption $chargeOption
                    ) {
                        return [
                            'index'      => $chargeOption->index,
                            'percent'    => $chargeOption->percent,
                            'amountUsdt' => $chargeOption->amountUsdt,
                        ];
                    })->toArray(),
                ]
            );
        } catch (BadResponseException $exception) {
            $message = $exception->getResponse()->getBody()->getContents();

            throw new \BadMethodCallException($message);
        }
    }

    /**
     * @param MultiProfitOption $multiProfitOption
     *
     * @throws \GuzzleHttp\Exception\GuzzleException
     */
    public function setMultiProfitOption(MultiProfitOption $multiProfitOption): void
    {
        if (!$multiProfitOption->cryptobot->getIpAddress()) {
            throw new \BadMethodCallException('Bot must have IP address');
        }

        try {
            $this->request(
                'PUT',
                "http://{$multiProfitOption->cryptobot->getIpAddress()}:{$multiProfitOption->cryptobot->getPort()}/order/profit/options/update?botUuid={$multiProfitOption->cryptobot->getUuid()}",
                [
                    'orderId'       => $multiProfitOption->orderId,
                    'profitOptions' => $multiProfitOption->profitOptions->map(function (
                        ProfitOption $chargeOption
                    ) {
                        return [
                            'index'           => $chargeOption->index,
                            'isTriggerOption' => (bool) $chargeOption->isTriggerOption,
                            'optionValue'     => $chargeOption->optionValue,
                            'optionUnit'      => $chargeOption->optionUnit,
                            'optionPercent'   => $chargeOption->optionPercent,
                        ];
                    })->toArray(),
                ]
            );
        } catch (BadResponseException $exception) {
            $message = $exception->getResponse()->getBody()->getContents();

            throw new \BadMethodCallException($message);
        }
    }

    public function updateLimit(CryptoBot $cryptoBot, array $currentLimit, ?CryptoTradeConfig $config): void
    {
        $this->request(
            'PUT',
            "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/trade/limit/update?botUuid={$cryptoBot->getUuid()}",
            [
                'symbol'                       => $config ? $config->getSymbol() : $currentLimit['symbol'],
                'USDTLimit'                    => $config ? $config->getUsdtLimit() : $currentLimit['USDTLimit'],
                'minPrice'                     => $currentLimit['minPrice'] ?? 0.001,
                'minQuantity'                  => $currentLimit['minQuantity'] ?? 0.001,
                'minNotional'                  => $currentLimit['minNotional'] ?? 0.001,
                'isEnabled'                    => $config ? $config->isEnabled() : $currentLimit['isEnabled'],
                'minPriceMinutesPeriod'        => $config ? $config->getMinPriceMinutesPeriod() : $currentLimit['minPriceMinutesPeriod'],
                'frameInterval'                => $config ? $config->getFrameInterval() : $currentLimit['frameInterval'],
                'framePeriod'                  => $config ? $config->getFramePeriod() : $currentLimit['framePeriod'],
                'buyPriceHistoryCheckInterval' => $config ? $config->getBuyPriceHistoryCheckInterval() : $currentLimit['buyPriceHistoryCheckInterval'],
                'buyPriceHistoryCheckPeriod'   => $config ? $config->getBuyPriceHistoryCheckPeriod() : $currentLimit['buyPriceHistoryCheckPeriod'],
                'profitOptions'                => $config ? $config->getProfitOptionsMapped() : ($currentLimit['profitOptions'] ?? []),
                'extraChargeOptions'           => $config ? $config->getExtraChargeOptions() : ($currentLimit['extraChargeOptions'] ?? []),
                'tradeFiltersBuy'              => $config ? $config->getBuyConditions() : ($currentLimit['tradeFiltersBuy'] ?? []),
                'tradeFiltersSell'             => $config ? $config->getSellConditions() : ($currentLimit['tradeFiltersSell'] ?? []),
                'tradeFiltersExtraCharge'      => $config ? $config->getAvgConditions() : ($currentLimit['tradeFiltersExtraCharge'] ?? []),
                'sentimentLabel'               => $config ? $config->getLabel() : ($currentLimit['sentimentLabel'] ?? []),
                'sentimentScore'               => $config ? $config->getScore() : ($currentLimit['sentimentScore'] ?? []),
            ]
        );
    }

    public function getCurrentLimit(CryptoBot $cryptoBot): ?array
    {
        if (!$cryptoBot->getIpAddress()) {
            throw new \BadMethodCallException('Bot must have IP address');
        }

        return $this->request(
            'GET',
            "http://{$cryptoBot->getIpAddress()}:{$cryptoBot->getPort()}/trade/limit/list?botUuid={$cryptoBot->getUuid()}",
            []
        );
    }

    public function updateOneTradeLimit(CryptoBot $cryptoBot, CryptoTradeConfig $config): void
    {
        $currentLimits = $this->getCurrentLimit($cryptoBot);
        foreach ($currentLimits as $currentLimit) {
            if ($currentLimit['symbol'] === $config->getSymbol()) {
                $this->updateLimit($cryptoBot, $currentLimit, $config);
            }
        }
    }
}
