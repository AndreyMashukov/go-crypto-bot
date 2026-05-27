<?php
namespace Bundles\CryptoBotContext\Entity;

use Bundles\CryptoBotContext\Entity\Embedded\SignalConfig;

class CryptoTradeConfig
{
    private ?int $id = null;

    private ?CryptoBot $cryptobot = null;

    private ?string $symbol = null;

    private ?float $usdtLimit = null;

    private ?bool $enabled = null;

    private int $minPriceMinutesPeriod = 200;

    private string $frameInterval = '2h';

    private int $framePeriod = 20;

    private string $buyPriceHistoryCheckInterval = '1d';

    private int $buyPriceHistoryCheckPeriod = 14;

    private array $extraChargeOptions = [];

    private array $profitOptions = [];

    private array $buyConditions = [];

    private array $sellConditions = [];

    private array $avgConditions = [];

    private bool $signalTrading = false;

    private readonly SignalConfig $signalConfig;

    private ?\DateTimeImmutable $positionUpdatedAt = null;

    private ?float $quantity = null;

    private ?float $priceTodayFirst = null;

    private ?float $priceTodayLast = null;

    private ?float $averagePrice = null;

    private ?string $label = null;

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
        return array_map(fn(array $option) => [
            'index'           => $option['index'],
            'optionValue'     => $option['optionValue'],
            'optionUnit'      => $option['optionUnit'],
            'optionPercent'   => $option['optionPercent'],
            'isTriggerOption' => (bool) $option['isTriggerOption'],
        ], $this->profitOptions);
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
