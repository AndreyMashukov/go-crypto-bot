<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Service;

use Bundles\UserContext\Entity\User;
use Flagception\Decorator\ContextDecoratorInterface;
use Flagception\Exception\AlreadyDefinedException;
use Flagception\Model\Context;

/**
 * Class UserContextDecorator.
 */
class UserContextDecorator implements ContextDecoratorInterface
{
    private UserProvider $userProvider;

    /**
     * ClientContextDecorator constructor.
     *
     * @param UserProvider $userProvider
     */
    public function __construct(UserProvider $userProvider)
    {
        $this->userProvider = $userProvider;
    }

    public function getName(): string
    {
        return 'user_context_decorator';
    }

    /**
     * @param Context $context
     *
     * @throws AlreadyDefinedException
     *
     * @return Context
     */
    public function decorate(Context $context): Context
    {
        /** @var User $user */
        $user = $this->userProvider->getCurrentUser();

        if ($user instanceof User) {
            $context->add('user_id', $user->getId());

            return $context;
        }

        $context->add('user_id', null);

        return $context;
    }
}
