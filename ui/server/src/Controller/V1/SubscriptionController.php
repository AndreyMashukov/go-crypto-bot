<?php
namespace App\Controller\V1;

use Bundles\OxaPayContext\Entity\Payment;
use Bundles\OxaPayContext\Service\PaymentManager;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;

class SubscriptionController extends AbstractFOSRestController
{
    public function __construct(private readonly PaymentManager $paymentManager)
    {
    }

    /**
     * @Rest\Route("/budget", methods={"POST"}, name="budget")
     * @Rest\View(serializerGroups={"payment_short"})
     *
     *
     */
    public function postBudget(Request $request): Payment
    {
        $user = $this->getUser();

        $amount        = (int) $request->get('amount');
        $allowedAmount = [100, 250, 500, 1000];
        if (!\in_array($amount, $allowedAmount, true)) {
            throw new BadRequestHttpException('Amount is out of range.');
        }

        return $this->paymentManager->getBudgetPaymentLink($user, $amount);
    }
}
