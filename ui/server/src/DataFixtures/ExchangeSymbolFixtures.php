<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\DataFixtures;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\ExchangeSymbol;
use Doctrine\Bundle\FixturesBundle\Fixture;
use Doctrine\Bundle\FixturesBundle\FixtureGroupInterface;
use Doctrine\Persistence\ObjectManager;

class ExchangeSymbolFixtures extends Fixture implements FixtureGroupInterface
{
    public const SYMBOL_LIST = [
        'NEOUSDT',
        'PERPUSDT',
        'ETHUSDT',
        'SOLUSDT',
        'BTCUSDT',
        'LTCUSDT',
        'XRPUSDT',
        'BNBUSDT',
        'TRXUSDT',
        'AVAXUSDT',
        'ADAUSDT',
        'DOGEUSDT',
        'BCHUSDT',
        'LINKUSDT',
        'MATICUSDT',
        'DOTUSDT',
        'UNIUSDT',
        'ETCUSDT',
        'XLMUSDT',
        'ATOMUSDT',
        'NEARUSDT',
        'ZECUSDT',
        'SHIBUSDT',
    ];

    public function load(ObjectManager $manager)
    {
        foreach (self::SYMBOL_LIST as $symbol) {
            foreach ([CryptoBot::PROVIDER_BINANCE, CryptoBot::PROVIDER_BYBIT] as $exchange) {
                $isEnabled = 1;
                if ('ZECUSDT' === $symbol && CryptoBot::PROVIDER_BYBIT === $exchange) {
                    $isEnabled = 0;
                }
                $symbolExchange = new ExchangeSymbol();
                $symbolExchange->setSymbol($symbol)
                    ->setExchange($exchange)
                    ->setEnabled($isEnabled);
                $manager->persist($symbolExchange);
            }
        }

        $manager->flush();
    }

    public static function getGroups(): array
    {
        return ['first'];
    }
}
