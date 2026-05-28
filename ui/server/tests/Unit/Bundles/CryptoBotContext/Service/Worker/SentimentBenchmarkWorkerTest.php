<?php
declare(strict_types=1);

namespace App\Tests\Unit\Bundles\CryptoBotContext\Service\Worker;

use Bundles\CryptoBotContext\Service\Worker\SentimentBenchmarkWorker;
use PHPUnit\Framework\TestCase;

final class SentimentBenchmarkWorkerTest extends TestCase
{
    public function testReturnsExpectedShape(): void
    {
        $out = SentimentBenchmarkWorker::work('BTCUSDT', 0, 1000);

        self::assertSame(['symbol', 'label', 'score', 'sample'], array_keys($out));
        self::assertSame('BTCUSDT', $out['symbol']);
        self::assertIsString($out['label']);
        self::assertIsFloat($out['score']);
        self::assertIsInt($out['sample']);
    }

    public function testOutputIsDeterministicForFixedInput(): void
    {
        $a = SentimentBenchmarkWorker::work('ETHUSDT', 0, 5000);
        $b = SentimentBenchmarkWorker::work('ETHUSDT', 0, 5000);

        self::assertSame($a, $b, 'identical input must yield identical output');
    }

    public function testSymbolIsPassedThroughVerbatim(): void
    {
        foreach (['NEOUSDT', 'SOLUSDT', 'LTCUSDT'] as $symbol) {
            $out = SentimentBenchmarkWorker::work($symbol, 0, 100);
            self::assertSame($symbol, $out['symbol']);
        }
    }

    public function testLabelMatchesScoreBucket(): void
    {
        $out = SentimentBenchmarkWorker::work('BTCUSDT', 0, 12345);

        $expectedLabel = match (true) {
            $out['score'] > 0.66 => 'bullish',
            $out['score'] < 0.33 => 'bearish',
            default              => 'neutral',
        };

        self::assertSame($expectedLabel, $out['label']);
    }

    public function testZeroSleepIsHonoured(): void
    {
        $start = hrtime(true);
        SentimentBenchmarkWorker::work('BTCUSDT', 0, 100);
        $elapsedMs = (hrtime(true) - $start) / 1_000_000.0;

        self::assertLessThan(50.0, $elapsedMs, 'zero-sleep call must not stall on usleep()');
    }
}
