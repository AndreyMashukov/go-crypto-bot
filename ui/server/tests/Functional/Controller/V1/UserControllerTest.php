<?php
namespace App\Tests\Functional\Controller\V1;

use App\Tests\RestTestCase;
use Bundles\UserContext\Entity\User;
use Symfony\Component\HttpFoundation\Request;

class UserControllerTest extends RestTestCase
{
    public function testShouldAllowToGetAuthorizedUserInfo(): void
    {
        $date = new \DateTimeImmutable();

        $url  = $this->getUrl('v1_user_get');
        $json = $this->deserialize($this->apiRequest($url));
        $this->assertJsonSnapshot($json);

        $user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $this->assertInstanceOf(User::class, $user);
        $this->assertEquals('127.0.0.1', $user->getLastLoginIp());
        $this->assertGreaterThanOrEqual($date->getTimestamp(), $user->getLastLogin()->getTimestamp());

        $json = $this->deserialize($this->apiRequest($url, Request::METHOD_GET, [], [
            'HTTP_cf-connecting-ip' => '2a02:26f7:d709:4000:4d4d:a8cf:2a22:5549',
        ]));
        $this->assertJsonSnapshot($json);

        $user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $this->assertInstanceOf(User::class, $user);
        $this->assertEquals('2a02:26f7:d709:4000:4d4d:a8cf:2a22:5549', $user->getLastLoginIp());
    }
}
