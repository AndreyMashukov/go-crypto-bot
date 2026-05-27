<?php
namespace Functional\Controller\External;

use App\Tests\RestTestCase;
use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Entity\RSSArticle;
use Bundles\OxaPayContext\Service\PaidServiceManager;
use Bundles\UserContext\Entity\User;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

class ExternalApiControllerTest extends RestTestCase
{
    public function testShouldAllowToBuySubscriptionAndGetSentimentList(): void
    {
        $user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $this->assertInstanceOf(User::class, $user);
        $user->setBudget(90);
        $this->em->flush();
$response = $this->apiRequest($this->getUrl('external_api_sentiment_list'));
        $this->assertEquals(Response::HTTP_FORBIDDEN, $response->getStatusCode());

        $this->deserialize($this->apiRequest($this->getUrl('v1_service_purchase'), Request::METHOD_POST, [
            'code' => PaidServiceManager::API_SUBSCRIPTION_CODE,
        ]), Response::HTTP_NO_CONTENT);
$user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $this->assertInstanceOf(User::class, $user);
        $this->assertEquals(1, $user->getBudget());

        $config = $this->em->getRepository(CryptoTradeConfig::class)->findOneBy([
            'symbol' => 'PERPUSDT',
        ]);
        $this->assertInstanceOf(CryptoTradeConfig::class, $config);
        $config->setScore(0.5);
        $config->setLabel('BEARISH');
        $article = new RSSArticle();
        $article->setLabel('BEARISH')
            ->setScore(0.5)
            ->setCoins(['PERP'])
            ->setUrl('https://test.url')
            ->setExpired(false)
            ->setExpiresAt(new \DateTimeImmutable('2024-01-01 00:00:00'))
            ->setUrlCrc32(crc32('https://test.url'))
        ;
        $this->em->persist($article);
        $this->em->flush();
        $json = $this->deserialize($this->apiRequest($this->getUrl('external_api_sentiment_list')));
        $this->assertJsonSnapshot($json, 'sentiment');
    }
}
