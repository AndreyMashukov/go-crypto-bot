<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Exception\BadTradeSignalException;
use Bundles\CryptoBotContext\Model\Signal;
use Bundles\CryptoBotContext\Model\SignalProfitOption;
use Bundles\CryptoBotContext\Repository\TradeRepository;

class SignalFilter
{
    private TradeRepository $tradeRepository;

    public function __construct(TradeRepository $tradeRepository)
    {
        $this->tradeRepository = $tradeRepository;
    }

    /**
     * @param Signal            $signal
     * @param CryptoTradeConfig $config
     *
     * @throws BadTradeSignalException
     *
     * @return Signal
     */
    public function process(Signal $signal, CryptoTradeConfig $config): void
    {
        $signalConfig      = $config->getSignalConfig();
        $profitOptions     = $signal->getProfitOptions();
        $lastProfitPercent = end($profitOptions);

        if (!$lastProfitPercent instanceof SignalProfitOption) {
            throw new BadTradeSignalException('At least one profit options should be defined.');
        }

        if ($lastProfitPercent->optionPercent < $signalConfig->getPercentFilter()) {
            throw new BadTradeSignalException('Required percent is not reached.');
        }

        $ratingMap = $this->tradeRepository->getBestMonthSymbols();
        $rating    = $ratingMap[$signal->getSymbol()] ?? null;

        if (!$rating && $signalConfig->isRatingFilter()) {
            throw new BadTradeSignalException('Required rating is not found.');
        }

        if (!$rating) {
            return;
        }

        $avgBuyPrice  = $rating['avgBuyPrice'];

        if ($signalConfig->isAvgBuyFilter() && $signal->getBuyPrice() > $avgBuyPrice) {
            throw new BadTradeSignalException('Average buy price condition is not reached.');
        }

        $avgSellPrice = $rating['avgSellPrice'];

        if ($signalConfig->isAvgSellFilter()) {
            $matched = false;

            /** @var SignalProfitOption $option */
            foreach ($signal->getProfitOptions() as $option) {
                if ($option->getSellPrice() <= $avgSellPrice) {
                    $matched = true;
                    break;
                }
            }

            if (!$matched) {
                throw new BadTradeSignalException('Average sell price condition is not reached.');
            }
        }
    }
}
