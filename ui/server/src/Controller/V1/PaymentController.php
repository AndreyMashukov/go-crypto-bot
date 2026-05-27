<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Controller\V1;

use Bundles\OxaPayContext\Entity\Payment;
use Bundles\UserContext\Entity\User;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use Symfony\Component\HttpKernel\Exception\AccessDeniedHttpException;

/**
 * @Rest\Route("/v1/payment", name="v1_payment_")
 */
class PaymentController extends AbstractFOSRestController
{
    /**
     * @Rest\Route("/{payment}", methods={"GET"}, name="get")
     * @Rest\View(serializerGroups={"payment"})
     *
     * @param Payment $payment
     *
     * @return Payment
     */
    public function getAction(Payment $payment): Payment
    {
        /** @var User $user */
        $user = $this->getUser();

        if (!$payment->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        return $payment;
    }
}
