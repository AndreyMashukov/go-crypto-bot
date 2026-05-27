<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Tests\Functional\Controller\PublicRoute;

use App\Tests\AsyncHandlerTestCase;
use Bundles\OxaPayContext\Entity\PromoCode;
use Bundles\UserContext\Entity\User;
use Bundles\UserContext\Model\Register;
use Lcobucci\JWT\Encoding\JoseEncoder;
use Lcobucci\JWT\Token;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * @group functional
 */
class RegistrationControllerTest extends AsyncHandlerTestCase
{
    /**
     * Should allow to register user and send code.
     */
    public function testShouldAllowToRegisterUserAndSendCode(): void
    {
        $email = 'amashukov@example.com';
        $url   = $this->getUrl('public_registration_code', [
            '_locale' => 'id',
        ]);

        $json = $this->deserialize($this->apiPublicRequest($url, Request::METHOD_POST, [
            'nickname'  => 'trader88',
            'email'     => $email,
            'secret'    => $this->getSecret($email),
            'promoCode' => 'TEST123',
        ]), Response::HTTP_NO_CONTENT);

        $this->assertNull($json);
        $this->assertCount(1, $this->emails);
        $emailMessage = $this->emails[0];
        $this->assertArrayHasKey(1, $emailMessage);
        $this->assertArrayHasKey('code', $emailMessage[1]);

        $url  = $this->getUrl('oauth2_token', []);
        $body = [
            'client_id'     => self::CLIENT_ID,
            'client_secret' => self::CLIENT_SECRET,
            'grant_type'    => 'password',
            'username'      => $email,
            'password'      => $emailMessage[1]['code'],
            'scope'         => $this->scopes,
        ];

        $json = $this->deserialize($this->apiPublicRequest($url, Request::METHOD_POST, $body));
        $this->assertArrayHasKey('access_token', $json);

        $parser    = new Token\Parser(new JoseEncoder());
        $parsedJWT = $parser->parse($json['access_token']);
        $userData  = $parsedJWT->claims()->get('user');

        $this->assertArrayHasKey('id', $userData);
        $this->assertArrayHasKey('roles', $userData);
        $this->assertArrayHasKey('username', $userData);
        $user = $this->em->find(User::class, $userData['id']);
        $this->assertInstanceOf(User::class, $user);
        $this->assertEquals(10, $user->getBudget());
        $this->assertNotNull($user->getSignalSubscriptionExpiresAt());
        $usedPromoCode = $user->getPromoCode();
        $this->assertInstanceOf(PromoCode::class, $usedPromoCode);
        $this->assertEquals(1, $usedPromoCode->getActivationCount());
    }

    /**
     * @param string $email
     *
     * @throws \Exception
     *
     * @return string
     */
    private function getSecret(string $email): string
    {
        $time   = (new \DateTime('now', new \DateTimeZone('UTC')))->getTimestamp();
        $string = $email . Register::AUTH_SECRET_SALT . $time;

        return hash('SHA256', $string);
    }
}
