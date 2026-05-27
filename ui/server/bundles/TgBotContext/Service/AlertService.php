<?php

declare(strict_types=1);

namespace Bundles\TgBotContext\Service;

use Bundles\TgBotContext\MessageBuilder;
use Bundles\TgBotContext\Model\ChatConfiguration;
use TgBotApi\BotApiBase\BotApiComplete;

class AlertService
{
    private readonly BotApiComplete $botApi;

    public function __construct(
        private readonly MessageBuilder $messageBuilder,
        BotApiComplete $botApi,
        private readonly string $env
    ) {
        $this->botApi         = $botApi;
    }

    public function alert(string $message): void
    {
        if ('prod' !== $this->env) {
            return;
        }

        try {
            foreach ($this->messageBuilder->getAlertMessages($message) as $alertMessage) {
                $alertMessage->send($this->botApi, new ChatConfiguration(-4284072972));
            }
        } catch (\Throwable $throwable) {
            unset($throwable);
        }
    }
}
