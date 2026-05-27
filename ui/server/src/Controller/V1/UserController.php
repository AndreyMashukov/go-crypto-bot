<?php
namespace App\Controller\V1;

use Bundles\UserContext\Entity\User;
use FOS\RestBundle\Controller\AbstractFOSRestController;

class UserController extends AbstractFOSRestController
{
    public function getMe(): User
    {
        return $this->getUser();
    }
}
