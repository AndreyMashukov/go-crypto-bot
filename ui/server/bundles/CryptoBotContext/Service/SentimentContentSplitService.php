<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Service;

class SentimentContentSplitService
{
    // Go-sentiment microservice is able to process short sentiments.
    public const MAX_SENTENCE_LENGTH = 2000;

    public function split(string $content): array
    {
        $batch      = [];
        $exploded   = explode('.', trim(str_replace("'", '', $content)));
        $resultPart = '';
        foreach ($exploded as $explode) {
            switch (true) {
                case mb_strlen("{$resultPart} {$explode}") <= self::MAX_SENTENCE_LENGTH:
                    $resultPart .= " {$explode}";
                    break;
                case $resultPart && mb_strlen($explode) <= self::MAX_SENTENCE_LENGTH:
                    $batch[]    = $resultPart;
                    $resultPart = $explode;
                    break;
                case mb_strlen($explode) > self::MAX_SENTENCE_LENGTH:
                    if ($resultPart) {
                        $batch[] = $resultPart;
                    }
                    $resultPart = mb_substr($explode, 0, self::MAX_SENTENCE_LENGTH);
                    break;
            }
        }

        if ($resultPart) {
            $batch[] = $resultPart;
        }

        return $batch;
    }
}
