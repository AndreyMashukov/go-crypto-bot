<?php

declare(strict_types=1);

namespace Bundles\UserContext\Service;

use Symfony\Bundle\SecurityBundle\Security;
use Symfony\Component\Security\Core\User\UserInterface;

class UserProvider
{
    public function __construct(private readonly Security $security)
    {
    }

    public function getCurrentUser(): ?UserInterface
    {
        $user = $this->security->getUser();

        return $user instanceof UserInterface ? $user : null;
    }
}
