<?php

declare(strict_types=1);

namespace App\Controller\PublicRoute;

use Bundles\OxaPayContext\Service\CommissionService;
use FOS\RestBundle\Controller\AbstractFOSRestController;

class CommissionController extends AbstractFOSRestController
{
    public function __construct(private readonly CommissionService $commissionService)
    {
    }

    public function getInfo(): array
    {
        return $this->commissionService->getCommissionConfigurationList();
    }
}
