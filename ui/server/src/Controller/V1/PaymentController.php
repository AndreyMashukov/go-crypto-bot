<?php
namespace App\Controller\V1;

use Bundles\OxaPayContext\Entity\Payment;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use Symfony\Component\HttpKernel\Exception\AccessDeniedHttpException;

class PaymentController extends AbstractFOSRestController
{
    /**
     * @Rest\Route("/{payment}", methods={"GET"}, name="get")
     * @Rest\View(serializerGroups={"payment"})
     *
     *
     */
    public function getAction(Payment $payment): Payment
    {
        $user = $this->getUser();

        if (!$payment->isOwnedBy($user)) {
            throw new AccessDeniedHttpException('Forbidden.');
        }

        return $payment;
    }
}
