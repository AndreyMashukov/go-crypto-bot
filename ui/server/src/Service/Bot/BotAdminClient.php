<?php

declare(strict_types=1);

namespace App\Service\Bot;

use Psr\Log\LoggerInterface;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Contracts\HttpClient\Exception\ExceptionInterface;
use Symfony\Contracts\HttpClient\HttpClientInterface;

/**
 * Thin proxy to the bot's admin REST server (market-trader :8080).
 *
 * Phase J: the closed-perimeter UI talks to bot's controllers
 * (server/src/controller/) exclusively through this client so we keep one
 * place to add retry / circuit-breaker / response normalisation later.
 */
final readonly class BotAdminClient
{
    public function __construct(
        private HttpClientInterface $httpClient,
        private LoggerInterface $logger,
        private string $botAdminUrl,
    ) {}

    /**
     * Forward an arbitrary request to the bot. Returns a Symfony Response
     * carrying the bot's raw payload + status code; the caller pipes it
     * back to the browser via the BotProxyController.
     *
     * @param array<string, scalar|null> $query
     */
    public function forward(string $method, string $path, array $query = [], ?string $body = null, string $contentType = 'application/json'): Response
    {
        $url = $this->botAdminUrl.'/'.ltrim($path, '/');

        try {
            $options = [
                'query'        => $query,
                'timeout'      => 10,
                'max_duration' => 12,
            ];
            if (null !== $body && '' !== $body) {
                $options['body']    = $body;
                $options['headers'] = ['Content-Type' => $contentType];
            }
            $response = $this->httpClient->request($method, $url, $options);

            $status = $response->getStatusCode();
            $raw    = $response->getContent(false);
        } catch (ExceptionInterface $exception) {
            $this->logger->error('Bot admin proxy failed', [
                'method'    => $method,
                'path'      => $path,
                'exception' => $exception->getMessage(),
            ]);

            return new JsonResponse(
                ['error' => 'bot admin unreachable: '.$exception->getMessage()],
                Response::HTTP_BAD_GATEWAY,
            );
        }

        // Normalise to JSON when the bot returns JSON; otherwise pass the raw
        // body through. The bot's controllers all emit application/json today,
        // so this branch is informational — keeps the door open for future
        // text/plain or binary responses.
        return new Response($raw, $status, ['Content-Type' => 'application/json']);
    }
}
