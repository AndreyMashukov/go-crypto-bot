<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Tests\Unit\Service;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Repository\CryptoBotRepository;
use Bundles\CryptoBotContext\Repository\TradeRepository;
use Bundles\CryptoBotContext\Service\CryptoBotService;
use Bundles\CryptoBotContext\Service\PivotService;
use Bundles\CryptoBotContext\Service\TradeListService;
use Bundles\UserContext\Entity\User;
use PHPUnit\Framework\TestCase;

/**
 * @group unit
 */
class TradeListServiceTest extends TestCase
{
    /**
     * Should allow to shuffle list of trades.
     */
    public function testShouldAllowToShuffleListOfTrades(): void
    {
        $repository       = $this->createMock(CryptoBotRepository::class);
        $cryptoBotService = $this->createMock(CryptoBotService::class);
        $tradeRepository  = $this->createMock(TradeRepository::class);

        $userA = $this->createMock(User::class);
        $userB = $this->createMock(User::class);
        $userA->method('getNickname')->willReturn('test1');
        $userB->method('getNickname')->willReturn('test2');

        $cryptobotA = $this->createMock(CryptoBot::class);
        $cryptobotB = $this->createMock(CryptoBot::class);
        $cryptobotA->method('getUser')->willReturn($userA);
        $cryptobotB->method('getUser')->willReturn($userB);
        $cryptobotA->method('getProvider')->willReturn('binance');
        $cryptobotB->method('getProvider')->willReturn('bybit');
        $repository
            ->method('findBy')
            ->willReturn([
                $cryptobotA,
                $cryptobotB,
            ]);

        $bullets   = [];
        $bullets[] = json_decode(file_get_contents(__DIR__ . '/../../_data/trades_1.json'), true);
        $bullets[] = json_decode(file_get_contents(__DIR__ . '/../../_data/trades_2.json'), true);

        $cryptoBotService
            ->expects($this->exactly(2))
            ->method('getTradeList')
            ->willReturnCallback(function () use ($bullets) {
                return array_shift($bullets);
            });

        $pivotService = $this->createMock(PivotService::class);

        $listService = new TradeListService($repository, $cryptoBotService, $tradeRepository, $pivotService);
        $trades      = $listService->getPublicTradeList();
        // file_put_contents(__DIR__ . '/../../_data/Unit/TradeListServiceTest/testShouldAllowToShuffleListOfTrades.json', json_encode($trades));
        $this->assertJsonStringEqualsJsonFile(__DIR__ . '/../../_data/Unit/TradeListServiceTest/testShouldAllowToShuffleListOfTrades.json', json_encode($trades));
    }
}
