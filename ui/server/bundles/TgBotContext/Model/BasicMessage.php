<?php
namespace Bundles\TgBotContext\Model;

use Bundles\TgBotContext\Model\Traits\SendTrait;
use TgBotApi\BotApiBase\BotApiComplete;
use TgBotApi\BotApiBase\Method\SendMessageMethod;

class BasicMessage implements Output, MessageInterface
{
    use SendTrait;

    public function __construct(private string $message, private ?Keyboard $keyboard)
    {
    }

    public function getKeyboard(): ?Keyboard
    {
        return $this->keyboard;
    }

    public function getMessage(): string
    {
        return $this->message;
    }

    /**
     *
     * @throws \Throwable
     */
    public function send(BotApiComplete $bot, ConfigurationInterface $configuration): void
    {
        $params   = ['parseMode' => 'html'];
        $keyboard = $this->getKeyboard();

        if ($keyboard instanceof Keyboard) {
            $params = array_merge([
                'replyMarkup' => $keyboard->getKeyboard(),
            ], $params);
        }

        $method = SendMessageMethod::create($configuration->getChatId(), $this->getMessage(), $params);
        $this->handleSend(function () use ($bot, $method) {
            $bot->sendMessage($method);
        });
    }
}
