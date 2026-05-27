<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Model;

class RSSFeedArticle
{
    public function __construct(private readonly string $url, private readonly string $title, private readonly string $shortText, private readonly array $fullText)
    {
    }

    public function getUrl(): string
    {
        return $this->url;
    }

    public function getTitle(): string
    {
        return $this->title;
    }

    public function getShortText(): string
    {
        return $this->shortText;
    }

    public function getFullText(): array
    {
        return $this->fullText;
    }
}
