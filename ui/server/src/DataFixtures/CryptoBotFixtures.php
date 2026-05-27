<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\DataFixtures;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Entity\Server;
use Bundles\UserContext\Entity\User;
use Doctrine\Bundle\FixturesBundle\Fixture;
use Doctrine\Bundle\FixturesBundle\FixtureGroupInterface;
use Doctrine\Common\DataFixtures\DependentFixtureInterface;
use Doctrine\Persistence\ObjectManager;

class CryptoBotFixtures extends Fixture implements FixtureGroupInterface, DependentFixtureInterface
{
    public const TEST_CRYPTO_BOT_UUID = '87bf6369-3a27-4d2f-b26c-f90aed0c955b';

    public function load(ObjectManager $manager): void
    {
        /** @var User $user */
        $user = $this->getReference(OAuthFixtures::USER_REFERENCE);
        /** @var Server $server */
        $server = $this->getReference(ServerFixtures::SERVER_REFERENCE_1);

        $user->setSignalSubscriptionExpiresAt(new \DateTimeImmutable('+2 hours'));
        $cryptoBot = (new CryptoBot($user))
            ->setStatus(CryptoBot::STATUS_RUNNING)
            ->setPort('8000')
            ->setServer($server)
            ->setContainerId('xxxxxxxx')
            ->setApiKey('test')
            ->setApiSecret('test')
            ->setProvider('binance')
            ->setUuid(self::TEST_CRYPTO_BOT_UUID);

        $config = new CryptoTradeConfig();
        $config->setCryptobot($cryptoBot)
            ->setSymbol('PERPUSDT')
            ->setEnabled(true)
            ->setSignalTrading(true)
            ->setUsdtLimit(100);
        $cryptoBot->addCryptoTradeConfig($config);
        $manager->persist($config);
        $manager->persist($cryptoBot);
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
            ServerFixtures::class,
        ];
    }
}
