<?php

declare(strict_types=1);

namespace Bundles\UserContext\Traits;

use Bundles\UserContext\Model\SecureDataInterface;

trait SecureValidation
{
    /**
     *
     * @throws \Exception
     *
     * @return bool
     */
    public function validateSecret(SecureDataInterface $authentication, int $timePeriod = 30)
    {
return 64 === mb_strlen((string) $authentication->getSecret()); 
    }
}
