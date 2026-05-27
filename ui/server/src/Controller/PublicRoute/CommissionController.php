<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Controller\PublicRoute;

use Bundles\OxaPayContext\Service\CommissionService;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;

/**
 * @Rest\Route("/public/commission", name="public_commission_")
 */
class CommissionController extends AbstractFOSRestController
{
    private CommissionService $commissionService;

    public function __construct(CommissionService $commissionService)
    {
        $this->commissionService = $commissionService;
    }

    /**
     * @Rest\Route("/info", methods={"GET"}, name="info")
     * @Rest\View
     *
     * @return array
     */
    public function getInfoAction(): array
    {
        return $this->commissionService->getCommissionConfigurationList();
    }
}
