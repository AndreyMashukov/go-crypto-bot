<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Functional\Controller\V1;

use App\Tests\RestTestCase;
use Bundles\CryptoBotContext\Entity\CryptoBot;
use Bundles\OxaPayContext\Service\PaidServiceManager;
use Bundles\UserContext\Entity\User;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * @group functional
 */
class ServiceControllerTest extends RestTestCase
{
    /**
     * Should allow to get list of service and purchase it.
     */
    public function testShouldAllowToGetListOfServiceAndPurchaseIt(): void
    {
//        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_service_get_list')));
//        $this->assertJsonSnapshot($json, 'list_1');

        $user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $this->assertInstanceOf(User::class, $user);
        $user->setSignalSubscriptionExpiresAt(null);
        $this->em->flush();
//        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_service_get_list')));
//        $this->assertJsonSnapshot($json, 'list_2');

        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_service_purchase'), Request::METHOD_POST, [
            'code' => PaidServiceManager::SIGNAL_SUBSCRIPTION_CODE,
        ]), Response::HTTP_BAD_REQUEST);
        $this->assertJsonSnapshot($json, 'no_balance');
        $user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $this->assertInstanceOf(User::class, $user);
        $user->setBudget(100);
        $this->em->flush();
        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_service_purchase'), Request::METHOD_POST, [
            'code' => PaidServiceManager::SIGNAL_SUBSCRIPTION_CODE,
        ]), Response::HTTP_NO_CONTENT);
        $this->assertNull($json);

        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_service_get_list')));
        $this->assertJsonSnapshot($json, 'final');
        $user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $this->assertInstanceOf(User::class, $user);
        $this->assertTrue($user->hasActiveSignalSubscription());
        $this->assertEquals(80, $user->getBudget());
    }

    /**
     * Should allow to pay for basic subscription.
     */
    public function testShouldAllowToPayForBasicSubscription(): void
    {
        $user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $this->assertInstanceOf(User::class, $user);
        $user->setBudget(1000);
        $user->setSignalSubscriptionExpiresAt(null);
        $this->em->flush();
        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_service_purchase'), Request::METHOD_POST, [
            'code' => PaidServiceManager::BASIC_SUBSCRIPTION_CODE,
        ]), Response::HTTP_NO_CONTENT);
        $this->assertNull($json);

        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_service_get_list')));
        $this->assertJsonSnapshot($json, 'final');

        $url  = $this->getUrl('v1_user_get');
        $json = $this->deserialize($this->apiRequest($url));
        $this->assertJsonSnapshot($json, 'user_info');

        $user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $this->assertInstanceOf(User::class, $user);
        $this->assertFalse($user->hasActiveSignalSubscription());
        $this->assertTrue($user->hasActiveBasicSubscription());
        $this->assertEquals(110, $user->getBudget());
    }

    /**
     * Should allow to pay for dedicated server.
     */
    public function testShouldAllowToPayForDedicatedServer(): void
    {
        $user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $this->assertInstanceOf(User::class, $user);
        $user->setBudget(20);
        $this->em->flush();
        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_service_purchase'), Request::METHOD_POST, [
            'code' => PaidServiceManager::DEDICATED_SERVER_BINANCE_CODE,
        ]), Response::HTTP_NO_CONTENT);
        $this->assertNull($json);

        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_service_get_list')));
        $this->assertJsonSnapshot($json, 'final');
        $deploy = $this->deserialize($this->apiRequest($this->getUrl('v1_cryptobot_get', [
            'cryptobot' => 1,
        ])));
        $this->assertJsonSnapshot($deploy, 'bot');
        $user = $this->em->getRepository(User::class)->findOneBy([
            'username' => $this->username,
        ]);
        $this->assertInstanceOf(User::class, $user);
        $this->assertTrue($user->hasActiveSignalSubscription());
        $this->assertFalse($user->hasActiveBasicSubscription());
        $this->assertEquals(5, $user->getBudget());
        $binanceBot = $user->getBinanceBot();
        $this->assertInstanceOf(CryptoBot::class, $binanceBot);
        $this->assertTrue($binanceBot->hasDedicatedServer());
        $this->assertTrue($binanceBot->hasActiveDedicatedServerSubscription());
        $this->assertInstanceOf(\DateTimeImmutable::class, $binanceBot->getDedicatedServerExpiresAt());
    }
}
