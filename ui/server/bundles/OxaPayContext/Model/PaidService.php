<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\OxaPayContext\Model;

use JMS\Serializer\Annotation as Serializer;

/**
 * @Serializer\ExclusionPolicy(Serializer\ExclusionPolicy::ALL)
 */
class PaidService
{
    /**
     * @var string
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"paid_service"})
     */
    private string $code;

    /**
     * @var bool
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"paid_service"})
     */
    private bool $isActive;

    /**
     * @var float
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"paid_service"})
     */
    private float $price;

    /**
     * @var int
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"paid_service"})
     */
    private int $days;

    /**
     * @var null|\DateTimeImmutable
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"paid_service"})
     */
    private ?\DateTimeImmutable $expiresAt;

    public function __construct(
        string $code,
        bool $isActive,
        float $price,
        int $days,
        ?\DateTimeImmutable $expiresAt
    ) {
        $this->code      = $code;
        $this->isActive  = $isActive;
        $this->price     = $price;
        $this->days      = $days;
        $this->expiresAt = $expiresAt;
    }

    public function getCode(): string
    {
        return $this->code;
    }

    public function isActive(): bool
    {
        return $this->isActive;
    }

    public function getPrice(): float
    {
        return $this->price;
    }

    public function getDays(): int
    {
        return $this->days;
    }

    public function getExpiresAt(): ?\DateTimeImmutable
    {
        return $this->expiresAt;
    }
}
