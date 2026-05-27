<?php

declare(strict_types=1);

namespace Bundles\CryptoBotContext\Model;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Symfony\Component\Validator\Constraints as Assert;

class ErrorCallback
{
    /**
     * @Assert\NotNull
     */
    public ?CryptoBot $bot = null;

    /**
     * @Assert\NotNull
     */
    public ?string $errorCode = null;

    /**
     * @Assert\NotNull
     */
    public ?string $errorMessage = null;

    /**
     * @Assert\NotNull
     */
    public ?bool $stop = null;
}
