<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Controller\V1;

use Bundles\OxaPayContext\Entity\Payment;
use Bundles\OxaPayContext\Service\PaymentManager;
use Bundles\UserContext\Entity\User;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;

/**
 * @Rest\Route("/v1/subscription", name="v1_subscription_")
 */
class SubscriptionController extends AbstractFOSRestController
{
    private PaymentManager $paymentManager;

    public function __construct(PaymentManager $paymentManager)
    {
        $this->paymentManager = $paymentManager;
    }

    /**
     * @Rest\Route("/budget", methods={"POST"}, name="budget")
     * @Rest\View(serializerGroups={"payment_short"})
     *
     * @param Request $request
     *
     * @return Payment
     */
    public function postBudgetAction(Request $request): Payment
    {
        /** @var User $user */
        $user = $this->getUser();

        $amount        = (int) $request->get('amount');
        $allowedAmount = [100, 250, 500, 1000];
        if (!\in_array($amount, $allowedAmount, true)) {
            throw new BadRequestHttpException('Amount is out of range.');
        }

        return $this->paymentManager->getBudgetPaymentLink($user, $amount);
    }
}
