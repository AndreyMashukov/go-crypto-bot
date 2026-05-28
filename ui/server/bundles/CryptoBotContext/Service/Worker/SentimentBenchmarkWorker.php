<?php
declare(strict_types=1);

namespace Bundles\CryptoBotContext\Service\Worker;

/**
 * Stand-alone worker payload invoked by the parallax benchmark.
 *
 * Lives outside the DI container and accepts only scalars, so it can travel
 * across the thread boundary without dragging Doctrine or Symfony state with
 * it. Each call simulates the I/O + CPU shape of a real per-symbol sentiment
 * recalculation: a sleep for the I/O round-trip, followed by a tight loop
 * that mirrors a small averaging computation.
 */
final class SentimentBenchmarkWorker
{
    /**
     * @return array{symbol: string, label: string, score: float, sample: int}
     */
    public static function work(string $symbol, int $sleepMs, int $cpuIters): array
    {
        if ($sleepMs > 0) {
            usleep($sleepMs * 1000);
        }

        $acc = 0;
        for ($i = 0; $i < $cpuIters; $i++) {
            $acc += ((int) sqrt($i + 1)) % 7;
        }

        $score = round(fmod($acc, 100) / 100.0, 2);
        $label = match (true) {
            $score > 0.66 => 'bullish',
            $score < 0.33 => 'bearish',
            default       => 'neutral',
        };

        return [
            'symbol' => $symbol,
            'label'  => $label,
            'score'  => $score,
            'sample' => $acc,
        ];
    }
}
