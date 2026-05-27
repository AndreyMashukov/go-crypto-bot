<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\OxaPayContext\Service;

use Bundles\OxaPayContext\Entity\Payment;
use Bundles\OxaPayContext\Model\OxaCallback;
use Bundles\OxaPayContext\Repository\PaymentRepository;
use Bundles\TgBotContext\Service\AlertService;
use Bundles\UserContext\Entity\User;
use Bundles\UserContext\Event\BudgetPurchaseEvent;
use Bundles\UserContext\Event\PaymentCompletedEvent;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Component\EventDispatcher\EventDispatcherInterface;

class PaymentManager
{
    private PaymentRepository $repository;

    private OxaPayClient $oxaPayClient;

    private EventDispatcherInterface $eventDispatcher;

    private EntityManagerInterface $entityManager;

    private AlertService $alertService;

    public function __construct(
        PaymentRepository $repository,
        OxaPayClient $oxaPayClient,
        EventDispatcherInterface $eventDispatcher,
        EntityManagerInterface $entityManager,
        AlertService $alertService
    ) {
        $this->repository      = $repository;
        $this->oxaPayClient    = $oxaPayClient;
        $this->eventDispatcher = $eventDispatcher;
        $this->entityManager   = $entityManager;
        $this->alertService    = $alertService;
    }

    public function validateCallback(OxaCallback $callback): void
    {
        $payment = $this->repository->findOneBy([
            'orderId' => $callback->getOrderId(),
        ]);

        if (!$payment instanceof Payment) {
            throw new \BadMethodCallException('Payment is not found.');
        }

        if (OxaPayClient::OXA_PAY_STATUS_EXPIRED === $callback->getStatus()) {
            $payment->setStatus(Payment::STATUS_EXPIRED);
            $this->repository->add($payment, true);

            return;
        }

        if (OxaPayClient::OXA_PAY_STATUS_PAID !== $callback->getStatus()) {
            return;
        }

        $payment->setEmail($callback->getEmail());
        $payment->setCompletedAt(new \DateTimeImmutable('now'));
        $payment->setStatus(Payment::STATUS_PAID);

        $this->eventDispatcher->dispatch(new BudgetPurchaseEvent($payment->getUser(), $callback->getPrice()));

        $this->entityManager->wrapInTransaction(function () use ($payment) {
            $this->eventDispatcher->dispatch(new PaymentCompletedEvent($payment));
            $this->repository->add($payment, true);

            $this->alertService->alert("PaymentManager: invoice #{$payment->getId()} ({$payment->getAmount()}$) from User {$payment->getUser()->getEmail()} is paid");
        });
        // todo: success payment notification
    }

    public function getBudgetPaymentLink(User $user, float $amount): Payment
    {
        // todo: get last not expired payment for this payment type
        $payment = new Payment(
            $user,
            $amount,
            'Budget recharge'
        );
        $this->oxaPayClient->createInvoice($payment);

        $this->repository->add($payment, true);
        $this->alertService->alert("PaymentManager: new invoice #{$payment->getId()} ({$payment->getAmount()}$) from User {$user->getEmail()}");

        return $payment;
    }
}
