<?php

declare(strict_types=1);

namespace Bundles\UserContext\Service;

use Bundles\UserContext\Entity\User;
use Flagception\Decorator\ContextDecoratorInterface;
use Flagception\Exception\AlreadyDefinedException;
use Flagception\Model\Context;

class UserContextDecorator implements ContextDecoratorInterface
{
    /**
     * ClientContextDecorator constructor.
     */
    public function __construct(private readonly UserProvider $userProvider)
    {
    }

    public function getName(): string
    {
        return 'user_context_decorator';
    }

    /**
     *
     * @throws AlreadyDefinedException
     *
     */
    public function decorate(Context $context): Context
    {
        $user = $this->userProvider->getCurrentUser();

        if ($user instanceof User) {
            $context->add('user_id', $user->getId());

            return $context;
        }

        $context->add('user_id', null);

        return $context;
    }
}
