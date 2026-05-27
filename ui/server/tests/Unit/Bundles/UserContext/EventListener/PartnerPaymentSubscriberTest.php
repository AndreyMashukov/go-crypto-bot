<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Unit\Bundles\UserContext\EventListener;

use Bundles\OxaPayContext\Entity\Payment;
use Bundles\OxaPayContext\Entity\PromoCode;
use Bundles\UserContext\Entity\User;
use Bundles\UserContext\Event\PaymentCompletedEvent;
use Bundles\UserContext\EventListener\PartnerPaymentSubscriber;
use Doctrine\ORM\EntityManagerInterface;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;

/**
 * @group unit
 */
class PartnerPaymentSubscriberTest extends TestCase
{
    private PartnerPaymentSubscriber $subscriber;

    /** @var EntityManagerInterface|MockObject */
    private EntityManagerInterface $entityManager;

    protected function setUp(): void
    {
        $this->entityManager = $this->createMock(EntityManagerInterface::class);
        $this->subscriber    = new PartnerPaymentSubscriber($this->entityManager);
    }

    public function testShouldAllowToAddPartnerBalanceAfterPaymentCompleted(): void
    {
        $this->entityManager
            ->expects($this->once())
            ->method('flush');

        $partner = $this->createMock(User::class);
        $partner
            ->expects($this->once())
            ->method('getPartnerBudget')
            ->willReturn(5.00);

        $partner
            ->expects($this->once())
            ->method('setPartnerBudget')
            ->with(10.00);

        $partner->expects($this->once())
            ->method('isPartner')
            ->willReturn(true);

        $promoCode = $this->createMock(PromoCode::class);
        $promoCode->expects($this->once())
            ->method('getPartner')
            ->willReturn($partner);
        $promoCode->expects($this->once())
            ->method('getPartnerFeePercent')
            ->willReturn(5.00);

        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getPromoCode')
            ->willReturn($promoCode);

        $payment = $this->createMock(Payment::class);
        $payment
            ->expects($this->once())
            ->method('getUser')
            ->willReturn($user);

        $payment
            ->expects($this->once())
            ->method('getAmount')
            ->willReturn(100.00);

        $event = $this->createMock(PaymentCompletedEvent::class);
        $event
            ->expects($this->once())
            ->method('getPayment')
            ->willReturn($payment);

        $this->subscriber->onPaymentCompleted($event);
    }
}
