<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\TgBotContext\Service;

use Bundles\TgBotContext\MessageBuilder;
use Bundles\TgBotContext\Model\ChatConfiguration;
use TgBotApi\BotApiBase\BotApiComplete;

class AlertService
{
    private MessageBuilder $messageBuilder;

    private BotApiComplete $botApi;

    private string $env;

    public function __construct(
        MessageBuilder $messageBuilder,
        BotApiComplete $botApi,
        string $env
    ) {
        $this->messageBuilder = $messageBuilder;
        $this->botApi         = $botApi;
        $this->env            = $env;
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
