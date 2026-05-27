<?php
namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Exception\SentimentNoRelatedCoinsException;
use Bundles\CryptoBotContext\Model\RSSFeedArticle;
use Bundles\CryptoBotContext\Model\SentimentResult;
use Bundles\CryptoBotContext\Service\Traits\SentimentAverageScoreTrait;
use Bundles\TgBotContext\Service\AlertService;
use GuzzleHttp\ClientInterface;
use Psr\Log\LoggerInterface;
use Symfony\Component\HttpFoundation\Request;

class SentimentService
{
    use SentimentAverageScoreTrait;

    public function __construct(private AlertService $alertService, private LoggerInterface $logger, private ClientInterface $client, private string $env)
    {
    }

    public function sentimentAnalysis(RSSFeedArticle $article): SentimentResult
    {
        $results = [
            SentimentResult::LABEL_BEARISH => [0, 0.00],
            SentimentResult::LABEL_BULLISH => [0, 0.00],
            SentimentResult::LABEL_NEUTRAL => [0, 0.00],
        ];

        $coins = [];
        $texts = $article->getFullText();

        if ($texts === []) {
            $texts[] = "{$article->getTitle()}. {$article->getShortText()}";
        }

        $host = 'host.docker.internal';
        if ('prod' === $this->env) {
            $host = 'go_crypto_saas_sentiment';
        }

        foreach ($texts as $text) {
            $text = str_replace('-', '', str_replace('"', '', $text));
            $text = preg_replace('/\n/ui', '', $text);

            try {
                $response = $this->client->request(Request::METHOD_POST, "http://{$host}:8080/sentiment/predict", [
                    'timeout' => 40,
                    'json'    => [
                        'text' => trim($text),
                    ],
                ]);
                $json = json_decode($response->getBody()->getContents(), true);
                if ($json) {
                    ++$results[$json['label']][0];
                    $results[$json['label']][1] += $json['score'];

                    $coins = $json['coins'] ?? [];
                }
            } catch (\Throwable $throwable) {
                $this->logger->error($throwable->getMessage(), [
                    'line' => $throwable->getLine(),
                    'file' => $throwable->getFile(),
                ]);

                $this->alertService->alert("SentimentService: {$throwable->getMessage()}");
            }
        }

        if (!$coins) {
            throw new SentimentNoRelatedCoinsException('No related coins found.', $article);
        }

        return $this->getAverageResult($article, $results, $coins);
    }
}
