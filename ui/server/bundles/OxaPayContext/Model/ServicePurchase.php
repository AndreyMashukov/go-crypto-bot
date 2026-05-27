<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\OxaPayContext\Model;

use Symfony\Component\Validator\Constraints as Assert;

class ServicePurchase
{
    /**
     * @Assert\NotBlank
     * @Assert\Choice(choices={"signal_subscription", "basic_subscription", "dedicated_binance_server", "dedicated_bybit_server", "api_subscription"})
     *
     * @var null|string
     */
    public ?string $code = null;
}
