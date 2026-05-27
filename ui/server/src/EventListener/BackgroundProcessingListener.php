<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\EventListener;

use App\Service\BackgroundProcessing;
use Symfony\Component\EventDispatcher\EventSubscriberInterface;
use Symfony\Component\HttpKernel\Event\TerminateEvent;

class BackgroundProcessingListener implements EventSubscriberInterface
{
    private BackgroundProcessing $backgroundProcessing;

    public function __construct(BackgroundProcessing $backgroundProcessing)
    {
        $this->backgroundProcessing = $backgroundProcessing;
    }

    public static function getSubscribedEvents()
    {
        return [
            TerminateEvent::class => ['onTerminate', 10],
        ];
    }

    public function onTerminate(TerminateEvent $event): void
    {
        unset($event);

        $this->backgroundProcessing->runTasks();
    }
}
