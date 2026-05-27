<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Model;

use Bundles\CryptoBotContext\Entity\CryptoBot;
use Symfony\Component\Validator\Constraints as Assert;

class ErrorCallback
{
    /**
     * @Assert\NotNull
     *
     * @var null|CryptoBot
     */
    public ?CryptoBot $bot = null;

    /**
     * @Assert\NotNull
     *
     * @var null|string
     */
    public ?string $errorCode = null;

    /**
     * @Assert\NotNull
     *
     * @var null|string
     */
    public ?string $errorMessage = null;

    /**
     * @Assert\NotNull
     *
     * @var null|bool
     */
    public ?bool $stop = null;
}
