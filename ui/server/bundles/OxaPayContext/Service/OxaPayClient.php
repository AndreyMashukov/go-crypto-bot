<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\OxaPayContext\Service;

use Bundles\OxaPayContext\Entity\Payment;
use GuzzleHttp\ClientInterface;
use GuzzleHttp\Exception\RequestException;
use Symfony\Component\HttpFoundation\Request;

class OxaPayClient
{
    public const OXA_PAY_STATUS_PAID = 'Paid';

    public const OXA_PAY_STATUS_EXPIRED = 'Expired';

    public const MAX_PAYMENT_LIFETIME_MINUTES = 2880;

    private ClientInterface $client;

    private string $baseUrl = 'https://api.oxapay.com';

    private string $apiKey;

    private string $backEndHost;

    private string $frontEndHost;

    public function __construct(
        string $backEndHost,
        string $frontEndHost,
        string $apiKey,
        ClientInterface $client
    ) {
        $this->client       = $client;
        $this->apiKey       = $apiKey;
        $this->backEndHost  = $backEndHost;
        $this->frontEndHost = $frontEndHost;
    }

    public function createInvoice(Payment $payment): void
    {
        try {
            $response = $this->client->request(Request::METHOD_POST, "{$this->baseUrl}/merchants/request", [
                'json' => [
                    'merchant'       => $this->apiKey,
                    'amount'         => $payment->getAmount(),
                    'currency'       => $payment->getCurrency(),
                    'lifeTime'       => self::MAX_PAYMENT_LIFETIME_MINUTES,
                    'feePaidByPayer' => false,
                    'underPaidCover' => 10,
                    'description'    => $payment->getDescription(),
                    'orderId'        => $payment->getOrderId(),
                    'email'          => $payment->getEmail(),
                    'callbackUrl'    => "{$this->backEndHost}/public/callback/oxa",
                    'returnUrl'      => "{$this->frontEndHost}/account?payment={$payment->getId()}",
                ],
            ]);
        } catch (RequestException $exception) {
            throw new \BadMethodCallException($exception->getMessage(), $exception->getCode(), $exception);
        } catch (\Throwable $exception) {
            throw new \RuntimeException($exception->getMessage(), $exception->getCode(), $exception);
        }

        $json = json_decode($response->getBody()->getContents(), true);

        if (!$json) {
            throw new \RuntimeException('Empty response from oxa pay!');
        }

        $payment->setTrackId($json['trackId']);
        $payment->setPaymentLink($json['payLink']);
        $payment->setExpiresAt(\DateTimeImmutable::createFromFormat('U', $json['expiredAt']));
    }
}
