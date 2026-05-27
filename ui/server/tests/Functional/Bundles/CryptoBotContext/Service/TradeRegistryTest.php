<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Functional\Bundles\CryptoBotContext\Service;

use App\Tests\RestTestCase;
use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Repository\TradeRepository;
use Bundles\CryptoBotContext\Service\TradeRegistry;
use Bundles\OxaPayContext\Entity\Transaction;
use Bundles\UserContext\Entity\User;
use PHPUnit\Framework\MockObject\MockObject;

/**
 * @group functional
 */
class TradeRegistryTest extends RestTestCase
{
    /** @var MockObject|TradeRepository */
    private TradeRepository $tradeRepository;

    protected function services(): void
    {
        parent::services();

        $this->tradeRepository = $this->createMock(TradeRepository::class);
        self::getContainer()->set('test.trade_repository', $this->tradeRepository);
    }

    /**
     * Should apply dynamic commission percent.
     *
     * @dataProvider profitDataProvider
     *
     * @param float $profitUsdt
     * @param float $tradingMonthlyVolume
     * @param float $initialBudget
     * @param float $expectedBudget
     */
    public function testShouldApplyDynamicCommissionPercent(
        float $profitUsdt,
        float $tradingMonthlyVolume,
        float $initialBudget,
        float $expectedBudget
    ): void {
        $tradeData = [
            'profit'       => $profitUsdt,
            'orderId'      => 999,
            'open'         => '2024-12-01 00:00:00',
            'close'        => '2024-12-01 01:00:00',
            'buyQuantity'  => 1,
            'sellQuantity' => 1,
            'buy'          => 100,
            'sell'         => 110,
            'symbol'       => 'BTCUSDT',
            'percent'      => 10.00,
        ];

        $this->tradeRepository
            ->expects($this->once())
            ->method('findOneBy')
            ->willReturn(null);

        $user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $this->assertInstanceOf(User::class, $user);
        $cryptoBot = $user->getBinanceBot();
        $this->assertInstanceOf(CryptoBot::class, $cryptoBot);
        $user->setBudget($initialBudget);
        $this->em->flush();

        $transactions = $user->getTransactions();
        $this->assertEquals(0, $transactions->count());

        $this->tradeRepository
            ->expects($this->once())
            ->method('getProfitByPeriod')
            ->with($cryptoBot, 'month')
            ->willReturn([
                [
                    'title'       => (new \DateTimeImmutable('now'))->format('Y-m'),
                    'tradeVolume' => $tradingMonthlyVolume,
                ],
            ]);

        /** @var TradeRegistry $tradeRegistry */
        $tradeRegistry = self::getContainer()->get('test.trade_registry');
        $tradeRegistry->registerTrade($tradeData, $cryptoBot);
        $this->em->refresh($user);
        $this->assertEquals($expectedBudget, $user->getBudget());

        $transactions = $user->getTransactions();
        $this->assertEquals(1, $transactions->count());

        /** @var Transaction $transaction */
        $transaction = $transactions->first();
        $this->assertInstanceOf(Transaction::class, $transaction);
        $this->assertEquals($initialBudget - $expectedBudget, $transaction->getAmount());
    }

    public function profitDataProvider(): \Generator
    {
        yield 'test #1: level 1' => [
            'profitUsdt'           => 100.00,
            'tradingMonthlyVolume' => 5000.00,
            'initialBudget'        => 100,
            'expectedBudget'       => 50,
        ];

        yield 'test #2: level 2' => [
            'profitUsdt'           => 100.00,
            'tradingMonthlyVolume' => 25000.00,
            'initialBudget'        => 100,
            'expectedBudget'       => 65,
        ];

        yield 'test #3: level 3' => [
            'profitUsdt'           => 100.00,
            'tradingMonthlyVolume' => 50000.00,
            'initialBudget'        => 100,
            'expectedBudget'       => 70,
        ];

        yield 'test #4: level 4' => [
            'profitUsdt'           => 100.00,
            'tradingMonthlyVolume' => 50000.01,
            'initialBudget'        => 100,
            'expectedBudget'       => 75,
        ];

        yield 'test #5: level 4 (Min profit 1$)' => [
            'profitUsdt'           => 0.50,
            'tradingMonthlyVolume' => 50000.01,
            'initialBudget'        => 100,
            'expectedBudget'       => 99,
        ];
    }
}
