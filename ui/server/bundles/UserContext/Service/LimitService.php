<?php
namespace Bundles\UserContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\UserContext\Exception\MaxPairLimitReachedException;

class LimitService
{
    /**
     * @throws MaxPairLimitReachedException
     */
    public function checkLimits(CryptoBot $cryptoBot): void
    {
        $user               = $cryptoBot->getUser();
        $tradingSymbolCount = $cryptoBot->getActiveSymbols()->count();

        if ($tradingSymbolCount > $user->getMaxPairs()) {
            throw new MaxPairLimitReachedException('You have reached max symbol count, max count is: ' . $user->getMaxPairs());
        }
    }
}
