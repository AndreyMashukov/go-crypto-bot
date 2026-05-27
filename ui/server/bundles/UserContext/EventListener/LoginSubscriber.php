<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\EventListener;

use App\Service\IpExtractor;
use Bundles\UserContext\Entity\User;
use Bundles\UserContext\Service\UserProvider;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Component\EventDispatcher\EventSubscriberInterface;
use Symfony\Component\HttpKernel\Event\RequestEvent;
use Symfony\Component\HttpKernel\KernelEvents;

class LoginSubscriber implements EventSubscriberInterface
{
    private UserProvider $userProvider;

    private EntityManagerInterface $entityManager;

    private IpExtractor $ipExtractor;

    public function __construct(
        UserProvider $userProvider,
        EntityManagerInterface $entityManager,
        IpExtractor $ipExtractor
    ) {
        $this->userProvider  = $userProvider;
        $this->entityManager = $entityManager;
        $this->ipExtractor   = $ipExtractor;
    }

    /**
     * @return array
     */
    public static function getSubscribedEvents(): array
    {
        return [
            KernelEvents::REQUEST => 'onKernelRequest',
        ];
    }

    /**
     * @param RequestEvent $event
     */
    public function onKernelRequest(RequestEvent $event)
    {
        $request = $event->getRequest();
        $userIp  = $this->ipExtractor->extractIp($request);

        $user = $this->userProvider->getCurrentUser();
        if ($user instanceof User) {
            $user->setLastLogin(new \DateTime('now'));
            if ($userIp) {
                $user->setLastLoginIp($userIp);
            }
            $this->entityManager->flush();
        }
    }
}
