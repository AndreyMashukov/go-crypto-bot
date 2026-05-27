<?php
namespace App\Tests\Functional\Controller\PublicRoute;

use App\Tests\RestTestCase;
use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Service\CryptoBotService;

class CryptoBotControllerTest extends RestTestCase
{
    private CryptoBotService $cryptoBotService;

    #[\Override]
    protected function services(): void
    {
        $this->cryptoBotService = $this->createMock(CryptoBotService::class);
        self::getContainer()->set('test.cryptobot_service', $this->cryptoBotService);
    }

    public function testShouldAllowGetBotPositionsFromCache(): void
    {
        self::getContainer()->get('test.cached_trade_list_interface')->invalidateCache();

        $cryptoBot = $this->em->find(CryptoBot::class, 1);

        $result = json_decode(file_get_contents(__DIR__ . '/positions.json'), true);
        $this->cryptoBotService
            ->expects($this->once())
            ->method('getPositionsList')
            ->with($cryptoBot)
            ->willReturn($result);

        for ($i = 0; $i < 10; ++$i) {
            $json = $this->deserialize($this->apiPublicRequest($this->getUrl('public_crypto_bot_positions')));
            $this->assertJsonSnapshot($json);
        }
    }
}
