<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\EventSubscriber;

use Bundles\CryptoBotContext\Event\NewTradeEvent;
use Bundles\CryptoBotContext\Event\StopBotEvent;
use Bundles\CryptoBotContext\Service\TradeSyncManager;
use Bundles\OxaPayContext\Entity\Transaction;
use Bundles\OxaPayContext\Service\CommissionService;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Component\EventDispatcher\EventSubscriberInterface;

class TradeEventSubscriber implements EventSubscriberInterface
{
    private EntityManagerInterface $entityManager;

    private TradeSyncManager $tradeSyncManager;

    private CommissionService $commissionService;

    public function __construct(
        EntityManagerInterface $entityManager,
        TradeSyncManager $tradeSyncManager,
        CommissionService $commissionService
    ) {
        $this->entityManager     = $entityManager;
        $this->tradeSyncManager  = $tradeSyncManager;
        $this->commissionService = $commissionService;
    }

    public static function getSubscribedEvents(): array
    {
        return [
            NewTradeEvent::class => 'onNewTrade',
            StopBotEvent::class  => 'onBotStop',
        ];
    }

    public function onNewTrade(NewTradeEvent $event): void
    {
        $trade = $event->getTrade();
        $user  = $trade->getUser();

        if ($user->hasActiveBasicSubscription()) {
            return;
        }

        $commission    = $this->commissionService->getCommission($trade->getBot());
        $commissionFee = max($commission->getMinValueUsd(), (($commission->getPercent() / 100) * $trade->getProfit()));

        $user->setBudget($user->getBudget() - $commissionFee);
        $transaction = new Transaction();
        $transaction
            ->setUser($user)
            ->setAmount($commissionFee);

        $this->entityManager->persist($transaction);
    }

    public function onBotStop(StopBotEvent $event): void
    {
        $this->tradeSyncManager->syncTrades($event->getCryptoBot(), null);
    }
}
