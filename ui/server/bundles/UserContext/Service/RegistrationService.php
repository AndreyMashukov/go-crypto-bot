<?php
namespace Bundles\UserContext\Service;

use Bundles\OxaPayContext\Entity\PromoCode;
use Bundles\TgBotContext\Service\AlertService;
use Bundles\UserContext\Entity\User;
use Bundles\UserContext\Event\BudgetPurchaseEvent;
use Bundles\UserContext\Model\Register;
use Bundles\UserContext\Model\User as UserAlias;
use Doctrine\ORM\Exception\ORMException;
use Doctrine\Persistence\ManagerRegistry;
use Symfony\Component\EventDispatcher\EventDispatcherInterface;
use Symfony\Component\PasswordHasher\Hasher\UserPasswordHasherInterface;
use Symfony\Component\Validator\Validator\ValidatorInterface;

class RegistrationService
{
    public function __construct(private readonly ManagerRegistry $registry, private readonly UserPasswordHasherInterface $passwordEncoder, private readonly ValidatorInterface $validator, private readonly EventDispatcherInterface $eventDispatcher, private readonly AlertService $alertService)
    {
    }

    /**
     * @throws ORMException
     */
    public function registerByEmail(Register $register): array
    {
        $user     = $this->findUser(trim((string) $register->getEmail()));
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

        if ($user->getId() === 0) {
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

        $manager = $this->registry->getManager();
        $manager->flush();

        return [$code, $user];
    }

    /**
     * @throws ORMException
     */
    private function findUser(string $email): User
    {
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
