<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Tests\Functional\OAuth;

use App\Tests\RestTestCase;
use Bundles\UserContext\Entity\User;
use Lcobucci\JWT\Encoding\JoseEncoder;
use Lcobucci\JWT\Token;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * @group functional
 */
class AuthTest extends RestTestCase
{
    public function testShouldAllowToAuth(): void
    {
        $user = $this->em->getRepository(User::class)->findOneBy([
            'email' => 'john@batur.loc',
        ]);
        $this->assertInstanceOf(User::class, $user);
        self::getContainer()->get('test.brute_force_security')->invalidate($user);

        $url  = $this->getUrl('oauth2_token', []);
        $body = [
            'client_id'     => self::CLIENT_ID,
            'client_secret' => self::CLIENT_SECRET,
            'grant_type'    => 'password',
            'username'      => 'john@batur.loc',
            'password'      => '112233',
            'scope'         => ['user'],
        ];

        static::$client->request(Request::METHOD_POST, $url, $body);
        $response = static::$client->getResponse();
        $content  = $this->deserialize($response);

        $this->assertArrayHasKey('access_token', $content);

        $token = $content['access_token'];

        $payload = json_decode(base64_decode(explode('.', $token)[1], true), true);

        $this->assertArrayHasKey('aud', $payload);
        $this->assertArrayHasKey('jti', $payload);
        $this->assertArrayHasKey('nbf', $payload);
        $this->assertArrayHasKey('exp', $payload);
        $this->assertArrayHasKey('sub', $payload);
        $this->assertArrayHasKey('scopes', $payload);

        $body = [
            'client_id'     => self::CLIENT_ID,
            'client_secret' => self::CLIENT_SECRET,
            'grant_type'    => 'refresh_token',
            'refresh_token' => $content['refresh_token'],
            'scope'         => ['user'],
        ];

        $url  = $this->getUrl('oauth2_refresh', []);

        unset($_SERVER['user_claim']);

        static::$client->request(Request::METHOD_POST, $url, $body);
        $response = static::$client->getResponse();
        $content  = $this->deserialize($response);

        $this->assertArrayHasKey('access_token', $content);

        $parser    = new Token\Parser(new JoseEncoder());
        $parsedJWT = $parser->parse($content['access_token']);
        $userData  = $parsedJWT->claims()->get('user');

        $this->assertArrayHasKey('roles', $userData);
        $this->assertArrayHasKey('username', $userData);
        $this->assertEquals($this->username, $userData['username']);
    }

    /**
     * Should allow to authorize broker, borrower with same SCOPES.
     *
     * @dataProvider dataProviderAuth
     */
    public function testShouldAllowToAuthorize(array $params, int $code): void
    {
        $url  = $this->getUrl('oauth2_token', []);
        $body = array_merge([
            'client_id'     => self::CLIENT_ID,
            'client_secret' => self::CLIENT_SECRET,
            'grant_type'    => 'password',
        ], $params);

        static::$client->request(Request::METHOD_POST, $url, $body);
        $response = static::$client->getResponse();
        $json     = $this->deserialize($response, $code);
        $this->assertJsonSnapshot($json);
    }

    public function dataProviderAuth(): \Generator
    {
        yield 'test #1' => [
            'params' => [
                'username' => 'john',
                'password' => 'passwordd',
                'scope'    => ['user'],
            ],
            'code' => 400,
        ];

        yield 'test #2' => [
            'params' => [
                'username' => 'john',
                'password' => '112233',
                'scope'    => ['user'],
            ],
            'code' => 200,
        ];

        yield 'test #3' => [
            'params' => [
                'username' => 'john@batur.loc',
                'password' => '112233',
                'scope'    => ['user'],
            ],
            'code' => 200,
        ];
    }

    /**
     * Should prevent bruteforce.
     */
    public function testShouldPreventBruteForce(): void
    {
        $user = $this->em->getRepository(User::class)->findOneBy([
            'email' => 'john@batur.loc',
        ]);
        $this->assertInstanceOf(User::class, $user);
        self::getContainer()->get('test.brute_force_security')->invalidate($user);

        for ($i = 0; $i < 6; ++$i) {
            $url  = $this->getUrl('oauth2_token', []);
            $body = [
                'client_id'     => self::CLIENT_ID,
                'client_secret' => self::CLIENT_SECRET,
                'grant_type'    => 'password',
                'username'      => 'john@batur.loc',
                'password'      => '112235',
                'scope'         => ['user'],
            ];

            static::$client->request(Request::METHOD_POST, $url, $body);
            $response = static::$client->getResponse();

            if ($i < 2) {
                $this->assertEquals(Response::HTTP_BAD_REQUEST, $response->getStatusCode(), $i);
            } else {
                $this->assertEquals(Response::HTTP_TOO_MANY_REQUESTS, $response->getStatusCode(), $i);
            }
        }
    }
}
