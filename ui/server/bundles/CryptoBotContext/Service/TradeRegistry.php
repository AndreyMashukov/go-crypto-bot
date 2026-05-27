<?php
namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\Trade;
use Bundles\CryptoBotContext\Event\NewTradeEvent;
use Bundles\CryptoBotContext\Repository\TradeRepository;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Component\EventDispatcher\EventDispatcherInterface;

class TradeRegistry
{
    public function __construct(private readonly EntityManagerInterface $entityManager, private readonly TradeRepository $tradeRepository, private readonly EventDispatcherInterface $eventDispatcher)
    {
    }

    public function registerTrade(array $tradeData, CryptoBot $cryptoBot): void
    {
        $user  = $cryptoBot->getUser();
        $trade = $this->tradeRepository->findOneBy([
            'user'    => $user->getId(),
            'orderId' => $tradeData['orderId'],
        ]);

        if ($trade instanceof Trade) {
            return;
        }

        $this->entityManager->wrapInTransaction(function () use ($user, $cryptoBot, $tradeData) {
            $trade = new Trade();
            $trade->setUser($user);
            $trade->setProfit($tradeData['profit']);
            $trade->setOrderId($tradeData['orderId']);
            $trade->setBuyDate(\DateTimeImmutable::createFromFormat('Y-m-d H:i:s', $tradeData['open']));
            $trade->setSellDate(\DateTimeImmutable::createFromFormat('Y-m-d H:i:s', $tradeData['close']));
            $trade->setBuyQty($tradeData['buyQuantity']);
            $trade->setSellQty($tradeData['sellQuantity']);
            $trade->setBuyPrice($tradeData['buy']);
            $trade->setSellPrice($tradeData['sell']);
            $trade->setSymbol($tradeData['symbol']);
            $trade->setPercent($tradeData['percent']);
            $trade->setBot($cryptoBot);

            $this->entityManager->persist($trade);
            $this->eventDispatcher->dispatch(new NewTradeEvent($trade));
            $this->entityManager->flush();
        });
    }
}
