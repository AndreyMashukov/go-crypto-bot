<?php
namespace App\Controller\V1;

use Bundles\OxaPayContext\Form\ServicePurchaseType;
use Bundles\OxaPayContext\Model\PaidService;
use Bundles\OxaPayContext\Model\ServicePurchase;
use Bundles\OxaPayContext\Service\PaidServiceManager;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;

class ServiceController extends AbstractFOSRestController
{
    public function __construct(private readonly PaidServiceManager $paidServiceManager)
    {
    }

    /**
     * @return PaidService[]
     */
    public function getList(): array
    {
        $user = $this->getUser();

        return $this->paidServiceManager->getServiceList($user);
    }

    public function post(Request $request)
    {
        $user = $this->getUser();
        $form = $this->createForm(ServicePurchaseType::class, new ServicePurchase(), [
            'method' => Request::METHOD_POST,
        ])->handleRequest($request);

        if (!$form->isSubmitted()) {
            $form->submit([]);
        }

        if (!$form->isValid()) {
            return [
                'form' => $form,
            ];
        }

        $data = $form->getData();

        try {
            $service = $this->paidServiceManager->getPaidService($data->code, $user);
            $this->paidServiceManager->purchase($user, $service);
        } catch (\BadMethodCallException $exception) {
            throw new BadRequestHttpException($exception->getMessage(), $exception);
        }
        return null;
    }
}
