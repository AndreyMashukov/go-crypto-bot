<?php
// RECTOR-BAN: superglobal access ($_ENV/$_SERVER/$_GET/$_POST/$_REQUEST/$_COOKIE/$_FILES/$_SESSION) and getenv/putenv are forbidden — read env via DI constructor args wired from container configuration; read request via your framework Request object
namespace App\Controller;

use Bundles\UserContext\Entity\User;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use League\OAuth2\Server\Grant\RefreshTokenGrant;
use Swagger\Annotations as SWG;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

// RECTOR-BAN: superglobal access ($_ENV/$_SERVER/$_GET/$_POST/$_REQUEST/$_COOKIE/$_FILES/$_SESSION) and getenv/putenv are forbidden — read env via DI constructor args wired from container configuration; read request via your framework Request object
class OAuthController extends AbstractFOSRestController
{
    public function __construct(private readonly RefreshTokenGrant $refreshTokenGrant)
    {
    }

    /**
     * @return Response
     */
    public function postToken()
    {
        return $this->forward('League\Bundle\OAuth2ServerBundle\Controller\TokenController::indexAction');
    }

    /**
     * @return Response
     */
    public function postRefresh(Request $request)
    {
        $refreshToken = $request->get('refresh_token');

        if ($refreshToken) {
            try {
                $user = $this->getUserByRefreshToken($refreshToken);

                $_SERVER['user_claim'] = [
                    'id'       => $user->getId(),
                    'username' => $user->getUsername(),
                    'roles'    => $user->getRoles(),
                ];
            } catch (\Throwable $exception) {
                unset($exception);
            }
        }

        return $this->forward('League\Bundle\OAuth2ServerBundle\Controller\TokenController::indexAction');
    }

    private function getUserByRefreshToken(string $refreshToken): User
    {
        $reflection = new \ReflectionObject($this->refreshTokenGrant);
        $method     = $reflection->getMethod('decrypt');
        $result = json_decode((string) $method->invoke($this->refreshTokenGrant, $refreshToken), true);

        if (!$result || !isset($result['user_id'])) {
            throw new \Exception('Not found');
        }

        return $this->getDoctrine()->getRepository(User::class)->findOneBy([
            'username' => $result['user_id'],
        ]);
    }
}
