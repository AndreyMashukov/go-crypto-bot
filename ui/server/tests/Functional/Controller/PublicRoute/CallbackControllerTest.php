<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Functional\Controller\PublicRoute;

use App\DataFixtures\CryptoBotFixtures;
use App\Tests\RestTestCase;
use Bundles\TgBotContext\MessageBuilder;
use PHPUnit\Framework\MockObject\MockObject;
use Symfony\Component\HttpFoundation\Request;

/**
 * @group functional
 */
class CallbackControllerTest extends RestTestCase
{
    /** @var MessageBuilder|MockObject */
    private MessageBuilder $messageBuilder;

    protected function services(): void
    {
        parent::services();

        $this->messageBuilder = $this->createMock(MessageBuilder::class);
        self::getContainer()->set('test.tg_message_builder', $this->messageBuilder);
    }

    /**
     * Should match bot by UUID.
     */
    public function testShouldMatchBotByUUID(): void
    {
        $this->messageBuilder
            ->expects($this->once())
            ->method('fromOrderMessage')
            ->willReturn([]);

        $response = $this->apiPublicRequest($this->getUrl('public_callback_telegram'), Request::METHOD_POST, [
            'bot'       => CryptoBotFixtures::TEST_CRYPTO_BOT_UUID,
            'dateTime'  => 'aaa',
            'symbol'    => 'aaa',
            'amount'    => 0.0005,
            'price'     => 10.00001,
            'operation' => 'aaa',
            'details'   => 'aaa',
        ]);
        $json = $this->deserialize($response);
        $this->assertJsonSnapshot($json);
    }
}
