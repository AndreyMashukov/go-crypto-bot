<?php
namespace Bundles\CryptoBotContext\Service\Traits;

use Bundles\CryptoBotContext\Model\RSSFeedArticle;
use Bundles\CryptoBotContext\Model\SentimentResult;

trait SentimentAverageScoreTrait
{
    protected function calculateAverage(array $results): array
    {
        $average = [
            SentimentResult::LABEL_BEARISH => 0.00,
            SentimentResult::LABEL_BULLISH => 0.00,
            SentimentResult::LABEL_NEUTRAL => 0.00,
        ];

        foreach ($results as $label => $result) {
            $average[$label] = round($result[1] / ($result[0] ?: 1), 2);
        }

        return $average;
    }

    protected function processAverage(array $average): array
    {
        return match (true) {
            $average[SentimentResult::LABEL_BEARISH] > $average[SentimentResult::LABEL_BULLISH] => [SentimentResult::LABEL_BEARISH, $average[SentimentResult::LABEL_BEARISH] - $average[SentimentResult::LABEL_BULLISH]],
            $average[SentimentResult::LABEL_BULLISH] > $average[SentimentResult::LABEL_BEARISH] => [SentimentResult::LABEL_BULLISH, $average[SentimentResult::LABEL_BULLISH] - $average[SentimentResult::LABEL_BEARISH]],
            default => [SentimentResult::LABEL_NEUTRAL, $average[SentimentResult::LABEL_NEUTRAL]],
        };
    }

    protected function getAverageResult(RSSFeedArticle $article, array $results, array $relatedCoins): SentimentResult
    {
        $average      = $this->calculateAverage($results);
        $relatedCoins = array_unique($relatedCoins);

        [$label, $score] = $this->processAverage($average);

        return new SentimentResult($article, $label, $score, $relatedCoins);
    }
}
