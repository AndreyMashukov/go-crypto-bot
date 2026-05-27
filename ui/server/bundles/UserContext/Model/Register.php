<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\UserContext\Model;

use App\Constraint as CoreAssert;
use Bundles\OxaPayContext\Entity\PromoCode;
use Symfony\Component\Validator\Constraints as Assert;

class Register implements SecureDataInterface
{
    public const AUTH_SECRET_SALT = 'sfgfdg4retw34wert';

    /**
     * @Assert\NotBlank
     *
     * @var null|string
     */
    private $nickname;

    /**
     * @Assert\NotBlank
     * @CoreAssert\Email(mode="strict")
     *
     * @var null|string
     */
    private $email;

    /**
     * @Assert\NotBlank
     *
     * @var null|string
     */
    private $secret;

    /**
     * @var null|PromoCode
     */
    private ?PromoCode $promoCode = null;

    private string $locale;

    public function __construct(string $locale)
    {
        $this->locale = $locale;
    }

    /**
     * @return null|string
     */
    public function getEmail(): ?string
    {
        return $this->email;
    }

    /**
     * @param string $email
     *
     * @return Register
     */
    public function setEmail(string $email): self
    {
        $this->email = $email;

        return $this;
    }

    /**
     * @param null|string $secret
     *
     * @return Register
     */
    public function setSecret(?string $secret): self
    {
        $this->secret = $secret;

        return $this;
    }

    public function getSecret(): ?string
    {
        return $this->secret;
    }

    public function getSecureString(): string
    {
        return $this->getEmail() . self::AUTH_SECRET_SALT;
    }

    public function getNickname(): ?string
    {
        return $this->nickname;
    }

    public function setNickname(?string $nickname): self
    {
        $this->nickname = $nickname;

        return $this;
    }

    public function getPromoCode(): ?PromoCode
    {
        return $this->promoCode;
    }

    public function setPromoCode(?PromoCode $promoCode): self
    {
        $this->promoCode = $promoCode;

        return $this;
    }

    public function getLocale(): string
    {
        return $this->locale;
    }
}
