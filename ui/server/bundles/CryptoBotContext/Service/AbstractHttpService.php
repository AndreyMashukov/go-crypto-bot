<?php
namespace Bundles\CryptoBotContext\Service;

use GuzzleHttp\Exception\GuzzleException;
use GuzzleHttp\ClientInterface;

class AbstractHttpService
{
    public function __construct(private readonly ClientInterface $client)
    {
    }

    /**
     * @throws GuzzleException
     */
    protected function request(string $method, string $uri, array $json, int $timeout = 60): array
    {
        $params = [
            'timeout' => $timeout,
        ];

        if ($json !== []) {
            $params['json'] = $json;
        }

        $response = $this->client->request($method, $uri, $params);

        if ($response->getStatusCode() >= 400) {
            throw new \BadMethodCallException('Something went wrong.');
        }

        $json = json_decode($response->getBody()->getContents(), true);

        return $json ?: [];
    }

    protected function requestPlain(string $method, string $uri, array $json, int $timeout = 60): string
    {
        $params = [
            'timeout' => $timeout,
        ];

        if ($json !== []) {
            $params['json'] = $json;
        }

        $response = $this->client->request($method, $uri, $params);

        if ($response->getStatusCode() >= 400) {
            throw new \BadMethodCallException('Something went wrong.');
        }

        return $response->getBody()->getContents();
    }
}
