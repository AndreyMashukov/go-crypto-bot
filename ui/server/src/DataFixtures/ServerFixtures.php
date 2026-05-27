<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\DataFixtures;

use Bundles\CryptoBotContext\Entity\Server;
use Doctrine\Bundle\FixturesBundle\Fixture;
use Doctrine\Bundle\FixturesBundle\FixtureGroupInterface;
use Doctrine\Persistence\ObjectManager;

class ServerFixtures extends Fixture implements FixtureGroupInterface
{
    public const SERVER_REFERENCE_1 = 'server_1_reference';

    public const SERVER_LIST = [
        [
            'ip'        => '127.0.0.1',
            'slots'     => 2,
            'cpu'       => 4,
            'ram'       => 2,
            'master'    => false,
            'reference' => self::SERVER_REFERENCE_1,
        ],
    ];

    public function load(ObjectManager $manager)
    {
        foreach (self::SERVER_LIST as $server) {
            $serverEntity = (new Server())
                ->setIp($server['ip'])
                ->setSlots($server['slots'])
                ->setCpu($server['cpu'])
                ->setRam($server['ram'])
                ->setMaster($server['master'])
            ;

            $manager->persist($serverEntity);
            $this->addReference($server['reference'], $serverEntity);
        }

        $manager->flush();
    }

    public static function getGroups(): array
    {
        return ['first'];
    }
}
