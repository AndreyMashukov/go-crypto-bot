<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Service;

class SentimentContentSplitService
{
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
                    if ($resultPart !== '' && $resultPart !== '0') {
                        $batch[] = $resultPart;
                    }
                    $resultPart = mb_substr($explode, 0, self::MAX_SENTENCE_LENGTH);
                    break;
            }
        }

        if ($resultPart !== '' && $resultPart !== '0') {
            $batch[] = $resultPart;
        }

        return $batch;
    }
}
