<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Entity\CryptoTradeConfig;
use Bundles\CryptoBotContext\Model\Signal;
use Bundles\CryptoBotContext\Repository\CryptoTradeConfigRepository;

class SignalHandler
{
    private CryptoBotService $cryptoBotService;

    private CryptoTradeConfigRepository $tradeConfigRepository;

    private SignalModification $signalModification;

    private SignalFilter $signalFilter;

    public function __construct(
        CryptoTradeConfigRepository $tradeConfigRepository,
        CryptoBotService $cryptoBotService,
        SignalModification $signalModification,
        SignalFilter $signalFilter
    ) {
        $this->cryptoBotService      = $cryptoBotService;
        $this->tradeConfigRepository = $tradeConfigRepository;
        $this->signalModification    = $signalModification;
        $this->signalFilter          = $signalFilter;
    }

    public function handleSignal(Signal $signal): void
    {
        $subscribedConfigs = $this->tradeConfigRepository->getSignalSubscribers($signal);

        /** @var CryptoTradeConfig $config */
        foreach ($subscribedConfigs as $config) {
            try {
                $signal = $this->signalModification->modify($signal, $config);
                $this->signalFilter->process($signal, $config);
                $this->cryptoBotService->sendTradeSignal(
                    $config->getCryptobot(),
                    $signal
                );
            } catch (\Throwable $throwable) {
                unset($throwable);
            }
        }
    }
}
