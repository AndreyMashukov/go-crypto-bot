<?php
namespace App\Tests\Functional\Controller\V1;

use App\Tests\RestTestCase;
use Bundles\OxaPayContext\Entity\Payment;
use Bundles\UserContext\Entity\User;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

class SubscriptionControllerTest extends RestTestCase
{
    public function testShouldAllowToPayBudgetByOxa(): void
    {
        $url  = $this->getUrl('v1_subscription_budget');
        $json = $this->deserialize($this->apiRequest($url, Request::METHOD_POST, [
            'amount' => 100,
        ]));
        $this->assertArrayHasKey('id', $json);
        $payment = $this->em->find(Payment::class, $json['id']);
        $this->assertEquals(0, $payment->getUser()->getBudget());
        $this->assertInstanceOf(Payment::class, $payment);
        $this->assertEquals(Payment::STATUS_PENDING, $payment->getStatus());
        $this->assertInstanceOf(\DateTimeImmutable::class, $payment->getExpiresAt());
        $this->assertMatchesRegularExpression('/^https:\/\/oxapay\.com\/mpay\/\d+$/ui', $payment->getPaymentLink());
        $this->assertEquals($json['paymentLink'], $payment->getPaymentLink());
        $this->assertEquals(100, $payment->getAmount());
        $this->assertEquals('USDT', $payment->getCurrency());

        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_payment_get', [
            'payment' => $payment->getId(),
        ])));
        $this->assertJsonSnapshot($json, 'pending');

        $url      = $this->getUrl('public_callback_oxa');
        $callback = json_decode(file_get_contents(__DIR__ . '/dataset/oxa_callback.json'), true);

        $callback['orderId'] = $payment->getOrderId();

        $response = $this->apiPublicRequest($url, Request::METHOD_POST, $callback);
        $this->assertEquals(Response::HTTP_OK, $response->getStatusCode());
        $this->assertEquals('OK', $response->getContent());

        $json = $this->deserialize($this->apiRequest($this->getUrl('v1_payment_get', [
            'payment' => $payment->getId(),
        ])));
        $this->assertJsonSnapshot($json, 'paid');

        $user = $this->em->find(User::class, $payment->getUser()->getId());
        $this->assertFalse($user->hasActiveBasicSubscription());
        $this->assertEquals(50, $user->getBudget());

        $url  = $this->getUrl('v1_user_get');
        $json = $this->deserialize($this->apiRequest($url));
        $this->assertJsonSnapshot($json);
    }
}
