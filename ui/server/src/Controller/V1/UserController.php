<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Controller\V1;

use Bundles\UserContext\Entity\User;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;

/**
 * @Rest\Route("/v1/user", name="v1_user_")
 */
class UserController extends AbstractFOSRestController
{
    /**
     * @Rest\Route("/me", name="get", methods={"GET"})
     * @Rest\View(serializerGroups={"user_public", "user_extended", "subscription"})
     *
     * @return User
     */
    public function getMeAction(): User
    {
        /** @var User $user */
        $user = $this->getUser();

        return $user;
    }
}
