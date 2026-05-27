<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\DataFixtures;

use Bundles\OxaPayContext\Entity\PromoCode;
use Bundles\UserContext\Entity\User;
use Doctrine\Bundle\FixturesBundle\Fixture;
use Doctrine\Bundle\FixturesBundle\FixtureGroupInterface;
use Doctrine\Common\DataFixtures\DependentFixtureInterface;
use Doctrine\Persistence\ObjectManager;

class SubscriptionFixtures extends Fixture implements FixtureGroupInterface, DependentFixtureInterface
{
    public function load(ObjectManager $manager)
    {
        /** @var User $user */
        $user = $this->getReference(OAuthFixtures::USER_REFERENCE);

        $promoCode = new PromoCode($user, 'TEST123');
        $promoCode->setActivationCount(0);
        $promoCode->setMaxActivationLimit(10);
        $promoCode->setComplimentaryBudget(10);
        $promoCode->setComplimentarySignalsDays(7);
        $manager->persist($promoCode);

        $manager->flush();
    }

    public static function getGroups(): array
    {
        return ['first'];
    }

    public function getDependencies()
    {
        return [
            OAuthFixtures::class,
        ];
    }
}
