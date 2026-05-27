<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Service;

use Symfony\Component\Security\Core\Security;
use Symfony\Component\Security\Core\User\UserInterface;
use Trikoder\Bundle\OAuth2Bundle\Security\Authentication\Token\OAuth2Token;

class UserProvider
{
    private Security $security;

    public function __construct(Security $security)
    {
        $this->security = $security;
    }

    public function getCurrentUser(): ?UserInterface
    {
        $token = $this->security->getToken();

        if (!$token instanceof OAuth2Token) {
            return null;
        }

        $user = $token->getUser();

        if (!$user instanceof UserInterface) {
            return null;
        }

        return $user;
    }
}
