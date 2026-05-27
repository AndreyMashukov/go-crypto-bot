<?php
namespace Bundles\CryptoBotContext\EntitySubscriber;

use Doctrine\Bundle\DoctrineBundle\Attribute\AsDoctrineListener;
use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Service\CryptoBotService;
use Doctrine\ORM\Event\PreUpdateEventArgs;
use Doctrine\ORM\Events;

#[AsDoctrineListener(event: Events::preUpdate)]
class CryptoTradeConfigSubscriber
{
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
    public function __construct(private readonly CryptoBotService $cryptoBotService)
    {
    }
    public function preUpdate(PreUpdateEventArgs $eventArgs): void
    {
        $entity = $eventArgs->getObject();

        if (!$entity instanceof CryptoTradeConfig) {
            return;
        }

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
                unset($exception);
            }
        }

        if ($remoteUpdateRequired) {
            try {
                $this->cryptoBotService->updateOneTradeLimit($cryptoBot, $entity);
            } catch (\Throwable $exception) {
                unset($exception);
            }
        }
    }
}
