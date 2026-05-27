<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Controller\PublicRoute;

use Bundles\OxaPayContext\Entity\PromoCode;
use Bundles\OxaPayContext\Repository\PromoCodeRepository;
use Bundles\UserContext\Entity\User;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;
use Symfony\Component\HttpKernel\Exception\NotFoundHttpException;

/**
 * @Rest\Route("/public/promocode", name="public_promocode_")
 */
class PromoCodeController extends AbstractFOSRestController
{
    private PromoCodeRepository $repository;

    public function __construct(PromoCodeRepository $repository)
    {
        $this->repository = $repository;
    }

    /**
     * @Rest\Route("/{code}/test", methods={"GET"}, name="test")
     * @Rest\View(serializerGroups={"promocode"})
     *
     * @param string $code
     *
     * @return PromoCode
     */
    public function getTestAction(string $code): PromoCode
    {
        $promoCode = $this->repository->findOneBy([
            'code' => $code,
        ]);

        if (!$promoCode instanceof PromoCode) {
            throw new NotFoundHttpException("PromoCode {$code} is not found.");
        }

        /** @var null|User $user */
        $user = $this->getUser();

        if (!$promoCode->canUse($user)) {
            throw new BadRequestHttpException("PromoCode {$code} can not be used.");
        }

        return $promoCode;
    }
}
