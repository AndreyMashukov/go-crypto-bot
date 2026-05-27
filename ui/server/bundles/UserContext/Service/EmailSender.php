<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Service;

use Bundles\UserContext\Entity\User;
use GuzzleHttp\ClientInterface;
use GuzzleHttp\Exception\ClientException;

class EmailSender
{
    private ClientInterface $client;

    private string $apiKey;

    public function __construct(string $apiKey, ClientInterface $client)
    {
        $this->apiKey = $apiKey;
        $this->client = $client;
    }

    public function send(int $templateId, array $parameters, User $user)
    {
        try {
            $contact = $this->getContact($user->getEmail());
            $this->updateContact($user);
        } catch (ClientException $exception) {
            $this->createContact($user);
            $contact = $this->getContact($user->getEmail());
        }
        $response = $this->client->request('POST', 'https://api.brevo.com/v3/smtp/email', [
            'json' => [
                'sender' => [
                    'email' => 'sales@example.com',
                    'name'  => 'Andrei Mashukov',
                ],
                'templateId'      => $templateId,
                'messageVersions' => [
                    [
                        'to' => [
                            ['email' => $contact['email']],
                        ],
                        'params' => $parameters,
                    ],
                ],
            ],
            'headers' => [
                'api-key' => $this->apiKey,
            ],
        ]);
        if (201 === $response->getStatusCode()) {
            return;
        }
        throw new \LogicException('Message was not sent.');
    }

    public function getContact(string $email): array
    {
        $response = $this->client->request('GET', "https://api.brevo.com/v3/contacts/{$email}", [
            'headers' => [
                'api-key' => $this->apiKey,
            ],
        ]);

        return json_decode($response->getBody()->getContents(), true);
    }

    public function updateContact(User $user): void
    {
        $response = $this->client->request('PUT', "https://api.brevo.com/v3/contacts/{$user->getEmail()}", [
            'json' => [
                'listIds'    => [13],
                'attributes' => [
                    'FNAME' => $user->getFirstName(),
                    'LNAME' => $user->getLastName(),
                ],
            ],
            'headers' => [
                'api-key' => $this->apiKey,
            ],
        ]);
        if (204 === $response->getStatusCode()) {
            return;
        }
        throw new \LogicException('Contact was not updated.');
    }

    public function createContact(User $user): void
    {
        $response = $this->client->request('POST', 'https://api.brevo.com/v3/contacts', [
            'json' => [
                'email'      => $user->getEmail(),
                'listIds'    => [13],
                'attributes' => [
                    'FNAME' => $user->getFirstName(),
                    'LNAME' => $user->getLastName(),
                ],
            ],
            'headers' => [
                'api-key' => $this->apiKey,
            ],
        ]);

        if (201 === $response->getStatusCode()) {
            return;
        }
        throw new \LogicException('Contact was not created.');
    }
}
