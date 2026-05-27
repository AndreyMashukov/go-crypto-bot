<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\EntitySubscriber;

use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Service\CryptoBotService;
use Doctrine\Common\EventSubscriber;
use Doctrine\ORM\Event\PreUpdateEventArgs;
use Doctrine\ORM\Events;

class CryptoTradeConfigSubscriber implements EventSubscriber
{
    /**
     * The list of parameters which affect BUY price calculation logic.
     */
    public const CANCEL_EXCHANGE_BUY_ORDER_ON = [
        'usdtLimit'                            => true,
        'enabled'                              => true,
        'signalTrading'                        => true,
        'minPriceMinutesPeriod'                => true,
        'framePeriod'                          => true,
        'frameInterval'                        => true,
        'buyPriceHistoryCheckInterval'         => true,
        'buyPriceHistoryCheckPeriod'           => true,
        'buyConditions'                        => true,
        'signalConfig.ratingFilter'            => true,
        'signalConfig.avgBuyFilter'            => true,
        'signalConfig.avgSellFilter'           => true,
        'signalConfig.avgBuyCorrection'        => true,
        'signalConfig.avgSellCorrection'       => true,
        'signalConfig.percentFilter'           => true,
        'signalConfig.sellPriceCorrectionMode' => true,
    ];

    public const BOT_REMOTE_UPDATE_REQUIRED = [
        'label' => true,
        'score' => true,
    ];

    private CryptoBotService $cryptoBotService;

    public function __construct(CryptoBotService $cryptoBotService)
    {
        $this->cryptoBotService = $cryptoBotService;
    }

    public function getSubscribedEvents()
    {
        return [
            Events::preUpdate,
        ];
    }

    public function preUpdate(PreUpdateEventArgs $eventArgs): void
    {
        $entity = $eventArgs->getObject();

        if (!$entity instanceof CryptoTradeConfig) {
            return;
        }

        // todo: move this logic to service, refactor required in the future.
        $cryptoBot = $entity->getCryptobot();
        if (!$cryptoBot->isRunning()) {
            return;
        }

        $changeSet            = $eventArgs->getEntityChangeSet();
        $cancelBuy            = false;
        $remoteUpdateRequired = false;

        foreach (array_keys($changeSet) as $field) {
            if (isset(self::CANCEL_EXCHANGE_BUY_ORDER_ON[$field])) {
                $cancelBuy = true;
            }

            if (isset(self::BOT_REMOTE_UPDATE_REQUIRED[$field])) {
                $remoteUpdateRequired = true;
            }
        }

        if ($cancelBuy) {
            try {
                $this->cryptoBotService->cancelExchangeOrder(
                    $cryptoBot,
                    $entity->getSymbol(),
                    CryptoBotService::OPERATION_BUY
                );
            } catch (\Throwable $exception) {
                // It can be something wrong on server (bot) or Order could not exist on Crypto Exchange.
                unset($exception);
            }
        }

        if ($remoteUpdateRequired) {
            try {
                $this->cryptoBotService->updateOneTradeLimit($cryptoBot, $entity);
            } catch (\Throwable $exception) {
                // It can be something wrong on server (bot) or Order could not exist on Crypto Exchange.
                unset($exception);
            }
        }
    }
}
