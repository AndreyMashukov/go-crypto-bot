<?php
namespace Bundles\CryptoBotContext\Exception;

use Bundles\CryptoBotContext\Model\RSSFeedArticle;

class SentimentNoRelatedCoinsException extends \Exception
{
    public function __construct($message, private readonly RSSFeedArticle $article)
    {
        parent::__construct($message);
    }

    public function getArticle(): RSSFeedArticle
    {
        return $this->article;
    }
}
