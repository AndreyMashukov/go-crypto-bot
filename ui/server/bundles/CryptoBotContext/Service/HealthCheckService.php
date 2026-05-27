<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoBot;

class HealthCheckService extends AbstractHttpService
{
    public function healthCheck(CryptoBot $cryptobot): array
    {
        if (!$cryptobot->getIpAddress()) {
            throw new \BadMethodCallException('Bot must have IP address');
        }

        return $this->request(
            'GET',
            "http://{$cryptobot->getIpAddress()}:{$cryptobot->getPort()}/health/check?botUuid={$cryptobot->getUuid()}",
            []
        );
    }
}
