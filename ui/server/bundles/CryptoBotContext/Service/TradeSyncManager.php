<?php
namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Entity\Server;
use Bundles\CryptoBotContext\Entity\CryptoBot;
use Psr\Log\LoggerInterface;

class TradeSyncManager
{
    public function __construct(private readonly CryptoBotService $cryptoBotService, private readonly TradeRegistry $tradeRegistry, private readonly DeployDomain $deployDomain, private readonly LoggerInterface $logger)
    {
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
                if ($cryptoBot->getServer() instanceof Server) {
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
