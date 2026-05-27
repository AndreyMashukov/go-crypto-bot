<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Entity;

use Bundles\CryptoBotContext\Entity\Embedded\SignalConfig;
use Bundles\CryptoBotContext\Repository\CryptoTradeConfigRepository;
use Doctrine\ORM\Mapping as ORM;
use JMS\Serializer\Annotation as Serializer;
use Symfony\Component\Validator\Constraints as Assert;

/**
 * @ORM\Entity(repositoryClass=CryptoTradeConfigRepository::class)
 *
 * @Serializer\ExclusionPolicy(Serializer\ExclusionPolicy::ALL)
 */
class CryptoTradeConfig
{
    /**
     * @ORM\Id
     * @ORM\GeneratedValue
     * @ORM\Column(name="cfg_id", type="integer")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private $id;

    /**
     * @ORM\ManyToOne(targetEntity=CryptoBot::class, inversedBy="cryptoTradeConfigs")
     * @ORM\JoinColumn(name="cfg_cryptobot", nullable=false, referencedColumnName="ctb_id")
     *
     * @Assert\NotNull
     */
    private $cryptobot;

    /**
     * @ORM\Column(name="cfg_symbol", type="string", length=255)
     *
     * @Assert\NotBlank(groups={"Default", "cryptotrade_config"})
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private $symbol;

    /**
     * @ORM\Column(name="cfg_usdt_limit", type="float")
     *
     * @Assert\NotBlank(groups={"Default", "cryptotrade_config"})
     * @Assert\GreaterThanOrEqual(value="15", groups={"Default", "cryptotrade_config"})
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private $usdtLimit;

    /**
     * @ORM\Column(name="cfg_is_enabled", type="boolean")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private $enabled;

    /**
     * @ORM\Column(name="cfg_min_price_minutes_period", type="integer", options={"default": 200})
     *
     * @Assert\NotNull(groups={"Default", "cryptotrade_config"})
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private int $minPriceMinutesPeriod = 200;

    /**
     * @ORM\Column(name="cfg_frame_interval", type="string", length=5, options={"default": "2h"})
     *
     * @Assert\Choice(choices={"1m", "15m", "30m", "1h", "2h", "4h", "6h", "12h", "1d", "1w", "1M"})
     * @Assert\NotNull(groups={"Default", "cryptotrade_config"})
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private string $frameInterval = '2h';

    /**
     * @ORM\Column(name="cfg_frame_period", type="integer", options={"default": 20})
     *
     * @Assert\NotNull(groups={"Default", "cryptotrade_config"})
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private int $framePeriod = 20;

    /**
     * @ORM\Column(name="cfg_buy_price_history_check_interval", type="string", length=5, options={"default": "1d"})
     *
     * @Assert\NotNull(groups={"Default", "cryptotrade_config"})
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private string $buyPriceHistoryCheckInterval = '1d';

    /**
     * @ORM\Column(name="cfg_buy_price_history_check_period", type="integer", options={"default": 14})
     *
     * @Assert\NotNull(groups={"Default", "cryptotrade_config"})
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private int $buyPriceHistoryCheckPeriod = 14;

    /**
     * @ORM\Column(name="cfg_extra_charge_options", type="json")
     *
     * @Assert\NotNull
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private array $extraChargeOptions = [];

    /**
     * @ORM\Column(name="cfg_profit_options", type="json")
     *
     * @Assert\NotNull
     * @Assert\Length(min="1")
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private array $profitOptions = [];

    /**
     * @ORM\Column(name="cfg_buy_conditions", type="json")
     *
     * @Assert\NotNull
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private array $buyConditions = [];

    /**
     * @ORM\Column(name="cfg_sell_conditions", type="json")
     *
     * @Assert\NotNull
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private array $sellConditions = [];

    /**
     * @ORM\Column(name="cfg_avg_conditions", type="json")
     *
     * @Assert\NotNull
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private array $avgConditions = [];

    /**
     * @ORM\Column(name="cfg_signal_trading", type="boolean", options={"default": 0})
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private bool $signalTrading = false;

    /**
     * @ORM\Embedded(class="Bundles\CryptoBotContext\Entity\Embedded\SignalConfig", columnPrefix="signal_config_")
     *
     * @Assert\Valid
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private SignalConfig $signalConfig;

    /**
     * @ORM\Column(name="cfg_position_updated_at", type="datetime_immutable", nullable=true)
     */
    private ?\DateTimeImmutable $positionUpdatedAt = null;

    /**
     * @ORM\Column(name="cfg_quantity", type="float", nullable=true)
     */
    private ?float $quantity = null;

    /**
     * @ORM\Column(name="cfg_price_today_first", type="float", nullable=true)
     */
    private ?float $priceTodayFirst = null;

    /**
     * @ORM\Column(name="cfg_price_today_last", type="float", nullable=true)
     */
    private ?float $priceTodayLast = null;

    /**
     * @ORM\Column(name="cfg_average_price", type="float", nullable=true)
     */
    private ?float $averagePrice = null;

    /**
     * @ORM\Column(name="cfg_label", type="string", length=20, nullable=true)
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private ?string $label = null;

    /**
     * @ORM\Column(name="cfg_score", type="float", nullable=true)
     *
     * @Serializer\Expose
     * @Serializer\Groups(groups={"cryptotrade_config"})
     */
    private ?float $score = null;

    public function __construct()
    {
        $this->signalConfig = new SignalConfig();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getCryptobot(): ?CryptoBot
    {
        return $this->cryptobot;
    }

    public function setCryptobot(?CryptoBot $cryptobot): self
    {
        $this->cryptobot = $cryptobot;

        return $this;
    }

    public function getSymbol(): ?string
    {
        return $this->symbol;
    }

    public function setSymbol(string $symbol): self
    {
        $this->symbol = $symbol;

        return $this;
    }

    public function getUsdtLimit(): ?float
    {
        return $this->usdtLimit;
    }

    public function setUsdtLimit(float $usdtLimit): self
    {
        $this->usdtLimit = $usdtLimit;

        return $this;
    }

    public function isEnabled(): ?bool
    {
        return $this->enabled;
    }

    public function setEnabled(bool $enabled): self
    {
        $this->enabled = $enabled;

        return $this;
    }

    public function getMinPriceMinutesPeriod(): int
    {
        return $this->minPriceMinutesPeriod;
    }

    public function setMinPriceMinutesPeriod(int $minPriceMinutesPeriod): self
    {
        $this->minPriceMinutesPeriod = $minPriceMinutesPeriod;

        return $this;
    }

    public function getFrameInterval(): string
    {
        return $this->frameInterval;
    }

    public function setFrameInterval(string $frameInterval): self
    {
        $this->frameInterval = $frameInterval;

        return $this;
    }

    public function getFramePeriod(): int
    {
        return $this->framePeriod;
    }

    public function setFramePeriod(int $framePeriod): self
    {
        $this->framePeriod = $framePeriod;

        return $this;
    }

    public function getBuyPriceHistoryCheckInterval(): string
    {
        return $this->buyPriceHistoryCheckInterval;
    }

    public function setBuyPriceHistoryCheckInterval(string $buyPriceHistoryCheckInterval): self
    {
        $this->buyPriceHistoryCheckInterval = $buyPriceHistoryCheckInterval;

        return $this;
    }

    public function getBuyPriceHistoryCheckPeriod(): int
    {
        return $this->buyPriceHistoryCheckPeriod;
    }

    public function setBuyPriceHistoryCheckPeriod(int $buyPriceHistoryCheckPeriod): self
    {
        $this->buyPriceHistoryCheckPeriod = $buyPriceHistoryCheckPeriod;

        return $this;
    }

    public function getExtraChargeOptions(): array
    {
        return $this->extraChargeOptions;
    }

    public function setExtraChargeOptions(array $extraChargeOptions): self
    {
        $this->extraChargeOptions = $extraChargeOptions;

        return $this;
    }

    public function getProfitOptions(): array
    {
        return $this->profitOptions;
    }

    public function getProfitOptionsMapped(): array
    {
        return array_map(function (array $option) {
            return [
                'index'           => $option['index'],
                'optionValue'     => $option['optionValue'],
                'optionUnit'      => $option['optionUnit'],
                'optionPercent'   => $option['optionPercent'],
                'isTriggerOption' => (bool) $option['isTriggerOption'],
            ];
        }, $this->profitOptions);
    }

    public function setProfitOptions(array $profitOptions): self
    {
        $this->profitOptions = $profitOptions;

        return $this;
    }

    public function setBuyConditions(array $buyConditions): self
    {
        $this->buyConditions = $buyConditions;

        return $this;
    }

    public function getBuyConditions(): array
    {
        return $this->buyConditions;
    }

    public function setSellConditions(array $sellConditions): self
    {
        $this->sellConditions = $sellConditions;

        return $this;
    }

    public function getSellConditions(): array
    {
        return $this->sellConditions;
    }

    public function setAvgConditions(array $avgConditions): self
    {
        $this->avgConditions = $avgConditions;

        return $this;
    }

    public function getAvgConditions(): array
    {
        return $this->avgConditions;
    }

    public function isSignalTrading(): ?bool
    {
        return $this->signalTrading;
    }

    public function setSignalTrading(bool $signalTrading): self
    {
        $this->signalTrading = $signalTrading;

        return $this;
    }

    public function getSignalConfig(): SignalConfig
    {
        return $this->signalConfig;
    }

    public function getPositionUpdatedAt(): ?\DateTimeImmutable
    {
        return $this->positionUpdatedAt;
    }

    public function setPositionUpdatedAt(?\DateTimeImmutable $positionUpdatedAt): self
    {
        $this->positionUpdatedAt = $positionUpdatedAt;

        return $this;
    }

    public function getQuantity(): ?float
    {
        return $this->quantity;
    }

    public function setQuantity(?float $quantity): self
    {
        $this->quantity = $quantity;

        return $this;
    }

    public function getPriceTodayFirst(): ?float
    {
        return $this->priceTodayFirst;
    }

    public function setPriceTodayFirst(?float $priceTodayFirst): self
    {
        $this->priceTodayFirst = $priceTodayFirst;

        return $this;
    }

    public function getPriceTodayLast(): ?float
    {
        return $this->priceTodayLast;
    }

    public function setPriceTodayLast(?float $priceTodayLast): self
    {
        $this->priceTodayLast = $priceTodayLast;

        return $this;
    }

    public function getAveragePrice(): ?float
    {
        return $this->averagePrice;
    }

    public function setAveragePrice(?float $averagePrice): self
    {
        $this->averagePrice = $averagePrice;

        return $this;
    }

    public function getLabel(): ?string
    {
        return $this->label;
    }

    public function setLabel(?string $label): self
    {
        $this->label = $label;

        return $this;
    }

    public function getScore(): ?float
    {
        return $this->score;
    }

    public function setScore(?float $score): self
    {
        $this->score = $score;

        return $this;
    }
}
