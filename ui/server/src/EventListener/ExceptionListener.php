<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\EventListener;

use App\Exception\MessageException;
use Psr\Log\LoggerInterface;
use Symfony\Component\ErrorHandler\Exception\FlattenException;
use Symfony\Component\EventDispatcher\EventSubscriberInterface;
use Symfony\Component\Messenger\Event\WorkerMessageFailedEvent;
use Symfony\Component\Messenger\Stamp\ErrorDetailsStamp;

class ExceptionListener implements EventSubscriberInterface
{
    private LoggerInterface $logger;

    public function __construct(LoggerInterface $logger)
    {
        $this->logger = $logger;
    }

    public function onMessageFailed(WorkerMessageFailedEvent $event)
    {
        $this->logger->error($event->getThrowable()->getMessage());
        $message         = $event->getThrowable()->getMessage();
        $customException = new MessageException($message, 500);

        $reflectionClass = new \ReflectionClass($event);
        $prop            = $reflectionClass->getProperty('throwable');
        $prop->setAccessible(true);
        $prop->setValue($event, $customException);
        $stamp         = ErrorDetailsStamp::create($event->getThrowable());
        $previousStamp = $event->getEnvelope()->last(ErrorDetailsStamp::class);
        if (null === $previousStamp) {
            $event->addStamps($stamp);

            return;
        }

        $reflectionClass = new \ReflectionClass($previousStamp);
        $prop            = $reflectionClass->getProperty('flattenException');
        $prop->setAccessible(true);

        $cusFlat = FlattenException::createFromThrowable($customException);

        $prop->setValue($previousStamp, $cusFlat);

        // Do not append duplicate information
        if (!$previousStamp->equals($stamp)) {
            $event->addStamps($stamp);
        }
    }

    public static function getSubscribedEvents(): array
    {
        return [
            WorkerMessageFailedEvent::class => ['onMessageFailed', 255],
        ];
    }
}
