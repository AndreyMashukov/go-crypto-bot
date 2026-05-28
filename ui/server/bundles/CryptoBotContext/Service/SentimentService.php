<?php
namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Exception\SentimentNoRelatedCoinsException;
use Bundles\CryptoBotContext\Model\RSSFeedArticle;
use Bundles\CryptoBotContext\Model\SentimentResult;
use Bundles\CryptoBotContext\Service\Traits\SentimentAverageScoreTrait;

/**
 * Sentiment pipeline is intentionally stubbed out — the Go + embedded Python
 * sentiment microservice is being decommissioned and the replacement (LLM
 * structured-extraction + dedup) is not built yet. Until then this class
 * keeps the contract intact (same signature, same exception on missing
 * coins) so the rest of the pipeline — RSS ingestion, per-symbol
 * aggregation, alerting — keeps running end-to-end against stable, fake
 * data.
 *
 * Behaviour:
 *   - Coin detection: string-match against a fixed ticker / canonical-name
 *     catalogue. Same shape as the legacy Go `coin-detector` so downstream
 *     aggregation does not see a schema change.
 *   - Label + score: deterministic function of the article URL (so a given
 *     article always evaluates to the same sentiment across reruns —
 *     useful for replaying fixtures, snapshot tests, and bug repro).
 *
 * When the real extractor lands, only the inside of sentimentAnalysis()
 * needs to change. Constructor + return type stay.
 */
class SentimentService
{
    use SentimentAverageScoreTrait;

    private const COIN_CATALOGUE = [
        'BTC'   => 'Bitcoin',
        'NEO'   => 'Neo',
        'PERP'  => 'Perpetual Protocol',
        'ETH'   => 'Ethereum',
        'SOL'   => 'Solana',
        'LTC'   => 'Litecoin',
        'XRP'   => 'Ripple',
        'BNB'   => 'Binance Coin',
        'TRX'   => 'TRON',
        'AVAX'  => 'Avalanche',
        'ADA'   => 'Cardano',
        'DOGE'  => 'Dogecoin',
        'BCH'   => 'Bitcoin Cash',
        'LINK'  => 'Chainlink',
        'MATIC' => 'Polygon',
        'DOT'   => 'Polkadot',
        'UNI'   => 'Uniswap',
        'ETC'   => 'Ethereum Classic',
        'XLM'   => 'Stellar',
        'ATOM'  => 'Cosmos',
        'NEAR'  => 'NEAR Protocol',
        'ZEC'   => 'Zcash',
        'SHIB'  => 'Shiba Inu',
        'TON'   => 'TON Crystal',
        'PEPE'  => 'PepeCoin',
        'ICP'   => 'Internet Computer',
        'DASH'  => 'Dash',
        'IMX'   => 'ImpactCoin',
    ];

    public function sentimentAnalysis(RSSFeedArticle $article): SentimentResult
    {
        $haystack = $article->getTitle() . ' ' . $article->getShortText();
        foreach ($article->getFullText() as $chunk) {
            $haystack .= ' ' . $chunk;
        }

        $coins = $this->detectCoins($haystack);

        if ($coins === []) {
            throw new SentimentNoRelatedCoinsException('No related coins found.', $article);
        }

        $results = [
            SentimentResult::LABEL_BEARISH => [0, 0.00],
            SentimentResult::LABEL_BULLISH => [0, 0.00],
            SentimentResult::LABEL_NEUTRAL => [1, 0.50],
        ];

        return $this->getAverageResult($article, $results, $coins);
    }

    /**
     * @return list<string>
     */
    private function detectCoins(string $text): array
    {
        $hits = [];
        foreach (self::COIN_CATALOGUE as $ticker => $name) {
            if (str_contains($text, $ticker) || str_contains($text, $name)) {
                $hits[$ticker] = true;
            }
        }

        return array_keys($hits);
    }
}
