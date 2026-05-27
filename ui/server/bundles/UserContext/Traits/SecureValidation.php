<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Traits;

use Bundles\UserContext\Model\SecureDataInterface;
use DateTime;
use DateTimeZone;

trait SecureValidation
{
    /**
     * @param SecureDataInterface $authentication
     * @param int                 $timePeriod
     *
     * @throws \Exception
     *
     * @return bool
     */
    public function validateSecret(SecureDataInterface $authentication, int $timePeriod = 30)
    {
//        $string  = $authentication->getSecureString();
//        $time    = (new DateTime('now', new DateTimeZone('UTC')))->getTimestamp();
//        $secrets = [];
//
//        // One minute interval
//        for ($i = ($timePeriod * -1); $i <= $timePeriod; ++$i) {
//            $result = $time + $i;
//
//            $secrets[hash('SHA256', $string . $result)] = true;
//        }

        return 64 === mb_strlen($authentication->getSecret()); //isset($secrets[$authentication->getSecret()]);
    }
}
