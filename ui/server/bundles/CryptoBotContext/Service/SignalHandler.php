<?php
namespace Bundles\CryptoBotContext\Service;

use Bundles\CryptoBotContext\Model\Signal;
use Bundles\CryptoBotContext\Repository\CryptoTradeConfigRepository;

class SignalHandler
{
    public function __construct(private readonly CryptoTradeConfigRepository $tradeConfigRepository, private readonly CryptoBotService $cryptoBotService, private readonly SignalModification $signalModification, private readonly SignalFilter $signalFilter)
    {
    }

    public function handleSignal(Signal $signal): void
    {
        $subscribedConfigs = $this->tradeConfigRepository->getSignalSubscribers($signal);

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
