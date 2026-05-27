<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Controller\V1;

use Bundles\OxaPayContext\Form\ServicePurchaseType;
use Bundles\OxaPayContext\Model\PaidService;
use Bundles\OxaPayContext\Model\ServicePurchase;
use Bundles\OxaPayContext\Service\PaidServiceManager;
use Bundles\UserContext\Entity\User;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;

/**
 * @Rest\Route("/v1/service", name="v1_service_")
 */
class ServiceController extends AbstractFOSRestController
{
    private PaidServiceManager $paidServiceManager;

    public function __construct(PaidServiceManager $paidServiceManager)
    {
        $this->paidServiceManager = $paidServiceManager;
    }

    /**
     * @Rest\Route("/list", methods={"GET"}, name="get_list")
     * @Rest\View(serializerGroups={"paid_service"})
     *
     * @return PaidService[]
     */
    public function getListAction(): array
    {
        /** @var User $user */
        $user = $this->getUser();

        return $this->paidServiceManager->getServiceList($user);
    }

    /**
     * @Rest\Route("/purchase", name="purchase", methods={"POST"})
     * @Rest\View
     *
     * @param Request $request
     */
    public function postAction(Request $request)
    {
        /** @var User $user */
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

        /** @var ServicePurchase $data */
        $data = $form->getData();

        try {
            $service = $this->paidServiceManager->getPaidService($data->code, $user);
            $this->paidServiceManager->purchase($user, $service);
        } catch (\BadMethodCallException $exception) {
            throw new BadRequestHttpException($exception->getMessage(), $exception);
        }
    }
}
