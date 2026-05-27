<?php
namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Entity\Embedded\SignalConfig;
use Bundles\CryptoBotContext\Model\Signal;
use Bundles\CryptoBotContext\Model\SignalProfitOption;
use Bundles\CryptoBotContext\Repository\TradeRepository;

class SignalModification
{
    public function __construct(private readonly TradeRepository $tradeRepository)
    {
    }

    public function modify(Signal $signal, CryptoTradeConfig $config): Signal
    {
        $signalConfig = $config->getSignalConfig();
        $ratingMap    = $this->tradeRepository->getBestMonthSymbols();
        $rating       = $ratingMap[$signal->getSymbol()] ?? null;

        if (!$rating) {
            return $signal;
        }

        $avgBuyPrice     = $rating['avgBuyPrice'];
        $buyPriceChanged = false;

        if ($signalConfig->isAvgBuyCorrection()) {
            $signal->setBuyPrice(min($avgBuyPrice, $signal->getBuyPrice()));
            $buyPriceChanged = true;
        }

        if ($signalConfig->isAvgSellCorrection()) {
            $avgSellPrice = match ($signalConfig->getSellPriceCorrectionMode()) {
                SignalConfig::SELL_PRICE_CORRECTION_MODE_EQUAL => $rating['avgSellPrice'],
                SignalConfig::SELL_PRICE_CORRECTION_MODE_MAX => max($rating['avgSellPrice'], $signal->getMaxSellPrice()),
                SignalConfig::SELL_PRICE_CORRECTION_MODE_MIN => min($rating['avgSellPrice'], $signal->getMaxSellPrice()),
                default => throw new \BadMethodCallException(''),
            };
            $avgPositionTimeHours = (float) max($rating['avgPositionTimeHours'], 1);
            $newProfitOptions     = [];
            $percent       = round(($avgSellPrice * 100 / $signal->getBuyPrice()) - 100, 2);
            $primaryOption = SignalProfitOption::create(
                $avgSellPrice,
                0,
                $avgPositionTimeHours,
                'h',
                $percent,
                true
            );
            $newProfitOptions[] = $primaryOption;
            $minSignalSellPrice = $signal->getMinSellPrice();
            $minSignalPercent   = round(($minSignalSellPrice * 100 / $signal->getBuyPrice()) - 100, 2);
            $lastPercent = $minSignalPercent;
            $lastOption  = SignalProfitOption::create(
                $minSignalSellPrice,
                2,
                $avgPositionTimeHours * 2,
                'h',
                $lastPercent,
                false
            );
            $middleSellPrice = ($primaryOption->getSellPrice() + $lastOption->getSellPrice()) / 2;
            $middlePercent   = round(($middleSellPrice * 100 / $signal->getBuyPrice()) - 100, 2);
            $middleOption    = SignalProfitOption::create(
                $middleSellPrice,
                1,
                ceil(($primaryOption->optionValue + $lastOption->optionValue) / 2),
                'h',
                $middlePercent,
                false
            );
            $newProfitOptions[] = $middleOption;
            $newProfitOptions[] = $lastOption;
            $signal->setProfitOptions($newProfitOptions);
        }

        if ($buyPriceChanged) {
            foreach ($signal->getProfitOptions() as $profitOption) {
                $percent = round(($profitOption->getSellPrice() * 100 / $signal->getBuyPrice()) - 100, 2);

                $profitOption->optionPercent = $percent;
            }
        }

        $primaryOption = $signal->getProfitOptions()[0];
        $percent       = round(($primaryOption->getSellPrice() * 100 / $signal->getBuyPrice()) - 100, 2);
        $signal->setPercent($percent);

        return $signal;
    }
}
