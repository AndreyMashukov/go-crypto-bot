<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\OxaPayContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Repository\TradeRepository;
use Bundles\OxaPayContext\Model\Commission;

class CommissionService
{
    public const COMMISSION_CONFIGURATION_LIST = [
        [
            'level'             => 1,
            'volume'            => 5000.00,
            'commissionPercent' => 50.00,
            'minCommissionUsd'  => 1.00,
            'title'             => '0.00<small>USDT</small> - 5000.00<small>USDT</small>',
        ],
        [
            'level'             => 2,
            'volume'            => 25000.00,
            'commissionPercent' => 35.00,
            'minCommissionUsd'  => 1.00,
            'title'             => '5000.00<small>USDT</small> - 25000.00<small>USDT</small>',
        ],
        [
            'level'             => 3,
            'volume'            => 50000.00,
            'commissionPercent' => 30.00,
            'minCommissionUsd'  => 1.00,
            'title'             => '25000.00<small>USDT</small> - 50000.00<small>USDT</small>',
        ],
        [
            'level'             => 4,
            'volume'            => PHP_FLOAT_MAX,
            'commissionPercent' => 25.00,
            'minCommissionUsd'  => 1.00,
            'title'             => '> 50000.00<small>USDT</small>',
        ],
    ];

    private TradeRepository $tradeRepository;

    public function __construct(TradeRepository $tradeRepository)
    {
        $this->tradeRepository = $tradeRepository;
    }

    public function getCommission(CryptoBot $cryptoBot): Commission
    {
        $user = $cryptoBot->getUser();

        $profitReport  = $this->tradeRepository->getProfitByPeriod($cryptoBot, 'month');
        $monthlyVolume = 0.00;

        if ($user->hasActiveBasicSubscription()) {
            return new Commission(
                0.00,
                0,
                0,
                $monthlyVolume
            );
        }

        if (isset($profitReport[0])) {
            $title = $profitReport[0]['title'];
            if ($title === (new \DateTimeImmutable('now'))->format('Y-m')) {
                $monthlyVolume = $profitReport[0]['tradeVolume'];
            }
        }

        $configurationList = $this->getCommissionConfigurationList();

        foreach ($configurationList as $configuration) {
            if ($configuration['volume'] >= $monthlyVolume) {
                return new Commission(
                    $configuration['commissionPercent'],
                    $configuration['minCommissionUsd'],
                    $configuration['level'],
                    $monthlyVolume
                );
            }
        }

        return new Commission(
            $configurationList[0]['commissionPercent'],
            $configurationList[0]['minCommissionUsd'],
            $configurationList[0]['level'],
            $monthlyVolume
        );
    }

    public function getCommissionConfigurationList(): array
    {
        return self::COMMISSION_CONFIGURATION_LIST;
    }
}
