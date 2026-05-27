<?php
namespace Functional\Controller\PublicRoute;

use App\DataFixtures\CryptoBotFixtures;
use App\Tests\RestTestCase;
use Bundles\TgBotContext\MessageBuilder;
use Symfony\Component\HttpFoundation\Request;

class CallbackControllerTest extends RestTestCase
{
    private MessageBuilder $messageBuilder;

    #[\Override]
    protected function services(): void
    {
        parent::services();

        $this->messageBuilder = $this->createMock(MessageBuilder::class);
        self::getContainer()->set('test.tg_message_builder', $this->messageBuilder);
    }

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
