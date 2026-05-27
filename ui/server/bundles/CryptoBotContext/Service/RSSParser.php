<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Entity\RSSArticle;
use Bundles\CryptoBotContext\Model\RSSFeedArticle;
use Bundles\CryptoBotContext\Repository\RSSArticleRepository;
use GuzzleHttp\ClientInterface;
use Psr\Log\LoggerInterface;
use Symfony\Component\HttpFoundation\Request;

class RSSParser
{
    private array $rssFeeds;

    private ClientInterface $client;

    private RSSArticleRepository $repository;

    private SentimentContentSplitService $splitService;

    private LoggerInterface $logger;

    public function __construct(
        array $rssFeeds,
        ClientInterface $client,
        RSSArticleRepository $repository,
        SentimentContentSplitService $splitService,
        LoggerInterface $logger
    ) {
        $this->rssFeeds     = $rssFeeds;
        $this->client       = $client;
        $this->repository   = $repository;
        $this->splitService = $splitService;
        $this->logger       = $logger;
    }

    /**
     * @return \Generator|RSSFeedArticle
     */
    public function iterateArticles(): \Generator
    {
        foreach ($this->rssFeeds as $rssFeed) {
            $seen = 0;

            try {
                $response = $this->client->request(Request::METHOD_GET, $rssFeed, [
                    'headers' => [
                        'User-Agent' => 'Googlebot-News',
                    ],
                ]);
                $xmlFeed  = $response->getBody()->getContents();
                if (!$xmlFeed) {
                    continue;
                }

                $simpleXml = new \SimpleXMLElement($xmlFeed);

                foreach ($simpleXml->channel->item as $channelItem) {
                    $publishedAt = new \DateTimeImmutable($channelItem->pubDate);
                    if ($publishedAt->getTimestamp() < (new \DateTimeImmutable('today'))->getTimestamp()) {
                        break;
                    }

                    $url = $channelItem->link;

                    $entity = $this->repository->findOneBy([
                        'urlCrc32' => crc32($url),
                        'url'      => $url,
                    ]);

                    if ($entity instanceof RSSArticle) {
                        ++$seen;
                        if ($seen > 10) {
                            break;
                        }
                        continue;
                    }

                    $content = $this->getContent($url);

                    yield new RSSFeedArticle(
                        $url,
                        $channelItem->title, // todo: title
                        trim(strip_tags($channelItem->description)), // todo: shortText
                        $content
                    );
                }
            } catch (\Throwable $throwable) {
                $this->logger->error($throwable->getMessage(), [
                    'line' => $throwable->getLine(),
                    'file' => $throwable->getFile(),
                ]);

                unset($throwable);
            }
        }
    }

    private function getContent(string $url): array
    {
        preg_match('/https?:\/\/(?P<host>[a-z0-9-\.]+)\//ui', $url, $matches);

        if (!isset($matches['host'])) {
            return [];
        }

        $xpathMap = [
            'www.newsbtc.com'    => '//div[contains(@class, "content-inner")]//p',
            'newsbtc.com'        => '//div[contains(@class, "content-inner")]//p',
            'bitcoinist.com'     => '//div[contains(@class, "entry-content")]//p',
            'www.coindesk.com'   => '//article[contains(@class, "default__ArticleWrapper")]//p',
            'www.investing.com'  => '//div[contains(@id, "article")]//p',
            'insidebitcoins.com' => '//article[contains(@class, "post")]//p',
        ];

        $xpathSelector = $xpathMap[$matches['host']] ?? null;
        if (!$xpathSelector) {
            return [];
        }

        try {
            $response = $this->client->request(Request::METHOD_GET, $url, [
                'headers' => [
                    'User-Agent' => 'Googlebot-News',
                    'Host'       => $matches['host'],
                ],
            ]);
            $content  = $response->getBody()->getContents();
            if (!$content) {
                return [];
            }

            $dom = new \DOMDocument('1.0', 'UTF-8');
            @$dom->loadHTML($content, LIBXML_NOERROR);
            $xpath = new \DOMXPath($dom);
            $list  = $xpath->query($xpathSelector);
            $parts = [];

            /** @var \DOMElement $item */
            foreach ($list as $item) {
                $contentPart = trim(strip_tags($item->textContent));

                if ($contentPart) {
                    $parts[] = $contentPart;
                }
            }

            return $this->splitService->split(implode(' ', $parts));
        } catch (\Throwable $throwable) {
            $this->logger->error($throwable->getMessage(), [
                'line' => $throwable->getLine(),
                'file' => $throwable->getFile(),
            ]);

            return [];
        }
    }
}
