<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

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
