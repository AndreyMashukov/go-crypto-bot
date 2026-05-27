<?php
namespace Bundles\UserContext\Service;

use Symfony\Component\Security\Core\Security;
use Symfony\Component\Security\Core\User\UserInterface;
use League\Bundle\OAuth2ServerBundle\Security\Authenticator\OAuth2Token;

class UserProvider
{
    private readonly Security $security;

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
