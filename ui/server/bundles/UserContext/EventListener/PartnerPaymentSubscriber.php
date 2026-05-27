<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\EventListener;

use Bundles\OxaPayContext\Entity\PromoCode;
use Bundles\UserContext\Event\PaymentCompletedEvent;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Component\EventDispatcher\EventSubscriberInterface;

class PartnerPaymentSubscriber implements EventSubscriberInterface
{
    private EntityManagerInterface $entityManager;

    public function __construct(EntityManagerInterface $entityManager)
    {
        $this->entityManager = $entityManager;
    }

    public static function getSubscribedEvents()
    {
        return [
            PaymentCompletedEvent::class => 'onPaymentCompleted',
        ];
    }

    public function onPaymentCompleted(PaymentCompletedEvent $event): void
    {
        $payment = $event->getPayment();
        $user    = $payment->getUser();

        $promoCode = $user->getPromoCode();
        if (!$promoCode instanceof PromoCode) {
            return;
        }

        $partner = $promoCode->getPartner();
        if (!$partner->isPartner()) {
            return;
        }

        $partnerFee = $payment->getAmount() * ($promoCode->getPartnerFeePercent() / 100);
        $partner->setPartnerBudget($partner->getPartnerBudget() + $partnerFee);
        $this->entityManager->flush();
    }
}
