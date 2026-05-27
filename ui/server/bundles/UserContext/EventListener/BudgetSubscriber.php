<?php

declare(strict_types=1);

namespace Bundles\UserContext\EventListener;

use Bundles\UserContext\Event\BudgetPurchaseEvent;
use Symfony\Component\EventDispatcher\EventSubscriberInterface;

class BudgetSubscriber implements EventSubscriberInterface
{
    public static function getSubscribedEvents(): array
    {
        return [
            BudgetPurchaseEvent::class => 'onBudgetPurchase',
        ];
    }

    public function onBudgetPurchase(BudgetPurchaseEvent $event): void
    {
        $user = $event->getUser();
        $user->setBudget($user->getBudget() + $event->getAmount());
    }
}
