<?php
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
    public function __construct(private readonly PaymentRepository $repository, private readonly OxaPayClient $oxaPayClient, private readonly EventDispatcherInterface $eventDispatcher, private readonly EntityManagerInterface $entityManager, private readonly AlertService $alertService)
    {
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

    }

    public function getBudgetPaymentLink(User $user, float $amount): Payment
    {
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
