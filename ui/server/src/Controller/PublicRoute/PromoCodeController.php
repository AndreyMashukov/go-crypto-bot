<?php
namespace App\Controller\PublicRoute;

use Bundles\OxaPayContext\Entity\PromoCode;
use Bundles\OxaPayContext\Repository\PromoCodeRepository;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;
use Symfony\Component\HttpKernel\Exception\NotFoundHttpException;

class PromoCodeController extends AbstractFOSRestController
{
    public function __construct(private readonly PromoCodeRepository $repository)
    {
    }

    public function getTest(string $code): PromoCode
    {
        $promoCode = $this->repository->findOneBy([
            'code' => $code,
        ]);

        if (!$promoCode instanceof PromoCode) {
            throw new NotFoundHttpException("PromoCode {$code} is not found.");
        }

        $user = $this->getUser();

        if (!$promoCode->canUse($user)) {
            throw new BadRequestHttpException("PromoCode {$code} can not be used.");
        }

        return $promoCode;
    }
}
