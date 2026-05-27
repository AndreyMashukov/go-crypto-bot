<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\UserContext\Exception\MaxPairLimitReachedException;

class LimitService
{
    /**
     * @param CryptoBot $cryptoBot
     *
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
