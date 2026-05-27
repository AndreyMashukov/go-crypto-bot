<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Model;

class RSSFeedArticle
{
    private string $url;

    private string $title;

    private string $shortText;

    private array $fullText;

    public function __construct(string $url, string $title, string $shortText, array $fullText)
    {
        $this->url       = $url;
        $this->title     = $title;
        $this->shortText = $shortText;
        $this->fullText  = $fullText;
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
