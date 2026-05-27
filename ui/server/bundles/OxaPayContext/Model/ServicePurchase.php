<?php

declare(strict_types=1);

namespace Bundles\OxaPayContext\Model;

use Symfony\Component\Validator\Constraints as Assert;

class ServicePurchase
{
    #[Assert\NotBlank]
    #[Assert\Choice(choices: ['signal_subscription', 'basic_subscription', 'dedicated_binance_server', 'dedicated_bybit_server', 'api_subscription'])]
    public ?string $code = null;
}
