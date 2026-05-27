<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Tests\Functional\Controller\PublicRoute;

use App\Tests\RestTestCase;
use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Service\CryptoBotService;
use PHPUnit\Framework\MockObject\MockObject;

/**
 * @group functional
 */
class CryptoBotControllerTest extends RestTestCase
{
    /** @var CryptoBotService|MockObject */
    private CryptoBotService $cryptoBotService;

    protected function services(): void
    {
        $this->cryptoBotService = $this->createMock(CryptoBotService::class);
        self::getContainer()->set('test.cryptobot_service', $this->cryptoBotService);
    }

    /**
     * Should allow to get bot positions from cache.
     */
    public function testShouldAllowGetBotPositionsFromCache(): void
    {
        self::getContainer()->get('test.cached_trade_list_interface')->invalidateCache();

        // Get first bot
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
