<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\EventListener;

use Bundles\UserContext\Entity\User;
use Bundles\UserContext\Exception\TooManyAuthAttemptsFailed;
use Bundles\UserContext\Service\BruteForceSecurity;
use Doctrine\Persistence\ManagerRegistry;
use Symfony\Component\EventDispatcher\EventSubscriberInterface;
use Symfony\Component\HttpKernel\Exception\BadRequestHttpException;
use Symfony\Component\HttpKernel\Exception\NotFoundHttpException;
use Symfony\Component\HttpKernel\Exception\TooManyRequestsHttpException;
use Symfony\Component\Security\Core\Encoder\UserPasswordEncoderInterface;
use Symfony\Component\Security\Core\Exception\UserNotFoundException;
use Symfony\Component\Security\Core\User\UserProviderInterface;
use Trikoder\Bundle\OAuth2Bundle\Event\UserResolveEvent;
use Trikoder\Bundle\OAuth2Bundle\OAuth2Events;

class UserListener implements EventSubscriberInterface
{
    private UserPasswordEncoderInterface $userPasswordEncoder;

    private UserProviderInterface $userProvider;

    private ManagerRegistry $registry;

    private BruteForceSecurity $bruteForceSecurity;

    public function __construct(
        UserPasswordEncoderInterface $userPasswordEncoder,
        UserProviderInterface $userProvider,
        ManagerRegistry $registry,
        BruteForceSecurity $bruteForceSecurity
    ) {
        $this->userPasswordEncoder = $userPasswordEncoder;
        $this->userProvider        = $userProvider;
        $this->registry            = $registry;
        $this->bruteForceSecurity  = $bruteForceSecurity;
    }

    public static function getSubscribedEvents()
    {
        return [
            OAuth2Events::USER_RESOLVE => 'userResolve',
        ];
    }

    public function userResolve(UserResolveEvent $event): void
    {
        $grant = $event->getGrant();

        if ('password' !== ((string) $grant)) {
            throw new BadRequestHttpException('Wrong Grant-Type.');
        }

        $username = $event->getUsername();
        $password = $event->getPassword();

        try {
            $user = $this->userProvider->loadUserByUsername($username);
        } catch (UserNotFoundException $exception) {
            throw new NotFoundHttpException('User is not found.', $exception);
        }

        if (!$user instanceof User) {
            throw new NotFoundHttpException('User is not found.');
        }

        if (!$user->isEnabled()) {
            throw new NotFoundHttpException('User is deactivated.');
        }

        if ($this->bruteForceSecurity->isBlocked($user)) {
            throw new TooManyRequestsHttpException(3600, 'Too many requests, please try later.');
        }

        if (!$this->userPasswordEncoder->isPasswordValid($user, $password)) {
            try {
                $this->bruteForceSecurity->trackFail($user);
            } catch (TooManyAuthAttemptsFailed $exception) {
                throw new TooManyRequestsHttpException(3600, $exception->getMessage(), $exception);
            }

            throw new BadRequestHttpException('Wrong password.');
        }

        $user->setLastLogin(new \DateTime('now'));
        $this->registry->getManager()->flush();

        // Todo This is temporary solution. Because of... look at AccessTokenTrait
        $_SERVER['user_claim'] = [
            'id'       => $user->getId(),
            'username' => $user->getUsername(),
            'roles'    => $user->getRoles(),
        ];

        $event->setUser($user);
    }
}
