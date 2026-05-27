<?php
namespace Bundles\UserContext\Model;

use Bundles\OxaPayContext\Entity\PromoCode;

class Register implements SecureDataInterface
{
    public const AUTH_SECRET_SALT = 'sfgfdg4retw34wert';

    private $nickname;

    private $email;

    private $secret;

    private ?PromoCode $promoCode = null;

    public function __construct(private readonly string $locale)
    {
    }

    public function getEmail(): ?string
    {
        return $this->email;
    }

    public function setEmail(string $email): self
    {
        $this->email = $email;

        return $this;
    }

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
