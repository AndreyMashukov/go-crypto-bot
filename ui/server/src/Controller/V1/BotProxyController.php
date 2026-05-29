<?php

declare(strict_types=1);

namespace App\Controller\V1;

use App\Service\Bot\BotAdminClient;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Attribute\Route;

/**
 * Phase J: closed-perimeter proxy. Every UI call to /api/bot/{path}
 * forwards method + query + body to the bot's admin REST endpoints
 * (server/src/controller/, listening on :8080 inside bot-net).
 *
 * No authentication — the UI lives in a closed perimeter, the bot
 * admin port is never exposed to the public internet.
 */
final class BotProxyController extends AbstractController
{
    public function __construct(private readonly BotAdminClient $botAdmin) {}

    #[Route(
        path: '/api/bot/{path}',
        name: 'api_bot_proxy',
        requirements: ['path' => '.+'],
        methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'],
    )]
    public function proxy(Request $request, string $path): Response
    {
        $query = $request->query->all();
        $body  = $request->getContent();
        $type  = (string) $request->headers->get('Content-Type', 'application/json');

        return $this->botAdmin->forward(
            $request->getMethod(),
            $path,
            $query,
            '' !== $body ? $body : null,
            $type,
        );
    }
}
