<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Model;

class SentimentResult
{
    public const LABEL_BEARISH = 'BEARISH';

    public const LABEL_BULLISH = 'BULLISH';

    public const LABEL_NEUTRAL = 'NEUTRAL';

    public function __construct(private readonly RSSFeedArticle $article, private readonly string $label, private readonly float $score, private readonly array $coins)
    {
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
