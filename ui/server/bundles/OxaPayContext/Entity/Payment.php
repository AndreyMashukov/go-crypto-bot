<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\OxaPayContext\Entity;

use Bundles\OxaPayContext\Repository\PaymentRepository;
use Bundles\UserContext\Entity\User;
use Doctrine\ORM\Mapping as ORM;
use JMS\Serializer\Annotation as Serializer;
use Ramsey\Uuid\Uuid;

/**
 * @ORM\Entity(repositoryClass=PaymentRepository::class)
 *
 * @Serializer\ExclusionPolicy(Serializer\ExclusionPolicy::ALL)
 */
class Payment
{
    public const STATUS_PENDING = 'pending';

    public const STATUS_PAID = 'paid';

    public const STATUS_CANCELLED = 'cancelled';

    public const STATUS_EXPIRED = 'expired';

    public const CURRENCY_USDT = 'USDT';

    /**
     * @ORM\Id
     * @ORM\GeneratedValue
     * @ORM\Column(name="pmt_id", type="integer")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"payment_short", "payment"})
     */
    private ?int $id = null;

    /**
     * @ORM\Column(name="pmt_amount", type="float")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"payment"})
     */
    private float $amount;

    /**
     * @ORM\Column(name="pmt_status", type="string")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"payment"})
     */
    private string $status = self::STATUS_PENDING;

    /**
     * @ORM\Column(name="pmt_currency", type="string", length=255)
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"payment"})
     */
    private string $currency = self::CURRENCY_USDT;

    /**
     * @ORM\Column(name="pmt_description", type="string", length=255)
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"payment"})
     */
    private string $description;

    /**
     * @ORM\Column(name="pmt_order_id", type="uuid", length=255)
     */
    private string $orderId;

    /**
     * @ORM\Column(name="pmt_email", type="string", length=255)
     */
    private string $email;

    /**
     * @ORM\Column(name="pmt_created_at", type="datetime_immutable")
     */
    private \DateTimeImmutable $createdAt;

    /**
     * @ORM\Column(name="pmt_track_id", type="integer", nullable=true)
     */
    private $trackId;

    /**
     * @ORM\Column(name="pmt_payment_link", type="string", length=255, nullable=true)
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"payment_short"})
     */
    private $paymentLink;

    /**
     * @ORM\Column(name="pmt_expires_at", type="datetime_immutable", nullable=true)
     */
    private $expiresAt;

    /**
     * @ORM\Column(name="pmt_completed_at", type="datetime_immutable", nullable=true)
     */
    private ?\DateTimeImmutable $completedAt = null;

    /**
     * @ORM\ManyToOne(targetEntity=User::class)
     * @ORM\JoinColumn(name="pmt_user", nullable=false)
     */
    private $user;

    public function __construct(
        User $user,
        float $amount,
        string $description
    ) {
        $this->orderId      = Uuid::uuid4();
        $this->user         = $user;
        $this->email        = $user->getEmail();
        $this->description  = $description;
        $this->amount       = $amount;
        $this->createdAt    = new \DateTimeImmutable('now');
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getAmount(): ?float
    {
        return $this->amount;
    }

    public function setAmount(float $amount): self
    {
        $this->amount = $amount;

        return $this;
    }

    public function getCurrency(): ?string
    {
        return $this->currency;
    }

    public function setCurrency(string $currency): self
    {
        $this->currency = $currency;

        return $this;
    }

    public function getDescription(): ?string
    {
        return $this->description;
    }

    public function setDescription(string $description): self
    {
        $this->description = $description;

        return $this;
    }

    public function getOrderId(): ?string
    {
        return $this->orderId;
    }

    public function setOrderId(string $orderId): self
    {
        $this->orderId = $orderId;

        return $this;
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

    public function getPaymentLink(): ?string
    {
        return $this->paymentLink;
    }

    public function setPaymentLink(string $paymentLink): self
    {
        $this->paymentLink = $paymentLink;

        return $this;
    }

    public function getStatus(): string
    {
        return $this->status;
    }

    public function setStatus(string $status): self
    {
        $this->status = $status;

        return $this;
    }

    public function getExpiresAt(): ?\DateTimeImmutable
    {
        return $this->expiresAt;
    }

    public function setExpiresAt(\DateTimeImmutable $expiresAt): self
    {
        $this->expiresAt = $expiresAt;

        return $this;
    }

    public function getTrackId(): ?int
    {
        return $this->trackId;
    }

    public function setTrackId(int $trackId): self
    {
        $this->trackId = $trackId;

        return $this;
    }

    public function getCompletedAt(): ?\DateTimeImmutable
    {
        return $this->completedAt;
    }

    public function setCompletedAt(?\DateTimeImmutable $completedAt): self
    {
        $this->completedAt = $completedAt;

        return $this;
    }

    public function getCreatedAt(): \DateTimeImmutable
    {
        return $this->createdAt;
    }

    public function getUser(): User
    {
        return $this->user;
    }

    public function isOwnedBy(User $user): bool
    {
        return $this->getUser()->getId() === $user->getId();
    }
}
