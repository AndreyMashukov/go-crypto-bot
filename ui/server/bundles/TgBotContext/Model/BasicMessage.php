<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\TgBotContext\Model;

use Bundles\TgBotContext\Model\Traits\SendTrait;
use TgBotApi\BotApiBase\BotApiComplete;
use TgBotApi\BotApiBase\Method\SendMessageMethod;

// $keyboard = (new Keyboard())
//                ->addInlineAction('База Аренды (Снять)', null, 'rent')
//                ->addInlineAction('База Продажи (Купить)', null, 'sell')
//                ->toOneRow(1)
//            ;
//
//            return [
//                new BasicMessage('Какая база квартир вас интересует?', $keyboard),
//            ];
// $item->send($this->botApi, $configuration);

class BasicMessage implements Output, MessageInterface
{
    use SendTrait;

    private string $message;

    private ?Keyboard $keyboard;

    public function __construct(string $message, ?Keyboard $keyboard)
    {
        $this->message  = $message;
        $this->keyboard = $keyboard;
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
     * @param BotApiComplete         $bot
     * @param ConfigurationInterface $configuration
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
