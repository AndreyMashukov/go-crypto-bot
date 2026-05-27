<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Psr\Log\LoggerInterface;

class TradeSyncManager
{
    private CryptoBotService $cryptoBotService;

    private TradeRegistry $tradeRegistry;

    private DeployDomain $deployDomain;

    private LoggerInterface $logger;

    public function __construct(
        CryptoBotService $cryptoBotService,
        TradeRegistry $tradeRegistry,
        DeployDomain $deployDomain,
        LoggerInterface $logger
    ) {
        $this->cryptoBotService = $cryptoBotService;
        $this->tradeRegistry    = $tradeRegistry;
        $this->deployDomain     = $deployDomain;
        $this->logger           = $logger;
    }

    public function syncTrades(CryptoBot $cryptoBot, ?callable $onError, bool $allowStop = false): void
    {
        if (!$cryptoBot->getIpAddress()) {
            return;
        }

        $stop = false;
        $user = $cryptoBot->getUser();

        try {
            $trades = $this->cryptoBotService->getTradeList($cryptoBot);
        } catch (\Throwable $throwable) {
            $this->logger->error($throwable->getMessage(), [
                'line' => $throwable->getLine(),
                'file' => $throwable->getFile(),
            ]);

            return;
        }

        foreach ($trades as $trade) {
            $this->tradeRegistry->registerTrade($trade, $cryptoBot);

            if ($user->getBudget() <= 0.00 && !$user->hasActiveBasicSubscription() && $allowStop) {
                $stop = true;
            }
        }

        if ($stop && !$cryptoBot->isStopped()) {
            try {
                if ($cryptoBot->getServer()) {
                    $this->deployDomain->doStop(
                        $cryptoBot,
                        $cryptoBot->getServer(),
                        'Please recharge budget or buy subscription.',
                        true
                    );
                }
            } catch (\Throwable $throwable) {
                if ($onError) {
                    $onError($throwable);
                }
            }
        }
    }
}
