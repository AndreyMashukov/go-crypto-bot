<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\DataFixtures;

use Bundles\UserContext\Entity\User;
use Doctrine\Bundle\FixturesBundle\Fixture;
use Doctrine\Bundle\FixturesBundle\FixtureGroupInterface;
use Doctrine\Persistence\ObjectManager;
use Exception;
use Symfony\Component\PasswordHasher\Hasher\UserPasswordHasherInterface;
use League\Bundle\OAuth2ServerBundle\Manager\ClientManagerInterface;
use League\Bundle\OAuth2ServerBundle\Model\Client as ClientModel;
use League\Bundle\OAuth2ServerBundle\Model\Grant;
use League\Bundle\OAuth2ServerBundle\Model\Scope;

/**
 * @SuppressWarnings(PHPMD)
 */
class OAuthFixtures extends Fixture implements FixtureGroupInterface
{
    public const USER_REFERENCE = 'test_user_1';

    private UserPasswordHasherInterface $passwordEncoder;

    private ClientManagerInterface $clientManager;

    public function __construct(UserPasswordHasherInterface $passwordEncoder, ClientManagerInterface $clientManager)
    {
        $this->passwordEncoder = $passwordEncoder;
        $this->clientManager   = $clientManager;
    }

    /**
     * @throws Exception
     */
    public function load(ObjectManager $manager): void
    {
        /** @var User $user */
        $user = (new User())
            ->setPhone('6285940749131')
            ->setFirstName('John')
            ->setNickname('trader99')
            ->setUsername('john')
            ->setRoles([User::ROLE_DEFAULT, User::ROLE_ADMIN])
            ->setEmail('john@batur.loc')
        ;

        $user->setPassword($this->passwordEncoder->hashPassword($user, '112233'));
        $manager->persist($user);

        $this->setReference('user', $user);

        // OAuth

        $secret = '7d094bf4175b0a95890b30a8c260597449b086aac70729444d72a4b2d11f3ee0ba05356ee4e63bd28f26f8f63ae40c685f6e0ae9512b38902c63e652b1c6621c';
        $client = new ClientModel(
            'c3ff36379fd0aff317297ed1d1b45b80',
            $secret
        );

        $client->setScopes(...[new Scope('user')]);
        $client->setActive(1);
        $client->setGrants(...[new Grant('password'), new Grant('refresh_token')]);

        $this->clientManager->save($client);

        $manager->flush();

        $this->addReference(self::USER_REFERENCE, $user);
    }

    public static function getGroups(): array
    {
        return ['first'];
    }
}
