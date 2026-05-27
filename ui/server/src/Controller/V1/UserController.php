<?php
namespace App\Controller\V1;

use Bundles\UserContext\Entity\User;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;

class UserController extends AbstractFOSRestController
{
    /**
     * @Rest\Route("/me", name="get", methods={"GET"})
     * @Rest\View(serializerGroups={"user_public", "user_extended", "subscription"})
     */
    public function getMe(): User
    {
        return $this->getUser();
    }
}
