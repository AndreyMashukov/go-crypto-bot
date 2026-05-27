<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Exception;

use Bundles\CryptoBotContext\Model\RSSFeedArticle;

class SentimentNoRelatedCoinsException extends \Exception
{
    private RSSFeedArticle $article;

    public function __construct($message, RSSFeedArticle $article)
    {
        parent::__construct($message);

        $this->article = $article;
    }

    public function getArticle(): RSSFeedArticle
    {
        return $this->article;
    }
}
