<?php

declare(strict_types=1);

namespace App\Controller\PublicRoute;

use Bundles\OxaPayContext\Service\CommissionService;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;

class CommissionController extends AbstractFOSRestController
{
    public function __construct(private readonly CommissionService $commissionService)
    {
    }

    /**
     * @Rest\Route("/info", methods={"GET"}, name="info")
     * @Rest\View
     */
    public function getInfo(): array
    {
        return $this->commissionService->getCommissionConfigurationList();
    }
}
