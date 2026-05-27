<?php

declare(strict_types=1);

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
    public function __construct(private readonly UserProvider $userProvider, private readonly EntityManagerInterface $entityManager, private readonly IpExtractor $ipExtractor)
    {
    }

    public static function getSubscribedEvents(): array
    {
        return [
            KernelEvents::REQUEST => 'onKernelRequest',
        ];
    }

    public function onKernelRequest(RequestEvent $event)
    {
        $request = $event->getRequest();
        $userIp  = $this->ipExtractor->extractIp($request);

        $user = $this->userProvider->getCurrentUser();
        if ($user instanceof User) {
            $user->setLastLogin(new \DateTime('now'));
            if ($userIp !== '' && $userIp !== '0') {
                $user->setLastLoginIp($userIp);
            }
            $this->entityManager->flush();
        }
    }
}
