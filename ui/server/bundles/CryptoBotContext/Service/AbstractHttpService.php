<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Service;

use GuzzleHttp\ClientInterface;

class AbstractHttpService
{
    private ClientInterface $client;

    public function __construct(ClientInterface $client)
    {
        $this->client = $client;
    }

    /**
     * @param string $method
     * @param string $uri
     * @param array  $json
     *
     * @throws \GuzzleHttp\Exception\GuzzleException
     *
     * @return array
     */
    protected function request(string $method, string $uri, array $json, int $timeout = 60): array
    {
        $params = [
            'timeout' => $timeout,
        ];

        if ($json) {
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

        if ($json) {
            $params['json'] = $json;
        }

        $response = $this->client->request($method, $uri, $params);

        if ($response->getStatusCode() >= 400) {
            throw new \BadMethodCallException('Something went wrong.');
        }

        return $response->getBody()->getContents();
    }
}
