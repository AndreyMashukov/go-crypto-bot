<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Service;

use Bundles\OxaPayContext\Entity\PromoCode;
use Bundles\TgBotContext\Service\AlertService;
use Bundles\UserContext\Entity\User;
use Bundles\UserContext\Event\BudgetPurchaseEvent;
use Bundles\UserContext\Model\Register;
use Bundles\UserContext\Model\User as UserAlias;
use Doctrine\ORM\EntityManager;
use Doctrine\ORM\Exception\ORMException;
use Doctrine\Persistence\ManagerRegistry;
use Symfony\Component\EventDispatcher\EventDispatcherInterface;
use Symfony\Component\PasswordHasher\Hasher\UserPasswordHasherInterface;
use Symfony\Component\Validator\Validator\ValidatorInterface;

class RegistrationService
{
    private UserPasswordHasherInterface $passwordEncoder;

    private ManagerRegistry $registry;

    private ValidatorInterface $validator;

    private EventDispatcherInterface $eventDispatcher;

    private AlertService $alertService;

    public function __construct(
        ManagerRegistry $registry,
        UserPasswordHasherInterface $passwordEncoder,
        ValidatorInterface $validator,
        EventDispatcherInterface $eventDispatcher,
        AlertService $alertService
    ) {
        $this->passwordEncoder = $passwordEncoder;
        $this->registry        = $registry;
        $this->validator       = $validator;
        $this->eventDispatcher = $eventDispatcher;
        $this->alertService    = $alertService;
    }

    /**
     * @param Register $register
     *
     * @throws ORMException
     */
    public function registerByEmail(Register $register): array
    {
        $user     = $this->findUser(trim($register->getEmail()));
        $locale   = $register->getLocale();
        $nickname = $register->getNickname();

        if (\in_array($locale, ['ru', 'en', 'id'], true)) {
            $user->setLanguage($locale);
        }

        $code = mt_rand(100000, 999999);
        $user->setPassword($this->passwordEncoder->hashPassword($user, $code));
        $user->setPlainPassword($code);

        if (!$user->getNickname()) {
            $user->setNickname($nickname);
        }

        if (!$user->getId()) {
            $violations = $this->validator->validate($user);
            if ($violations->count() > 0) {
                $message = sprintf('[%s] %s', $violations->get(0)->getPropertyPath(), $violations->get(0)->getMessage());

                throw new \BadMethodCallException($message);
            }

            $promoCode = $register->getPromoCode();

            if ($promoCode instanceof PromoCode && $promoCode->canUse($user)) {
                $user->setPromoCode($promoCode);

                $promoCode->setActivationCount($promoCode->getActivationCount() + 1);

                if ($complimentaryBudget = $promoCode->getComplimentaryBudget()) {
                    $this->eventDispatcher->dispatch(new BudgetPurchaseEvent($user, $complimentaryBudget));
                }
                if ($complimentarySignalsDays = $promoCode->getComplimentarySignalsDays()) {
                    $user->setSignalSubscriptionExpiresAt(new \DateTimeImmutable("+{$complimentarySignalsDays} days"));
                }

                $this->alertService->alert("RegistrationService: New user {$user->getEmail()}, with promocode: {$promoCode->getCode()}");
            } else {
                $this->alertService->alert("RegistrationService: New user {$user->getEmail()}");
            }
        }

        /** @var EntityManager $manager */
        $manager = $this->registry->getManager();
        $manager->flush();

        return [$code, $user];
    }

    /**
     * @param string $email
     *
     * @throws ORMException
     *
     * @return User
     */
    private function findUser(string $email): User
    {
        /** @var EntityManager $manager */
        $manager = $this->registry
            ->getManager();

        $user = $manager->getRepository(User::class)->createQueryBuilder('user')
            ->andWhere('user.email = :email')
            ->setParameter('email', $email)
            ->setMaxResults(1)
            ->getQuery()
            ->getOneOrNullResult()
        ;

        if ($user instanceof User) {
            return $user;
        }

        $user = (new User())
            ->setRoles([UserAlias::ROLE_DEFAULT])
            ->setEmail(null)
            ->setProfile(null)
            ->setUsername($email)
            ->setEmail($email);

        $manager->persist($user);

        return $user;
    }
}
