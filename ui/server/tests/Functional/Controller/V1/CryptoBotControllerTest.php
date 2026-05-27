<?php
namespace App\Tests\Functional\Controller\V1;

use App\Tests\RestTestCase;
use Bundles\UserContext\Entity\User;
use GuzzleHttp\ClientInterface;
use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\StreamInterface;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

class CryptoBotControllerTest extends RestTestCase
{
    private ClientInterface $guzzle;

    #[\Override]
    protected function services(): void
    {
        parent::services();

        $this->guzzle = $this->createMock(ClientInterface::class);
        self::getContainer()->set('test.client_interface', $this->guzzle);
    }

    public function testShouldAllowToDoCrudOperations(): void
    {
        $this->guzzle
            ->method('request')
            ->willReturnCallback(function (string $method, $uri, array $options = []) {
                $response = $this->createMock(ResponseInterface::class);
                $response->method('getStatusCode')->willReturn(200);
                $body = $this->createMock(StreamInterface::class);
                $json = '';
                if ('http://127.0.0.1:8095/deploy' === $uri) {
                    $json = file_get_contents(__DIR__ . '/dataset/deploy.json');
                }
                if (str_starts_with($uri, 'http://127.0.0.1:8000/health/check?botUuid=')) {
                    $json = file_get_contents(__DIR__ . '/dataset/health_check.json');
                }
                if (str_starts_with($uri, 'http://127.0.0.1:8000/trade/limit/list?botUuid=')) {
                    $json = file_get_contents(__DIR__ . '/dataset/trade_limit_list.json');
                }
                if (str_starts_with($uri, 'http://127.0.0.1:8000/trade/limit/update?botUuid=')) {
                    $json = file_get_contents(__DIR__ . '/dataset/trade_limit_update.json');
                }

                $body->method('getContents')->willReturn($json);
                $response->method('getBody')->willReturn($body);

                return $response;
            });

        $list = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_list')));
        $this->assertJsonSnapshot($list, 'list_empty');

        $new = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_post'), Request::METHOD_POST, [
            'apiKey'    => 'testapikey',
            'apiSecret' => 'apisecret',
            'provider'  => 'bybit',
        ]));
        $this->assertJsonSnapshot($new, 'new');

        $list = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_list')));
        $this->assertJsonSnapshot($list, 'list');

        $new = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_get', [
            'cryptobot' => $new['id'],
        ])));
        $this->assertJsonSnapshot($new, 'new');

        $updated = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_patch', [
            'cryptobot' => $new['id'],
        ]), Request::METHOD_PATCH, [
            'apiKey'             => 'testapikey2',
            'apiSecret'          => 'apisecret2',
            'cryptoTradeConfigs' => [
                [
                    'symbol'           => 'BTCUSDT',
                    'usdtLimit'        => 30.99,
                    'enabled'          => true,
                    'profitOptions'    => [
                        [
                            'index'           => 1,
                            'optionUnit'      => 'h',
                            'optionValue'     => 1,
                            'optionPercent'   => 2.25,
                            'isTriggerOption' => true,
                        ],
                    ],
                    'extraChargeOptions' => [
                        [
                            'index'      => 0,
                            'percent'    => -1.0,
                            'amountUsdt' => 10,
                        ],
                    ],
                ],
            ],
        ]));
        $this->assertJsonSnapshot($updated, 'updated');

        $user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $user->setBasicSubscriptionExpiresAt(new \DateTimeImmutable('+1 day'));
        $this->em->flush();

        $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_deploy', [
            'cryptobot' => $updated['id'],
        ]), Request::METHOD_PUT), Response::HTTP_NO_CONTENT);

        $deploy = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_get', [
            'cryptobot' => $updated['id'],
        ])));
        $this->assertJsonSnapshot($deploy, 'deploy');

        $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_delete', [
            'cryptobot' => $updated['id'],
        ]), Request::METHOD_DELETE), Response::HTTP_NO_CONTENT);
        $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_delete', [
            'cryptobot' => $updated['id'],
        ]), Request::METHOD_DELETE), Response::HTTP_NOT_FOUND);
    }

    public function testShouldAllowToGetServerList(): void
    {
        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_bot_server_list', [
            'cryptobot' => 1,
        ])));
        $this->assertJsonSnapshot($json);
    }

    public function testShouldAllowToSetBuyConditionsWithChildren(): void
    {
        $this->guzzle
            ->method('request')
            ->willReturnCallback(function (string $method, $uri, array $options = []) {
                $response = $this->createMock(ResponseInterface::class);
                $response->method('getStatusCode')->willReturn(200);
                $body = $this->createMock(StreamInterface::class);
                $json = '';

                if ('http://127.0.0.1:8095/deploy' === $uri) {
                    $json = file_get_contents(__DIR__ . '/dataset/deploy.json');
                }
                if (str_starts_with($uri, 'http://127.0.0.1:8000/health/check?botUuid=')) {
                    $json = file_get_contents(__DIR__ . '/dataset/health_check.json');
                }
                if (str_starts_with($uri, 'http://127.0.0.1:8000/trade/limit/list?botUuid=')) {
                    $json = file_get_contents(__DIR__ . '/dataset/trade_limit_list.json');
                }
                if (str_starts_with($uri, 'http://127.0.0.1:8000/trade/limit/update?botUuid=')) {
                    $json = file_get_contents(__DIR__ . '/dataset/trade_limit_update.json');
                }

                $body->method('getContents')->willReturn($json);
                $response->method('getBody')->willReturn($body);

                return $response;
            });

        $new = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_post'), Request::METHOD_POST, [
            'apiKey'    => 'testapikey',
            'apiSecret' => 'apisecret',
            'provider'  => 'bybit',
        ]));
        $this->assertJsonSnapshot($new, 'new');
        $updated = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_patch', [
            'cryptobot' => $new['id'],
        ]), Request::METHOD_PATCH, [
            'apiKey'             => 'testapikey2',
            'apiSecret'          => 'apisecret2',
            'cryptoTradeConfigs' => [
                [
                    'symbol'           => 'BTCUSDT',
                    'usdtLimit'        => 30.99,
                    'enabled'          => true,
                    'profitOptions'    => [
                        [
                            'index'           => 1,
                            'optionUnit'      => 'h',
                            'optionValue'     => 1,
                            'optionPercent'   => 2.25,
                            'isTriggerOption' => true,
                        ],
                    ],
                    'extraChargeOptions' => [
                        [
                            'index'      => 0,
                            'percent'    => -1.0,
                            'amountUsdt' => 10,
                        ],
                    ],
                ],
            ],
        ]));
        $conditions = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_put_buy_conditions', [
            'cryptobot' => $updated['id'],
        ]), Request::METHOD_PUT, [
            'symbol'     => $updated['cryptoTradeConfigs'][0]['symbol'],
            'conditions' => [
                [
                    'symbol'    => 'BTCUSDT',
                    'parameter' => 'aaa',
                    'condition' => '111',
                    'value'     => 50000,
                    'type'      => '222',
                    'children'  => [],
                ],
                [
                    'type'     => 'or',
                    'children' => [
                        [
                            'symbol'    => null,
                            'parameter' => '333',
                            'condition' => '444',
                            'value'     => null,
                            'type'      => '111',
                        ],
                    ],
                ],
            ],
        ]), 400);
        $this->assertJsonSnapshot($conditions, 'conditions_validate');

        $user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $user->setBasicSubscriptionExpiresAt(new \DateTimeImmutable('+1 day'));
        $this->em->flush();

        $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_deploy', [
            'cryptobot' => $updated['id'],
        ]), Request::METHOD_PUT), Response::HTTP_NO_CONTENT);

        $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_put_buy_conditions', [
            'cryptobot' => $updated['id'],
        ]), Request::METHOD_PUT, [
            'symbol'     => $updated['cryptoTradeConfigs'][0]['symbol'],
            'conditions' => [
                [
                    'symbol'    => 'BTCUSDT',
                    'parameter' => 'price',
                    'condition' => 'gte',
                    'value'     => 50000,
                    'type'      => 'or',
                    'children'  => [],
                ],
                [
                    'type'     => 'or',
                    'children' => [
                        [
                            'symbol'    => 'BTCUSDT',
                            'parameter' => 'price',
                            'condition' => 'gte',
                            'value'     => 50000,
                            'type'      => 'and',
                        ],
                        [
                            'symbol'    => 'BTCUSDT',
                            'parameter' => 'daily_percent',
                            'condition' => 'lte',
                            'value'     => -10.00,
                            'type'      => 'and',
                        ],
                    ],
                ],
            ],
        ]), Response::HTTP_NO_CONTENT);

        $new = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_get', [
            'cryptobot' => $updated['id'],
        ])));
        $this->assertJsonSnapshot($new, 'bot_updated');
    }

    public function testShouldAllowToGetListOfAvailableBotProviders(): void
    {
        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_available')));
        $this->assertJsonSnapshot($json, 'no_bots');

        $wrong = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_post'), Request::METHOD_POST, [
            'apiKey'    => 'testapikey',
            'apiSecret' => 'apisecret',
            'provider'  => 'fake',
        ]), Response::HTTP_BAD_REQUEST);
        $this->assertJsonSnapshot($wrong, 'wrong');

        $bybit = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_post'), Request::METHOD_POST, [
            'apiKey'    => 'testapikey',
            'apiSecret' => 'apisecret',
            'provider'  => 'bybit',
        ]));
        $this->assertJsonSnapshot($bybit, 'bybit');

        $duplicate = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_post'), Request::METHOD_POST, [
            'apiKey'    => 'testapikey',
            'apiSecret' => 'apisecret',
            'provider'  => 'bybit',
        ]), Response::HTTP_BAD_REQUEST);
        $duplicate['message'] = preg_replace('/instance #\d+/ui', 'instance #N', (string) $duplicate['message']);
        $this->assertJsonSnapshot($duplicate, 'duplicate');

        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_available')));
        $this->assertJsonSnapshot($json, 'created_bybit');

        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_available')));
        $this->assertJsonSnapshot($json, 'all_created');
    }

    public function testShouldAllowToGetExtendedCryptobotList(): void
    {
        $list = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_list_extended')));
        $this->assertJsonSnapshot($list);

        $user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $this->assertInstanceOf(User::class, $user);
        $user->setBasicSubscriptionExpiresAt(new \DateTimeImmutable('+1 day'));
        $this->em->flush();

        $list = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_list_extended')));
        $this->assertJsonSnapshot($list, 'zero_commission');
    }
}
