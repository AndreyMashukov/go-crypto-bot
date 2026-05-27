<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Unit\Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Entity\Embedded\SignalConfig;
use Bundles\CryptoBotContext\Model\Signal;
use Bundles\CryptoBotContext\Model\SignalProfitOption;
use Bundles\CryptoBotContext\Repository\TradeRepository;
use Bundles\CryptoBotContext\Service\SignalModification;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;

/**
 * @group unit
 */
class SignalModificationTest extends TestCase
{
    private SignalModification $signalModification;

    /** @var MockObject|TradeRepository */
    private TradeRepository $tradeRepository;

    protected function setUp(): void
    {
        $this->tradeRepository    = $this->createMock(TradeRepository::class);
        $this->signalModification = new SignalModification($this->tradeRepository);
    }

    /**
     * Should allow to modify signal.
     *
     * @dataProvider signalDataProvider
     *
     * @param float $minSellPrice
     * @param array $expects
     */
    public function testShouldAllowToModifySignal(float $minSellPrice, array $expects): void
    {
        $signal = $this->createMock(Signal::class);
        $config = $this->createMock(CryptoTradeConfig::class);

        $signalConfig = $this->createMock(SignalConfig::class);
        $config
            ->expects($this->once())
            ->method('getSignalConfig')
            ->willReturn($signalConfig);

        $signal->expects($this->once())
            ->method('getSymbol')
            ->willReturn('PEPEUSDT');

        $signalConfig
            ->expects($this->once())
            ->method('getSellPriceCorrectionMode')
            ->willReturn(SignalConfig::SELL_PRICE_CORRECTION_MODE_EQUAL);

        $signal->expects($this->once())
            ->method('getMinSellPrice')
            ->willReturn($minSellPrice);

        $setProfitOptions = [];

        $signal->expects($this->once())
            ->method('setProfitOptions')
            ->willReturnCallback(function (array $profitOptions) use (&$setProfitOptions, $signal) {
                $setProfitOptions = $profitOptions;

                return $signal;
            });

        $signal
            ->method('getProfitOptions')
            ->willReturnCallback(function () use (&$setProfitOptions) {
                return $setProfitOptions;
            });

        $buyPrice = 0.000015396;
        $signal->expects($this->once())
            ->method('setBuyPrice')
            ->willReturnCallback(function (float $value) use ($signal, &$buyPrice) {
                $buyPrice = $value;

                return $signal;
            });

        $signal
            ->method('getBuyPrice')
            ->willReturnCallback(function () use (&$buyPrice) {
                return $buyPrice;
            });

        $signal->expects($this->once())
            ->method('setPercent')
            ->with(2.5);

        $signalConfig->expects($this->once())
            ->method('isAvgBuyCorrection')
            ->willReturn(true);

        $signalConfig->expects($this->once())
            ->method('isAvgSellCorrection')
            ->willReturn(true);

        $this->tradeRepository->expects($this->once())
            ->method('getBestMonthSymbols')
            ->willReturn([
               'PEPEUSDT' => [
                   'avgBuyPrice'          => 0.000015196746801276445,
                   'avgSellPrice'         => 0.000015576333333333335,
                   'percent'              => 2.5,
                   'avgPositionTimeHours' => 13,
                   'totalProfit'          => 51.72,
                   'trades'               => 9,
                   'profitPerTrade'       => 5.75,
                   'symbol'               => 'PEPEUSDT',
               ],
            ]);

        $signalNew = $this->signalModification->modify($signal, $config);
        $this->assertEquals($signal, $signalNew);
        $this->assertCount(3, $setProfitOptions);
        /** @var SignalProfitOption $firstProfit */
        $firstProfit = $setProfitOptions[0];
        $this->assertEquals($expects[0][0], $firstProfit->getSellPrice());
        $this->assertEquals($expects[0][1], $firstProfit->index);
        $this->assertEquals($expects[0][2], $firstProfit->optionValue);
        $this->assertEquals($expects[0][3], $firstProfit->optionPercent);
        $this->assertEquals($expects[0][4], $firstProfit->optionUnit);
        $this->assertEquals($expects[0][5], $firstProfit->isTriggerOption);

        /** @var SignalProfitOption $secondProfit */
        $secondProfit = $setProfitOptions[1];
        $this->assertEquals($expects[1][0], $secondProfit->getSellPrice());
        $this->assertEquals($expects[1][1], $secondProfit->index);
        $this->assertEquals($expects[1][2], $secondProfit->optionValue);
        $this->assertEquals($expects[1][3], $secondProfit->optionPercent);
        $this->assertEquals($expects[1][4], $secondProfit->optionUnit);
        $this->assertEquals($expects[1][5], $secondProfit->isTriggerOption);

        /** @var SignalProfitOption $secondProfit */
        $secondProfit = $setProfitOptions[2];
        $this->assertEquals($expects[2][0], $secondProfit->getSellPrice());
        $this->assertEquals($expects[2][1], $secondProfit->index);
        $this->assertEquals($expects[2][2], $secondProfit->optionValue);
        $this->assertEquals($expects[2][3], $secondProfit->optionPercent);
        $this->assertEquals($expects[2][4], $secondProfit->optionUnit);
        $this->assertEquals($expects[2][5], $secondProfit->isTriggerOption);

        $this->assertEquals(0.000015196746801276445, $buyPrice);
    }

    public function signalDataProvider(): \Generator
    {
        yield 'test #1 min percent: 1.1 (0.000015363911016090486)' => [
            'minSellPrice' => 0.000015363911016090486,
            'expects'      => [
                [
                    0.000015576333333333335,
                    0,
                    13,
                    2.5,
                    'h',
                    true,
                ],
                [
                    0.00001547012217471191,
                    1,
                    20,
                    1.8,
                    'h',
                    false,
                ],
                [
                    0.000015363911016090486,
                    2,
                    26,
                    1.1,
                    'h',
                    false,
                ],
            ],
        ];

        yield 'test #1 min percent: 0.9 (0.00001533351752248793)' => [
            'minSellPrice' => 0.00001533351752248793,
            'expects'      => [
                [
                    0.000015576333333333335,
                    0,
                    13,
                    2.5,
                    'h',
                    true,
                ],
                [
                    0.00001545492542791063,
                    1,
                    20,
                    1.7,
                    'h',
                    false,
                ],
                [
                    0.00001533351752248793,
                    2,
                    26,
                    0.9,
                    'h',
                    false,
                ],
            ],
        ];
    }
}
