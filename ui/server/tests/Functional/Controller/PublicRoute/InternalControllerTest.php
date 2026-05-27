<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Functional\Controller\PublicRoute;

use App\Tests\RestTestCase;
use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Repository\TradeRepository;
use GuzzleHttp\ClientInterface;
use PHPUnit\Framework\MockObject\MockObject;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * @group functional
 */
class InternalControllerTest extends RestTestCase
{
    /** @var ClientInterface|MockObject */
    private ClientInterface $guzzle;

    /** @var MockObject|TradeRepository */
    private TradeRepository $tradeRepository;

    protected function services(): void
    {
        parent::services();

        $this->guzzle          = $this->createMock(ClientInterface::class);
        $this->tradeRepository = $this->createMock(TradeRepository::class);
        self::getContainer()->set('test.client_interface', $this->guzzle);
        self::getContainer()->set('test.trade_repository', $this->tradeRepository);
    }

    /**
     * Should allow to authenticate by simple token.
     */
    public function testShouldAllowToAuthenticateBySimpleToken(): void
    {
        $response = $this->apiPublicRequest($this->getUrl('public_internal_signal'), Request::METHOD_POST);
        $this->assertEquals(Response::HTTP_FORBIDDEN, $response->getStatusCode());

        $response = $this->apiPublicRequest($this->getUrl('public_internal_signal'), Request::METHOD_POST, [], [
            'HTTP_Crypto-Internal-Token' => 'internal-token',
        ]);
        $this->assertEquals(Response::HTTP_BAD_REQUEST, $response->getStatusCode());
    }

    /**
     * Should allow to handle signal for subscribed config.
     */
    public function testShouldAllowToHandleSignalForSubscribedConfig(): void
    {
        $botRequest = [];

        $this->guzzle
            ->method('request')
            ->willReturnCallback(function (string $method, $uri, array $options = []) use (&$botRequest) {
                $botRequest = [
                    'method'  => $method,
                    'uri'     => $uri,
                    'options' => $options,
                ];
            });

        $signal   = json_decode(file_get_contents(__DIR__ . '/signal.json'), true);
        $response = $this->apiPublicRequest($this->getUrl('public_internal_signal'), Request::METHOD_POST, $signal, [
            'HTTP_Crypto-Internal-Token' => 'internal-token',
        ]);
        $this->assertEquals(Response::HTTP_NO_CONTENT, $response->getStatusCode());
        $this->assertJsonSnapshot($botRequest, 'bot_request');
    }

    /**
     * Should allow to handle signal for subscribed config and modify.
     */
    public function testShouldAllowToHandleSignalForSubscribedConfigAndModify(): void
    {
        $botRequest = [];

        $config = $this->em->getRepository(CryptoTradeConfig::class)->findOneBy([
            'symbol'        => 'PERPUSDT',
            'signalTrading' => true,
        ]);
        $this->assertInstanceOf(CryptoTradeConfig::class, $config);
        $config->getSignalConfig()
            ->setAvgBuyCorrection(true)
            ->setAvgSellCorrection(true)
            ->setAvgBuyFilter(true)
            ->setAvgSellFilter(true)
            ->setPercentFilter(1.50)
            ->setRatingFilter(true)
        ;
        $this->em->flush();

        $this->tradeRepository->expects($this->exactly(2))
            ->method('getBestMonthSymbols')
            ->willReturn([
                'PERPUSDT' => [
                    'avgBuyPrice'          => 1.184375,
                    'avgSellPrice'         => 1.198,
                    'percent'              => 1.15,
                    'avgPositionTimeHours' => 8,
                    'totalProfit'          => 6.03,
                    'trades'               => 4,
                    'profitPerTrade'       => 1.51,
                    'symbol'               => 'PERPUSDT',
                ],
            ]);

        $this->guzzle
            ->method('request')
            ->willReturnCallback(function (string $method, $uri, array $options = []) use (&$botRequest) {
                $botRequest = [
                    'method'  => $method,
                    'uri'     => $uri,
                    'options' => $options,
                ];
            });

        $signal   = json_decode(file_get_contents(__DIR__ . '/signal.json'), true);
        $response = $this->apiPublicRequest($this->getUrl('public_internal_signal'), Request::METHOD_POST, $signal, [
            'HTTP_Crypto-Internal-Token' => 'internal-token',
        ]);
        $this->assertEquals(Response::HTTP_NO_CONTENT, $response->getStatusCode());
        $this->assertJsonSnapshot($botRequest, 'bot_request');
    }
}
