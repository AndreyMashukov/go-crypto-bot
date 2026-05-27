<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace App\Controller;

use Bundles\UserContext\Entity\User;
use FOS\RestBundle\Controller\AbstractFOSRestController;
use FOS\RestBundle\Controller\Annotations as Rest;
use League\OAuth2\Server\Grant\RefreshTokenGrant;
use Swagger\Annotations as SWG;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

/**
 * @Rest\Route("/oauth2", name="oauth2_")
 */
class OAuthController extends AbstractFOSRestController
{
    private RefreshTokenGrant $refreshTokenGrant;

    public function __construct(RefreshTokenGrant $refreshTokenGrant)
    {
        $this->refreshTokenGrant = $refreshTokenGrant;
    }

    /**
     * @Rest\Route("/token", name="token", methods={"POST"}, options={"expose": true})
     * @SWG\Post(
     *     description="Access Token request method `grant_type = password`",
     *     @SWG\Parameter(
     *         name="Body",
     *         in="body",
     *         required=true,
     *         @SWG\Schema(
     *             type="object",
     *             ref="#/definitions/OAuth2TokenBody",
     *         )
     *     ),
     *     @SWG\Response(
     *         response=200,
     *         description="Authorized successfully",
     *         @SWG\Schema(
     *             type="object",
     *             ref="#/definitions/OAuth2Response"
     *         )
     *     ),
     *     @SWG\Response(
     *         response=404,
     *         description="User is not found",
     *         @SWG\Schema(
     *             type="object",
     *             ref="#/definitions/NotFoundResponse"
     *         )
     *     ),
     *     @SWG\Response(
     *         response=400,
     *         description="Invalid Scope",
     *         @SWG\Schema(
     *             type="object",
     *             ref="#/definitions/InvalidScopeResponse"
     *         )
     *     ),
     *     @SWG\Response(
     *         response=401,
     *         description="Invalid Client",
     *         @SWG\Schema(
     *             type="object",
     *             ref="#/definitions/InvalidClientResponse"
     *         )
     *     )
     * )
     * @SWG\Tag(name="OAuth2")
     *
     * @return Response
     */
    public function postTokenAction()
    {
        return $this->forward('Trikoder\Bundle\OAuth2Bundle\Controller\TokenController::indexAction');
    }

    /**
     * @Rest\Route("/refresh", name="refresh", methods={"POST"}, options={"expose": true})
     * @SWG\Post(
     *     description="Refresh Token method `grant_type = refresh_token`",
     *     @SWG\Parameter(
     *         name="Body",
     *         in="body",
     *         required=true,
     *         @SWG\Schema(
     *             type="object",
     *             ref="#/definitions/OAuth2RefreshBody",
     *         )
     *     ),
     *     @SWG\Response(
     *         response=200,
     *         description="Authorized successfully",
     *         @SWG\Schema(
     *             type="object",
     *             ref="#/definitions/OAuth2Response"
     *         )
     *     ),
     *     @SWG\Response(
     *         response=404,
     *         description="User is not found",
     *         @SWG\Schema(
     *             type="object",
     *             ref="#/definitions/NotFoundResponse"
     *         )
     *     ),
     *     @SWG\Response(
     *         response=400,
     *         description="Invalid Grant Type",
     *         @SWG\Schema(
     *             type="object",
     *             ref="#/definitions/InvalidGrantResponse"
     *         )
     *     ),
     *     @SWG\Response(
     *         response=401,
     *         description="Invalid Client",
     *         @SWG\Schema(
     *             type="object",
     *             ref="#/definitions/InvalidClientResponse"
     *         )
     *     )
     * )
     * @SWG\Tag(name="OAuth2")
     *
     * @param Request $request
     *
     * @return Response
     *
     * @SuppressWarnings(PHPMD)
     */
    public function postRefreshAction(Request $request)
    {
        $refreshToken = $request->get('refresh_token', null);

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

        return $this->forward('Trikoder\Bundle\OAuth2Bundle\Controller\TokenController::indexAction');
    }

    private function getUserByRefreshToken(string $refreshToken): User
    {
        $reflection = new \ReflectionObject($this->refreshTokenGrant);
        $method     = $reflection->getMethod('decrypt');
        $method->setAccessible(true);
        $result = json_decode($method->invoke($this->refreshTokenGrant, $refreshToken), true);

        if (!$result || !isset($result['user_id'])) {
            throw new \Exception('Not found');
        }

        return $this->getDoctrine()->getRepository(User::class)->findOneBy([
            'username' => $result['user_id'],
        ]);
    }
}
