<?php
namespace App\Controller\V1;

use Bundles\OxaPayContext\Entity\Payment;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use Symfony\Component\HttpKernel\Exception\AccessDeniedHttpException;

class PaymentController extends AbstractFOSRestController
{
    public function getAction(Payment $payment): Payment
    {
        $user = $this->getUser();

        if (!$payment->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        return $payment;
    }
}
