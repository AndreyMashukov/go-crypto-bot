<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Tests;

use KunicMarko\JMSMessengerAdapter\Serializer;
use PHPUnit\Framework\MockObject\MockObject;
use Symfony\Component\Messenger\Envelope;
use Symfony\Component\Messenger\Handler\MessageHandlerInterface;
use Symfony\Component\Messenger\MessageBusInterface;
use Symfony\Component\Messenger\TraceableMessageBus;

/**
 * @group functional
 */
class AsyncHandlerTestCase extends RestTestCase
{
    /** @var MessageBusInterface|MockObject */
    protected $bus;

    protected array $messagesStack = [];

    /**
     * Messenger Handler Map:
     * <class> => <handler>.
     */
    public const HANDLER_MAP = [
//        BotMessage::class                => 'test.bot_message_handler',
//        MonitoringPipelineMessage::class => 'test.monitoring_handler',
//        MonitoringMessage::class         => 'test.monitoring_message_handler',
    ];

    protected function busMock(): void
    {
        $this->bus = $this->createMock(TraceableMessageBus::class);

        self::$container->set('test.message.bus', $this->bus);
    }

    protected function busHandle(): void
    {
        /** @var Serializer $serializer */
        $serializer = self::$container->get('test.queue_serializer');

        $this->bus
            ->method('dispatch')
            ->willReturnCallback(function ($queueMessage) use ($serializer) {
                if (!isset(self::HANDLER_MAP[\get_class($queueMessage)])) {
                    $json = $serializer->encode(new Envelope(new \stdClass()));

                    return $serializer->decode($json);
                }

                $handler = self::$container->get(self::HANDLER_MAP[\get_class($queueMessage)]);

                if (!$handler instanceof MessageHandlerInterface) {
                    throw new \InvalidArgumentException('Invalid Message Given.');
                }

                // Emulate async invocation.
                //$this->em->clear();
                $json     = $serializer->encode(new Envelope($queueMessage));
                $envelope = $serializer->decode($json);

                /** @var object $decoded */
                $decoded = $envelope->getMessage();

                $handler->__invoke($decoded);

                $this->messagesStack[] = $queueMessage;

                return $envelope;
            });
    }
}
