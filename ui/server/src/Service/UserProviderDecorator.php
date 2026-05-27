<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Service;

use Bundles\UserContext\Entity\User;
use Doctrine\Persistence\ManagerRegistry;
use Symfony\Component\Security\Core\Exception\UserNotFoundException;
use Symfony\Component\Security\Core\User\UserInterface;
use Symfony\Component\Security\Core\User\UserProviderInterface;

class UserProviderDecorator implements UserProviderInterface
{
    private UserProviderInterface $userProvider;

    private ManagerRegistry $registry;

    public function __construct(
        UserProviderInterface $userProvider,
        ManagerRegistry $registry
    ) {
        $this->userProvider = $userProvider;
        $this->registry     = $registry;
    }

    public function refreshUser(UserInterface $user)
    {
        return $this->userProvider->refreshUser($user);
    }

    public function supportsClass(string $class)
    {
        return $this->userProvider->supportsClass($class);
    }

    public function loadUserByUsername(string $username)
    {
        try {
            return $this->userProvider->loadUserByUsername($username);
        } catch (UserNotFoundException $exception) {
            $user = $this->registry->getRepository(User::class)->findOneBy([
                'phone' => $username,
            ]);

            if ($user instanceof User) {
                return $user;
            }

            $user = $this->registry->getRepository(User::class)->findOneBy([
                'email' => $username,
            ]);

            if ($user instanceof User) {
                return $user;
            }

            throw $exception;
        }
    }

    public function loadUserByIdentifier(string $identifier)
    {
        return $this->userProvider->loadUserByIdentifier($identifier);
    }
}
