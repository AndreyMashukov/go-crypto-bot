<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Model;

class SentimentResult
{
    public const LABEL_BEARISH = 'BEARISH';

    public const LABEL_BULLISH = 'BULLISH';

    public const LABEL_NEUTRAL = 'NEUTRAL';

    private RSSFeedArticle $article;

    private string $label;

    private float $score;

    private array $coins;

    public function __construct(RSSFeedArticle $article, string $label, float $score, array $coins)
    {
        $this->article = $article;
        $this->label   = $label;
        $this->score   = $score;
        $this->coins   = $coins;
    }

    public function getArticle(): RSSFeedArticle
    {
        return $this->article;
    }

    public function getLabel(): string
    {
        return $this->label;
    }

    public function getScore(): float
    {
        return $this->score;
    }

    public function getCoins(): array
    {
        return $this->coins;
    }
}
