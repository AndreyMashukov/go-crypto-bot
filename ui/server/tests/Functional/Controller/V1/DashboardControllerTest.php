<?php
namespace App\Tests\Functional\Controller\V1;

use App\Tests\RestTestCase;
use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Entity\Server;
use Bundles\CryptoBotContext\Service\CryptoBotService;
use Bundles\UserContext\Entity\User;
use GuzzleHttp\ClientInterface;
use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\StreamInterface;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

class DashboardControllerTest extends RestTestCase
{
    private ClientInterface $guzzle;

    private CryptoBotService $cryptoBotService;

    #[\Override]
    protected function services(): void
    {
        parent::services();

        $this->guzzle           = $this->createMock(ClientInterface::class);
        $this->cryptoBotService = $this->createMock(CryptoBotService::class);

        self::getContainer()->set('test.client_interface', $this->guzzle);
        self::getContainer()->set('test.cryptobot_service', $this->cryptoBotService);
    }

    public function testShouldAllowToUpdateSingleConfig(): void
    {
        $this->guzzle
            ->method('request')
            ->willReturnCallback(function (string $method, $uri) {
                $response = $this->createMock(ResponseInterface::class);
                $response->method('getStatusCode')->willReturn(200);
                $body = $this->createMock(StreamInterface::class);
                $json = '';

                if (str_starts_with($uri, 'http://127.0.0.1:8000/trade/limit/list?botUuid=')) {
                    $json = file_get_contents(__DIR__ . '/../V1/dataset/trade_limit_list.json');
                }
                if (str_starts_with($uri, 'http://127.0.0.1:8000/trade/limit/update?botUuid=')) {
                    $json = file_get_contents(__DIR__ . '/../V1/dataset/trade_limit_update.json');
                }

                $body->method('getContents')->willReturn($json);
                $response->method('getBody')->willReturn($body);

                return $response;
            });

        $user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $this->assertInstanceOf(User::class, $user);
        $server = $this->em->getRepository(Server::class)->findOneBy([
            'ip' => '127.0.0.1',
        ]);
        $this->assertInstanceOf(Server::class, $server);

        $user->setSignalSubscriptionExpiresAt(new \DateTimeImmutable('+2 hours'));
        $cryptoBot = (new CryptoBot($user))
            ->setStatus(CryptoBot::STATUS_RUNNING)
            ->setPort('8000')
            ->setServer($server)
            ->setContainerId('xxxxxxxx')
            ->setApiKey('test')
            ->setApiSecret('test')
            ->setProvider('binance')
            ->setUuid('87bf6369-3a27-4d2f-b26c-f90aed0c955b');

        $config = new CryptoTradeConfig();
        $config->setCryptobot($cryptoBot)
            ->setSymbol('PERPUSDT')
            ->setEnabled(true)
            ->setSignalTrading(true)
            ->setUsdtLimit(100)
            ->setExtraChargeOptions([['c'], ['b'], ['a']])
            ->setAvgConditions([])
        ;
        $cryptoBot->addCryptoTradeConfig($config);
        $this->em->persist($config);
        $this->em->persist($cryptoBot);
        $this->em->flush();

        $url = $this->getUrl('v1_dashboard_cryptobot_symbol_update', [
            'cryptobot' => $cryptoBot->getId(),
            'symbol'    => $config->getSymbol(),
        ]);
        $response = $this->apiRequest($url, Request::METHOD_PATCH, [
            'usdtLimit'    => 500.00,
            'signalConfig' => [
                'avgBuyFilter'            => true,
                'percentFilter'           => 2.55,
                'sellPriceCorrectionMode' => 'max',
                'signalPeriodDays'        => 14,
            ],
        ]);
        $json = $this->deserialize($response);
        $this->assertJsonSnapshot($json);

        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_symbol_list', [
            'cryptobot' => $cryptoBot->getId(),
        ])));
        $this->assertJsonSnapshot($json, 'symbol_list_exchange');
        $symbols = [];
        foreach ($json as $item) {
            $symbols[] = $item['symbol'];
        }
        $this->assertContains($config->getSymbol(), $symbols);
        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_dashboard_symbol_available', [
            'cryptobot' => $cryptoBot->getId(),
        ])));
        $this->assertJsonSnapshot($json, 'symbol_available_exchange');
        $symbols = [];
        foreach ($json as $item) {
            $symbols[] = $item['symbol'];
        }
        $this->assertNotContains($config->getSymbol(), $symbols);

        $addedSymbol = $json[0]['symbol'];

        $response = $this->apiRequest($this->getUrl('v1_dashboard_quick_symbol', [
            'cryptobot' => $cryptoBot->getId(),
        ]), Request::METHOD_POST, [
            'exchangeSymbol' => $json[0]['id'],
            'restartBot'     => 0,
        ]);
        $json = $this->deserialize($response, Response::HTTP_NO_CONTENT);
        $this->assertJsonSnapshot($json, 'quick_add');

        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_dashboard_symbol_available', [
            'cryptobot' => $cryptoBot->getId(),
        ])));
        $this->assertJsonSnapshot($json, 'symbol_available_exchange_2');
        $symbols = [];
        foreach ($json as $item) {
            $symbols[] = $item['symbol'];
        }
        $this->assertNotContains($addedSymbol, $symbols);

        $symbolConfig = $this->em->getRepository(CryptoTradeConfig::class)->findOneBy([
            'symbol'    => $addedSymbol,
            'cryptobot' => $cryptoBot,
        ]);
        $this->assertInstanceOf(CryptoTradeConfig::class, $symbolConfig);
        $this->assertFalse($symbolConfig->isEnabled());
        $this->assertJsonSnapshot($symbolConfig->getProfitOptions(), 'profit_new_symbol');
        $this->assertJsonSnapshot($symbolConfig->getBuyConditions(), 'buy_condition_new_symbol');
        $this->assertEmpty($symbolConfig->getAvgConditions());
        $this->assertEmpty($symbolConfig->getSellConditions());
        $this->assertEmpty($symbolConfig->getExtraChargeOptions());
        $this->assertEquals(50.00, $symbolConfig->getUsdtLimit());
    }

    public function testShouldAllowToGetTradeStackV2(): void
    {
        $this->guzzle
            ->method('request')
            ->willReturnCallback(function (string $method, $uri, array $options = []) {
                $response = $this->createMock(ResponseInterface::class);
                $response->method('getStatusCode')->willReturn(200);
                $body = $this->createMock(StreamInterface::class);
                $json = '';
                if (str_starts_with($uri, 'http://127.0.0.1:8000/health/check?botUuid=')) {
                    $json = file_get_contents(__DIR__ . '/dataset/health_check.json');
                }
                if (str_starts_with($uri, 'http://127.0.0.1:8000/trade/limit/list?botUuid=')) {
                    $json = file_get_contents(__DIR__ . '/dataset/trade_limit_list.json');
                }
                if (str_starts_with($uri, 'http://localhost:8080/stats/pivot/grid')) {
                    $json = file_get_contents(__DIR__ . '/../../pivot_grid.json');
                }

                $body->method('getContents')->willReturn($json);
                $response->method('getBody')->willReturn($body);

                return $response;
            });

        $stack = json_decode(file_get_contents(__DIR__ . '/dataset/trade_stack.json'), true);
        $this->cryptoBotService
            ->expects($this->once())
            ->method('getTradeStack')
            ->willReturn($stack);

        $cryptoBot = $this->em->find(CryptoBot::class, 1);
        $this->assertInstanceOf(CryptoBot::class, $cryptoBot);
        $config = $cryptoBot->getCryptoTradeConfigs()->first();
        $this->assertInstanceOf(CryptoTradeConfig::class, $config);
        $config->setLabel('BEARISH')->setScore(0.80);
        $this->em->flush();

        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_dashboard_stack_v2', [
            'cryptobot' => $cryptoBot->getId(),
        ])));
        $this->assertJsonSnapshot($json);
    }

    public function testShouldAllowToGetProfitReport(): void
    {
        $cryptoBot = $this->em->find(CryptoBot::class, 1);
        $this->assertInstanceOf(CryptoBot::class, $cryptoBot);

        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_dashboard_profit', [
            'cryptobot' => $cryptoBot->getId(),
        ])));
        $this->assertJsonSnapshot($json, 'empty');

        $tradeData = [
            'profit'       => 100,
            'orderId'      => 999,
            'open'         => '2024-12-01 00:00:00',
            'close'        => '2024-12-01 01:00:00',
            'buyQuantity'  => 1,
            'sellQuantity' => 1,
            'buy'          => 100,
            'sell'         => 110,
            'symbol'       => 'BTCUSDT',
            'percent'      => 10.00,
        ];

        $cryptoBot = $this->em->find(CryptoBot::class, $cryptoBot->getId());
        $this->assertInstanceOf(CryptoBot::class, $cryptoBot);

        $tradeRegistry = self::getContainer()->get('test.trade_registry');
        $tradeRegistry->registerTrade($tradeData, $cryptoBot);

        $cryptoBot = $this->em->find(CryptoBot::class, $cryptoBot->getId());
        $this->assertInstanceOf(CryptoBot::class, $cryptoBot);

        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_dashboard_profit', [
            'cryptobot' => $cryptoBot->getId(),
        ])));
        $this->assertJsonSnapshot($json, 'has_trade');

        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_dashboard_profit', [
            'cryptobot' => $cryptoBot->getId(),
            'period'    => 'month',
        ])));
        $this->assertJsonSnapshot($json, 'has_trade_month');
    }

    public function testShouldAllowGetBotPositionsFromCache(): void
    {
        $this->guzzle
            ->method('request')
            ->willReturnCallback(function (string $method, $uri, array $options = []) {
                $response = $this->createMock(ResponseInterface::class);
                $response->method('getStatusCode')->willReturn(200);
                $body = $this->createMock(StreamInterface::class);
                $json = '';
                if (str_starts_with($uri, 'http://localhost:8080/stats/pivot/grid')) {
                    $json = file_get_contents(__DIR__ . '/../../pivot_grid.json');
                }

                $body->method('getContents')->willReturn($json);
                $response->method('getBody')->willReturn($body);

                return $response;
            });

        $cryptoBot = $this->em->find(CryptoBot::class, 1);

        $cryptoBotConfig = new CryptoTradeConfig();
        $cryptoBotConfig->setCryptobot($cryptoBot)
            ->setSymbol('SOLUSDT')
            ->setBuyConditions([])
            ->setSellConditions([])
            ->setAvgConditions([])
            ->setEnabled(true)
            ->setSignalTrading(true)
            ->setPriceTodayFirst(160.76)
            ->setPriceTodayLast(150.00)
            ->setPositionUpdatedAt(new \DateTimeImmutable())
            ->setAveragePrice(170.76)
            ->setQuantity(0.878)
            ->setUsdtLimit(100)
        ;
        $this->em->persist($cryptoBotConfig);
        $this->em->flush();

        $result = json_decode(file_get_contents(__DIR__ . '/../PublicRoute/positions.json'), true);
        $this->cryptoBotService
            ->expects($this->exactly(2))
            ->method('getPositionsList')
            ->willReturn($result);

        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_dashboard_positions', [
            'cryptobot' => $cryptoBot->getId(),
        ])));
        $this->assertJsonSnapshot($json);

        $cryptoBotConfig = $this->em->find(CryptoTradeConfig::class, $cryptoBotConfig->getId());
        $cryptoBotConfig->setPriceTodayFirst(175.00)
            ->setPriceTodayLast(177.00);
        $this->em->flush();
        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_dashboard_positions', [
            'cryptobot' => $cryptoBot->getId(),
        ])));
        $this->assertJsonSnapshot($json, 'positive');
    }
}
